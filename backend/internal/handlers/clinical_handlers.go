package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"upasham-backend/internal/models"
	"upasham-backend/internal/services"
)

type ClinicalHandler struct {
	db      *sql.DB
	pdfServ *services.PDFService
}

func NewClinicalHandler(db *sql.DB, pdfServ *services.PDFService) *ClinicalHandler {
	return &ClinicalHandler{
		db:      db,
		pdfServ: pdfServ,
	}
}

type SaveEncounterRequest struct {
	BookingID       string                `json:"booking_id" binding:"required"`
	Diagnosis       string                `json:"diagnosis" binding:"required"`
	SubjectiveNotes string                `json:"subjective_notes"`
	ObjectiveNotes  string                `json:"objective_notes"`
	PrivateInternal string                `json:"private_internal"` // internal clinical notes hidden from patient
	Medicines       []models.MedicineLine `json:"medicines" binding:"required"`
	Tests           []models.TestOrder    `json:"tests"`
	Advice          string                `json:"advice"`
	FollowUpDate    string                `json:"follow_up_date"`
	Warnings        string                `json:"warnings"`
}

func (h *ClinicalHandler) SaveEncounterAndPrescription(c *gin.Context) {
	doctorID := c.GetString("doctor_id")
	doctorUserID := c.GetString("user_id")
	if doctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor profile required"})
		return
	}

	var req SaveEncounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch booking info
	var patientProfileID, bookingRef string
	err := h.db.QueryRow(`
		SELECT patient_profile_id, booking_ref FROM bookings WHERE id = $1 AND doctor_id = $2
	`, req.BookingID, doctorID).Scan(&patientProfileID, &bookingRef)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the assigned doctor for this booking"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// 1. Create or Update Encounter
	encounterID := "enc-" + uuid.New().String()[:8]
	_, err = tx.Exec(`
		INSERT INTO encounters (id, booking_id, doctor_id, patient_profile_id, started_at, ended_at, status)
		VALUES ($1, $2, $3, $4, NOW() - INTERVAL '15 minutes', NOW(), 'COMPLETED')
		ON CONFLICT (booking_id) DO UPDATE SET ended_at = NOW(), status = 'COMPLETED'
	`, encounterID, req.BookingID, doctorID, patientProfileID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create encounter: " + err.Error()})
		return
	}

	// 2. Clinical Notes (Patient visible and internal private notes)
	noteID := "cnote-" + uuid.New().String()[:8]
	_, _ = tx.Exec(`
		INSERT INTO clinical_notes (id, encounter_id, subjective, objective, assessment, plan, is_private, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, false, NOW())
	`, noteID, encounterID, req.SubjectiveNotes, req.ObjectiveNotes, req.Diagnosis, req.Advice)

	if req.PrivateInternal != "" {
		privateNoteID := "cnote-priv-" + uuid.New().String()[:8]
		_, _ = tx.Exec(`
			INSERT INTO clinical_notes (id, encounter_id, subjective, objective, assessment, plan, is_private, created_at)
			VALUES ($1, $2, 'INTERNAL NOTES', '', '', $3, true, NOW())
		`, privateNoteID, encounterID, req.PrivateInternal)
	}

	// 3. Prescription
	rxID := "rx-" + uuid.New().String()[:8]
	rxRef := fmt.Sprintf("RX-UP-%d", time.Now().UnixNano()%1000000)

	// Verification hash for security and verification
	hashInput := fmt.Sprintf("%s|%s|%s|%s|%d", rxRef, doctorID, patientProfileID, req.Diagnosis, time.Now().Unix())
	hashBytes := sha256.Sum256([]byte(hashInput))
	verificationHash := hex.EncodeToString(hashBytes[:16])

	_, err = tx.Exec(`
		INSERT INTO prescriptions (id, booking_id, encounter_id, prescription_ref, doctor_id, patient_profile_id,
		                          diagnosis, advice, follow_up_date, warnings, verification_hash, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1, NOW())
	`, rxID, req.BookingID, encounterID, rxRef, doctorID, patientProfileID,
		req.Diagnosis, req.Advice, req.FollowUpDate, req.Warnings, verificationHash)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save prescription: " + err.Error()})
		return
	}

	// 4. Medicine Lines
	for i, m := range req.Medicines {
		medID := fmt.Sprintf("med-%d-%d", time.Now().UnixNano()%100000, i)
		_, err = tx.Exec(`
			INSERT INTO medicine_lines (id, prescription_id, brand_name, generic_name, strength, dosage_form, dose, route, frequency, duration, instructions)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, medID, rxID, m.BrandName, m.GenericName, m.Strength, m.DosageForm, m.Dose, m.Route, m.Frequency, m.Duration, m.Instructions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save medicine line: " + err.Error()})
			return
		}
	}

	// 5. Diagnostic Test Orders
	for i, t := range req.Tests {
		tstID := fmt.Sprintf("tst-%d-%d", time.Now().UnixNano()%100000, i)
		_, _ = tx.Exec(`
			INSERT INTO test_orders (id, prescription_id, test_name, instructions, status)
			VALUES ($1, $2, $3, $4, 'PENDING')
		`, tstID, rxID, t.TestName, t.Instructions)
	}

	// 6. Prescription Version Snapshot
	snapshotData, _ := json.Marshal(gin.H{
		"rx_ref":     rxRef,
		"diagnosis":  req.Diagnosis,
		"medicines":  req.Medicines,
		"tests":      req.Tests,
		"advice":     req.Advice,
		"hash":       verificationHash,
		"created_at": time.Now(),
	})
	_, _ = tx.Exec(`
		INSERT INTO prescription_versions (id, prescription_id, version_num, snapshot_json, created_at)
		VALUES ($1, $2, 1, $3, NOW())
	`, fmt.Sprintf("rxv-%d", time.Now().UnixNano()), rxID, string(snapshotData))

	// 7. Transition Booking to "Completed"
	_, _ = tx.Exec(`
		UPDATE bookings SET status = 'Completed', updated_at = NOW() WHERE id = $1
	`, req.BookingID)

	_, _ = tx.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, 'In Progress', 'Completed', $3, 'Consultation concluded and electronic prescription issued', NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), req.BookingID, doctorUserID)

	// Audit Log
	_, _ = tx.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'ISSUE_PRESCRIPTION', 'Prescription', $3, $4, $5, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), doctorUserID, rxID, c.ClientIP(), fmt.Sprintf(`{"prescription_ref": "%s", "booking_ref": "%s"}`, rxRef, bookingRef))

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "Consultation completed and official prescription generated",
		"prescription_id":   rxID,
		"prescription_ref":  rxRef,
		"verification_hash": verificationHash,
	})
}

// DownloadPrescriptionPDF generates and returns the branded PDF
func (h *ClinicalHandler) DownloadPrescriptionPDF(c *gin.Context) {
	rxID := c.Param("id")

	var rx models.Prescription
	err := h.db.QueryRow(`
		SELECT pr.id, pr.booking_id, pr.encounter_id, pr.prescription_ref, pr.doctor_id, pr.patient_profile_id,
		       pr.diagnosis, pr.advice, pr.follow_up_date, pr.warnings, pr.verification_hash, pr.version, pr.created_at,
		       dp.full_name, dp.bmdc_number, dp.qualifications, dp.workplace,
		       pp.full_name, pp.age, pp.gender,
		       b.booking_ref
		FROM prescriptions pr
		JOIN bookings b ON pr.booking_id = b.id
		JOIN patient_profiles pp ON pr.patient_profile_id = pp.id
		JOIN doctor_profiles dp ON pr.doctor_id = dp.id
		WHERE pr.id = $1 OR pr.prescription_ref = $1
	`, rxID).Scan(
		&rx.ID, &rx.BookingID, &rx.EncounterID, &rx.PrescriptionRef, &rx.DoctorID, &rx.PatientProfileID,
		&rx.Diagnosis, &rx.Advice, &rx.FollowUpDate, &rx.Warnings, &rx.VerificationHash, &rx.Version, &rx.CreatedAt,
		&rx.DoctorName, &rx.DoctorBMDC, &rx.DoctorDegrees, &rx.DoctorWorkplace,
		&rx.PatientName, &rx.PatientAge, &rx.PatientGender,
		&rx.BookingRef,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		return
	}

	// Object-level authorization: clinical documents are never public.
	role := c.GetString("role")
	patientAccountID := c.GetString("patient_account_id")
	doctorID := c.GetString("doctor_id")
	var ownerAccountID string
	if err := h.db.QueryRow("SELECT patient_account_id FROM bookings WHERE id = $1", rx.BookingID).Scan(&ownerAccountID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription booking not found"})
		return
	}
	allowed := role == models.RoleAdmin ||
		(role == models.RolePatient && patientAccountID != "" && patientAccountID == ownerAccountID) ||
		(role == models.RoleDoctor && doctorID != "" && doctorID == rx.DoctorID)
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this prescription"})
		return
	}
	_, _ = h.db.Exec(`INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'READ_PRESCRIPTION_PDF', 'Prescription', $3, $4, '{}', NOW())`,
		fmt.Sprintf("aud-%d", time.Now().UnixNano()), c.GetString("user_id"), rx.ID, c.ClientIP())

	// Fetch medicine lines
	medRows, _ := h.db.Query(`
		SELECT id, prescription_id, brand_name, generic_name, strength, dosage_form, dose, route, frequency, duration, instructions
		FROM medicine_lines WHERE prescription_id = $1
	`, rx.ID)
	for medRows.Next() {
		var m models.MedicineLine
		if err := medRows.Scan(&m.ID, &m.PrescriptionID, &m.BrandName, &m.GenericName, &m.Strength, &m.DosageForm, &m.Dose, &m.Route, &m.Frequency, &m.Duration, &m.Instructions); err == nil {
			rx.Medicines = append(rx.Medicines, m)
		}
	}
	medRows.Close()

	// Fetch test orders
	tstRows, _ := h.db.Query(`
		SELECT id, prescription_id, test_name, instructions, status
		FROM test_orders WHERE prescription_id = $1
	`, rx.ID)
	for tstRows.Next() {
		var t models.TestOrder
		if err := tstRows.Scan(&t.ID, &t.PrescriptionID, &t.TestName, &t.Instructions, &t.Status); err == nil {
			rx.Tests = append(rx.Tests, t)
		}
	}
	tstRows.Close()

	pdfBytes, err := h.pdfServ.GeneratePrescriptionPDF(&rx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"Upasham_Prescription_%s.pdf\"", rx.PrescriptionRef))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
