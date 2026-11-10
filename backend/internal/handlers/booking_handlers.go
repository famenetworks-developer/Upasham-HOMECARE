package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"upasham-backend/internal/models"
	"upasham-backend/internal/services"
)

type BookingHandler struct {
	db        *sql.DB
	quoteServ *services.QuotationService
}

func NewBookingHandler(db *sql.DB, quoteServ *services.QuotationService) *BookingHandler {
	return &BookingHandler{
		db:        db,
		quoteServ: quoteServ,
	}
}

type QuotationRequest struct {
	ServiceType string  `json:"service_type" binding:"required"` // general, specialist
	DoctorID    *string `json:"doctor_id"`
}

type CreateBookingRequest struct {
	PatientProfileID string  `json:"patient_profile_id" binding:"required"`
	ServiceType      string  `json:"service_type" binding:"required"` // general, specialist
	DoctorID         *string `json:"doctor_id"`
	SpecialtyID      *string `json:"specialty_id"`
	ChiefComplaint   string  `json:"chief_complaint" binding:"required"`
	TriageNotes      string  `json:"triage_notes"`
	PreferredTime    string  `json:"preferred_time" binding:"required"`
	ConsentGiven     bool    `json:"consent_given" binding:"required"`
}

func (h *BookingHandler) GetQuotation(c *gin.Context) {
	var req QuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount, currency, err := h.quoteServ.CalculateQuotation(req.ServiceType, req.DoctorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_type": req.ServiceType,
		"amount":       amount,
		"currency":     currency,
		"notes":        "Official server-persisted quotation for Dhaka connected-care consultation",
	})
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	accountID := c.GetString("patient_account_id")
	userID := c.GetString("user_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient account required"})
		return
	}

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !req.ConsentGiven {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telehealth and data processing consent is mandatory"})
		return
	}

	// Verify patient profile belongs to this account (Object-level authorization)
	var profCount int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM patient_profiles WHERE id = $1 AND account_id = $2", req.PatientProfileID, accountID).Scan(&profCount)
	if profCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid patient profile for your account"})
		return
	}

	// Calculate and persist quotation strictly on server
	amount, currency, err := h.quoteServ.CalculateQuotation(req.ServiceType, req.DoctorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to calculate quotation: " + err.Error()})
		return
	}

	bookingID := "bk-" + uuid.New().String()[:8]
	bookingRef := fmt.Sprintf("UP-2026-%d", time.Now().UnixNano()%1000000)

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO bookings (id, booking_ref, patient_account_id, patient_profile_id, doctor_id, specialty_id,
		                      service_type, status, chief_complaint, triage_notes, preferred_time, quoted_amount, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`, bookingID, bookingRef, accountID, req.PatientProfileID, req.DoctorID, req.SpecialtyID,
		req.ServiceType, models.BookingStatusPaymentPending, req.ChiefComplaint, req.TriageNotes, req.PreferredTime, amount, currency)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist booking: " + err.Error()})
		return
	}

	// Record Consent
	_, _ = tx.Exec(`
		INSERT INTO consents (id, booking_id, patient_account_id, consent_type, accepted_at)
		VALUES ($1, $2, $3, 'TELEHEALTH_CONSENT', NOW())
	`, fmt.Sprintf("cns-%d", time.Now().UnixNano()), bookingID, accountID)

	// Record Status History
	_, _ = tx.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, 'Draft', 'Payment Pending', $3, 'Booking initiated, awaiting payment', NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), bookingID, userID)

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"booking_id":    bookingID,
		"booking_ref":   bookingRef,
		"status":        models.BookingStatusPaymentPending,
		"quoted_amount": amount,
		"currency":      currency,
		"message":       "Booking created. Please complete payment through bKash or SSLCOMMERZ.",
	})
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	id := c.Param("id")
	userRole := c.GetString("role")
	patientAccountID := c.GetString("patient_account_id")
	doctorID := c.GetString("doctor_id")

	var b models.Booking
	var docID, specID, schedAt, payID, vroomID sql.NullString
	var docName, specName, patName, rel string

	err := h.db.QueryRow(`
		SELECT b.id, b.booking_ref, b.patient_account_id, b.patient_profile_id, b.doctor_id, b.specialty_id,
		       b.service_type, b.status, b.chief_complaint, b.triage_notes, b.preferred_time, b.scheduled_at,
		       b.quoted_amount, b.currency, b.payment_id, b.video_room_id, b.created_at, b.updated_at,
		       pp.full_name, pp.relationship,
		       dp.full_name, sp.name
		FROM bookings b
		JOIN patient_profiles pp ON b.patient_profile_id = pp.id
		LEFT JOIN doctor_profiles dp ON b.doctor_id = dp.id
		LEFT JOIN specialties sp ON b.specialty_id = sp.id
		WHERE b.id = $1 OR b.booking_ref = $1
	`, id).Scan(
		&b.ID, &b.BookingRef, &b.PatientAccountID, &b.PatientProfileID, &docID, &specID,
		&b.ServiceType, &b.Status, &b.ChiefComplaint, &b.TriageNotes, &b.PreferredTime, &schedAt,
		&b.QuotedAmount, &b.Currency, &payID, &vroomID, &b.CreatedAt, &b.UpdatedAt,
		&patName, &rel, &docName, &specName,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Object-level permission check:
	// Patient can only view their own booking
	// Doctor can only view if assigned
	// Operations/Admin/Finance can view
	if userRole == models.RolePatient && b.PatientAccountID != patientAccountID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view this booking"})
		return
	}
	if userRole == models.RoleDoctor && (docID.Valid && docID.String != doctorID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not assigned to this booking"})
		return
	}

	if docID.Valid {
		b.DoctorID = &docID.String
		b.DoctorName = &docName
	}
	if specID.Valid {
		b.SpecialtyID = &specID.String
		b.SpecialtyName = &specName
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
	b.PatientName = patName
	b.Relationship = rel

	c.JSON(http.StatusOK, b)
}
