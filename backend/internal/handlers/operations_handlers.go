package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"upasham-backend/internal/models"
	"upasham-backend/internal/services"
)

type OperationsHandler struct {
	db        *sql.DB
	videoServ *services.VideoService
}

func NewOperationsHandler(db *sql.DB, videoServ *services.VideoService) *OperationsHandler {
	return &OperationsHandler{
		db:        db,
		videoServ: videoServ,
	}
}

type AssignDoctorRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
	DoctorID  string `json:"doctor_id" binding:"required"`
	Notes     string `json:"notes"`
}

type ConfirmAppointmentRequest struct {
	BookingID   string `json:"booking_id" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"` // e.g. "2026-10-13T20:30:00Z"
	Notes       string `json:"notes"`
}

type RescheduleRequest struct {
	BookingID   string `json:"booking_id" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"`
	Reason      string `json:"reason" binding:"required"`
}

type CancelBookingRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}

// GetPaidQueue lists bookings awaiting manual doctor assignment
func (h *OperationsHandler) GetPaidQueue(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT b.id, b.booking_ref, b.patient_account_id, b.patient_profile_id, b.doctor_id, b.specialty_id,
		       b.service_type, b.status, b.chief_complaint, b.triage_notes, b.preferred_time, b.scheduled_at,
		       b.quoted_amount, b.currency, b.payment_id, b.created_at,
		       pp.full_name, pp.relationship, pp.age, pp.gender,
		       p.gateway, p.gateway_tx_id, p.amount, p.paid_at,
		       dp.full_name, sp.name
		FROM bookings b
		JOIN patient_profiles pp ON b.patient_profile_id = pp.id
		JOIN payments p ON b.id = p.booking_id AND p.status = 'Paid'
		LEFT JOIN doctor_profiles dp ON b.doctor_id = dp.id
		LEFT JOIN specialties sp ON b.specialty_id = sp.id
		WHERE b.status IN ('Paid/Awaiting Assignment', 'Assigned', 'Confirmed')
		ORDER BY CASE WHEN b.status = 'Paid/Awaiting Assignment' THEN 0 WHEN b.status = 'Assigned' THEN 1 ELSE 2 END, b.created_at ASC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch operations queue: " + err.Error()})
		return
	}
	defer rows.Close()

	var queue []gin.H
	for rows.Next() {
		var b models.Booking
		var docID, specID, schedAt, payID sql.NullString
		var docName, specName sql.NullString
		var patName, rel, gender string
		var age int
		var gw, txID string
		var pAmount float64
		var paidAt *time.Time

		if err := rows.Scan(
			&b.ID, &b.BookingRef, &b.PatientAccountID, &b.PatientProfileID, &docID, &specID,
			&b.ServiceType, &b.Status, &b.ChiefComplaint, &b.TriageNotes, &b.PreferredTime, &schedAt,
			&b.QuotedAmount, &b.Currency, &payID, &b.CreatedAt,
			&patName, &rel, &age, &gender,
			&gw, &txID, &pAmount, &paidAt,
			&docName, &specName,
		); err == nil {
			queue = append(queue, gin.H{
				"id":               b.ID,
				"booking_ref":      b.BookingRef,
				"service_type":     b.ServiceType,
				"status":           b.Status,
				"chief_complaint":  b.ChiefComplaint,
				"preferred_time":   b.PreferredTime,
				"scheduled_at":     schedAt.String,
				"created_at":       b.CreatedAt,
				"patient_name":     patName,
				"relationship":     rel,
				"patient_age":      age,
				"patient_gender":   gender,
				"doctor_name":      docName.String,
				"doctor_id":        docID.String,
				"specialty_name":   specName.String,
				"gateway":          gw,
				"gateway_tx_id":    txID,
				"paid_amount":      pAmount,
				"paid_at":          paidAt,
			})
		}
	}

	c.JSON(http.StatusOK, queue)
}

// AssignDoctor manually assigns a physician to the booking
func (h *OperationsHandler) AssignDoctor(c *gin.Context) {
	var req AssignDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opsUserID := c.GetString("user_id")

	var currentStatus string
	err := h.db.QueryRow("SELECT status FROM bookings WHERE id = $1", req.BookingID).Scan(&currentStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Must be Paid/Awaiting Assignment or Assigned (reassign)
	if currentStatus != models.BookingStatusPaidAwaitingAssignment && currentStatus != models.BookingStatusAssigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot assign doctor in status: " + currentStatus})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Update booking
	_, err = tx.Exec(`
		UPDATE bookings
		SET doctor_id = $1, status = 'Assigned', updated_at = NOW()
		WHERE id = $2
	`, req.DoctorID, req.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
		return
	}

	// Insert into assignments table
	assignID := fmt.Sprintf("asg-%d", time.Now().UnixNano())
	_, _ = tx.Exec(`
		INSERT INTO assignments (id, booking_id, doctor_id, assigned_by_user_id, status, assigned_at)
		VALUES ($1, $2, $3, $4, 'ACCEPTED', NOW())
	`, assignID, req.BookingID, req.DoctorID, opsUserID)

	// Status history
	_, _ = tx.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, $3, 'Assigned', $4, $5, NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), req.BookingID, currentStatus, opsUserID, "Doctor assigned by Operations: "+req.Notes)

	// Audit Log
	_, _ = tx.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'ASSIGN_DOCTOR', 'Booking', $3, $4, $5, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), opsUserID, req.BookingID, c.ClientIP(), fmt.Sprintf(`{"doctor_id": "%s"}`, req.DoctorID))

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Doctor assigned successfully",
		"booking_id": req.BookingID,
		"status":     models.BookingStatusAssigned,
	})
}

