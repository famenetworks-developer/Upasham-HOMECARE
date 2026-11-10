package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"upasham-backend/internal/models"
)

type PatientHandler struct {
	db *sql.DB
}

func NewPatientHandler(db *sql.DB) *PatientHandler {
	return &PatientHandler{db: db}
}

type CreateProfileRequest struct {
	FullName          string `json:"full_name" binding:"required"`
	Relationship      string `json:"relationship" binding:"required"` // self, child, spouse, parent, other
	Gender            string `json:"gender" binding:"required"`
	Age               int    `json:"age" binding:"required"`
	BloodGroup        string `json:"blood_group"`
	KnownAllergies    string `json:"known_allergies"`
	ChronicConditions string `json:"chronic_conditions"`
	Notes             string `json:"notes"`
}

func (h *PatientHandler) GetProfiles(c *gin.Context) {
	accountID := c.GetString("patient_account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient account required"})
		return
	}

	rows, err := h.db.Query(`
		SELECT id, account_id, full_name, relationship, gender, age, blood_group, known_allergies, chronic_conditions, notes, created_at
		FROM patient_profiles WHERE account_id = $1
		ORDER BY CASE WHEN relationship = 'self' THEN 0 ELSE 1 END, created_at ASC
	`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patient profiles"})
		return
	}
	defer rows.Close()

	var profiles []models.PatientProfile
	for rows.Next() {
		var p models.PatientProfile
		if err := rows.Scan(&p.ID, &p.AccountID, &p.FullName, &p.Relationship, &p.Gender, &p.Age, &p.BloodGroup, &p.KnownAllergies, &p.ChronicConditions, &p.Notes, &p.CreatedAt); err == nil {
			profiles = append(profiles, p)
		}
	}

	c.JSON(http.StatusOK, profiles)
}

func (h *PatientHandler) CreateProfile(c *gin.Context) {
	accountID := c.GetString("patient_account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient account required"})
		return
	}

	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profileID := "pprof-" + uuid.New().String()[:8]
	_, err := h.db.Exec(`
		INSERT INTO patient_profiles (id, account_id, full_name, relationship, gender, age, blood_group, known_allergies, chronic_conditions, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`, profileID, accountID, req.FullName, req.Relationship, req.Gender, req.Age, req.BloodGroup, req.KnownAllergies, req.ChronicConditions, req.Notes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Family member profile added successfully",
		"id":      profileID,
	})
}

func (h *PatientHandler) GetMyBookings(c *gin.Context) {
	accountID := c.GetString("patient_account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient account required"})
		return
	}

	rows, err := h.db.Query(`
		SELECT b.id, b.booking_ref, b.patient_account_id, b.patient_profile_id, b.doctor_id, b.specialty_id,
		       b.service_type, b.status, b.chief_complaint, b.triage_notes, b.preferred_time, b.scheduled_at,
		       b.quoted_amount, b.currency, b.payment_id, b.video_room_id, b.created_at, b.updated_at,
		       pp.full_name, pp.relationship,
		       dp.full_name, dp.workplace,
		       sp.name
		FROM bookings b
		LEFT JOIN patient_profiles pp ON b.patient_profile_id = pp.id
		LEFT JOIN doctor_profiles dp ON b.doctor_id = dp.id
		LEFT JOIN specialties sp ON b.specialty_id = sp.id
		WHERE b.patient_account_id = $1
		ORDER BY b.created_at DESC
	`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		var docName, docWork, specName sql.NullString
		var schedAt sql.NullString
		var payID, vroomID, docID, specID sql.NullString

		if err := rows.Scan(
			&b.ID, &b.BookingRef, &b.PatientAccountID, &b.PatientProfileID, &docID, &specID,
			&b.ServiceType, &b.Status, &b.ChiefComplaint, &b.TriageNotes, &b.PreferredTime, &schedAt,
			&b.QuotedAmount, &b.Currency, &payID, &vroomID, &b.CreatedAt, &b.UpdatedAt,
			&b.PatientName, &b.Relationship,
			&docName, &docWork,
			&specName,
		); err == nil {
			if docID.Valid {
				b.DoctorID = &docID.String
			}
			if specID.Valid {
				b.SpecialtyID = &specID.String
			}
			if schedAt.Valid {
				b.ScheduledAt = &schedAt.String
			}
			if payID.Valid {
				b.PaymentID = &payID.String
			}
			if vroomID.Valid {
				b.VideoRoomID = &vroomID.String
			}
			if docName.Valid {
				b.DoctorName = &docName.String
			}
			if docWork.Valid {
				b.DoctorWorkplace = &docWork.String
			}
			if specName.Valid {
				b.SpecialtyName = &specName.String
			}
			bookings = append(bookings, b)
		}
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *PatientHandler) GetMyPrescriptions(c *gin.Context) {
	accountID := c.GetString("patient_account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient account required"})
		return
	}

	rows, err := h.db.Query(`
		SELECT pr.id, pr.booking_id, pr.encounter_id, pr.prescription_ref, pr.doctor_id, pr.patient_profile_id,
		       pr.diagnosis, pr.advice, pr.follow_up_date, pr.warnings, pr.verification_hash, pr.version, pr.created_at,
		       dp.full_name, dp.bmdc_number, dp.qualifications, dp.workplace,
		       pp.full_name, pp.age, pp.gender,
		       b.booking_ref
		FROM prescriptions pr
		JOIN bookings b ON pr.booking_id = b.id
		JOIN patient_profiles pp ON pr.patient_profile_id = pp.id
		JOIN doctor_profiles dp ON pr.doctor_id = dp.id
		WHERE b.patient_account_id = $1
		ORDER BY pr.created_at DESC
	`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prescriptions"})
		return
	}
	defer rows.Close()

	var prescriptions []models.Prescription
	for rows.Next() {
		var p models.Prescription
		if err := rows.Scan(
			&p.ID, &p.BookingID, &p.EncounterID, &p.PrescriptionRef, &p.DoctorID, &p.PatientProfileID,
			&p.Diagnosis, &p.Advice, &p.FollowUpDate, &p.Warnings, &p.VerificationHash, &p.Version, &p.CreatedAt,
			&p.DoctorName, &p.DoctorBMDC, &p.DoctorDegrees, &p.DoctorWorkplace,
			&p.PatientName, &p.PatientAge, &p.PatientGender,
			&p.BookingRef,
		); err == nil {
			// Fetch medicines
			medRows, _ := h.db.Query(`
				SELECT id, prescription_id, brand_name, generic_name, strength, dosage_form, dose, route, frequency, duration, instructions
				FROM medicine_lines WHERE prescription_id = $1
			`, p.ID)
			for medRows.Next() {
				var m models.MedicineLine
				if err := medRows.Scan(&m.ID, &m.PrescriptionID, &m.BrandName, &m.GenericName, &m.Strength, &m.DosageForm, &m.Dose, &m.Route, &m.Frequency, &m.Duration, &m.Instructions); err == nil {
					p.Medicines = append(p.Medicines, m)
				}
			}
			medRows.Close()

			// Fetch tests
			tstRows, _ := h.db.Query(`
				SELECT id, prescription_id, test_name, instructions, status
				FROM test_orders WHERE prescription_id = $1
			`, p.ID)
			for tstRows.Next() {
				var t models.TestOrder
				if err := tstRows.Scan(&t.ID, &t.PrescriptionID, &t.TestName, &t.Instructions, &t.Status); err == nil {
					p.Tests = append(p.Tests, t)
				}
			}
			tstRows.Close()

			prescriptions = append(prescriptions, p)
		}
	}

	c.JSON(http.StatusOK, prescriptions)
}