// ConfirmAppointment schedules consultation time and generates secure Jitsi room
func (h *OperationsHandler) ConfirmAppointment(c *gin.Context) {
	var req ConfirmAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opsUserID := c.GetString("user_id")

	var currentStatus, doctorID sql.NullString
	err := h.db.QueryRow("SELECT status, doctor_id FROM bookings WHERE id = $1", req.BookingID).Scan(&currentStatus, &doctorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if !doctorID.Valid || doctorID.String == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A doctor must be assigned before confirming appointment"})
		return
	}
	if err := services.ValidateBookingTransition(currentStatus.String, models.BookingStatusConfirmed, models.RoleOperations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := time.Parse(time.RFC3339, req.ScheduledAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be an RFC3339 UTC timestamp"})
		return
	}

	// Create Video Room for consultation
	room, err := h.videoServ.CreateRoomForBooking(req.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to provision video room: " + err.Error()})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Update booking status
	_, err = tx.Exec(`
		UPDATE bookings
		SET status = 'Confirmed', scheduled_at = $1, video_room_id = $2, updated_at = NOW()
		WHERE id = $3
	`, req.ScheduledAt, room.ID, req.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
		return
	}

	// Status history
	_, _ = tx.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, $3, 'Confirmed', $4, $5, NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), req.BookingID, currentStatus.String, opsUserID, "Appointment confirmed for "+req.ScheduledAt)

	// Audit Log
	_, _ = tx.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'CONFIRM_BOOKING', 'Booking', $3, $4, $5, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), opsUserID, req.BookingID, c.ClientIP(), fmt.Sprintf(`{"scheduled_at": "%s", "room_id": "%s"}`, req.ScheduledAt, room.ID))

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Appointment confirmed successfully",
		"booking_id":   req.BookingID,
		"status":       models.BookingStatusConfirmed,
		"scheduled_at": req.ScheduledAt,
		"room_name":    room.RoomName,
	})
}

func (h *OperationsHandler) Reschedule(c *gin.Context) {
	var req RescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	opsUserID := c.GetString("user_id")
	if _, err := time.Parse(time.RFC3339, req.ScheduledAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be an RFC3339 UTC timestamp"})
		return
	}
	var currentStatus string
	if err := h.db.QueryRow("SELECT status FROM bookings WHERE id = $1", req.BookingID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}
	if currentStatus != models.BookingStatusConfirmed && currentStatus != models.BookingStatusAssigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking cannot be rescheduled in status: " + currentStatus})
		return
	}

	_, err := h.db.Exec(`
		UPDATE bookings SET scheduled_at = $1, updated_at = NOW() WHERE id = $2
	`, req.ScheduledAt, req.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reschedule"})
		return
	}

	_, _ = h.db.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, $3, $3, $4, $5, NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), req.BookingID, currentStatus, opsUserID, "Rescheduled: "+req.Reason)
	_, _ = h.db.Exec(`INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'RESCHEDULE_BOOKING', 'Booking', $3, $4, $5, NOW())`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), opsUserID, req.BookingID, c.ClientIP(), fmt.Sprintf(`{"scheduled_at":%q}`, req.ScheduledAt))

	c.JSON(http.StatusOK, gin.H{"message": "Booking rescheduled to " + req.ScheduledAt})
}

func (h *OperationsHandler) Cancel(c *gin.Context) {
	var req CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	opsUserID := c.GetString("user_id")
	var currentStatus string
	if err := h.db.QueryRow("SELECT status FROM bookings WHERE id = $1", req.BookingID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}
	if err := services.ValidateBookingTransition(currentStatus, models.BookingStatusCancelled, models.RoleOperations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`
		UPDATE bookings SET status = 'Cancelled', updated_at = NOW() WHERE id = $1
	`, req.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel"})
		return
	}

	_, _ = h.db.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, $3, 'Cancelled', $4, $5, NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), req.BookingID, currentStatus, opsUserID, "Cancelled: "+req.Reason)
	_, _ = h.db.Exec(`INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'CANCEL_BOOKING', 'Booking', $3, $4, $5, NOW())`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), opsUserID, req.BookingID, c.ClientIP(), fmt.Sprintf(`{"reason":%q,"from_status":%q}`, req.Reason, currentStatus))

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled"})
}
