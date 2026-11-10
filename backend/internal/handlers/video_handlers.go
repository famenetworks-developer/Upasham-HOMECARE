package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"upasham-backend/internal/models"
	"upasham-backend/internal/services"
)

type VideoHandler struct {
	db        *sql.DB
	videoServ *services.VideoService
}

func NewVideoHandler(db *sql.DB, videoServ *services.VideoService) *VideoHandler {
	return &VideoHandler{
		db:        db,
		videoServ: videoServ,
	}
}

type AttendanceRequest struct {
	VideoRoomID string `json:"video_room_id" binding:"required"`
	EventType   string `json:"event_type" binding:"required"` // join, leave
}

func (h *VideoHandler) GetRoomAccess(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString("user_id")
	userRole := c.GetString("role")

	room, isHost, err := h.videoServ.ValidateMeetingAccess(roomID, userID, userRole)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Fetch booking details
	var b models.Booking
	var patName, docName string
	_ = h.db.QueryRow(`
		SELECT b.id, b.booking_ref, b.status, b.chief_complaint, b.scheduled_at,
		       pp.full_name, dp.full_name
		FROM bookings b
		JOIN video_rooms vr ON vr.booking_id = b.id
		JOIN patient_profiles pp ON b.patient_profile_id = pp.id
		LEFT JOIN doctor_profiles dp ON b.doctor_id = dp.id
		WHERE vr.id = $1
	`, roomID).Scan(&b.ID, &b.BookingRef, &b.Status, &b.ChiefComplaint, &b.ScheduledAt, &patName, &docName)

	token := room.ParticipantToken
	if isHost {
		token = room.HostToken
		// Mark booking "In Progress" if currently "Confirmed"
		if b.Status == models.BookingStatusConfirmed {
			_, _ = h.db.Exec("UPDATE bookings SET status = 'In Progress', updated_at = NOW() WHERE id = $1", b.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id":         room.ID,
		"room_name":       room.RoomName,
		"room_pin":        room.RoomPIN,
		"jitsi_domain":    room.JitsiDomain,
		"access_token":    token,
		"is_host":         isHost,
		"expires_at":      room.ExpiresAt,
		"booking_ref":     b.BookingRef,
		"patient_name":    patName,
		"doctor_name":     docName,
		"chief_complaint": b.ChiefComplaint,
		"fallback_phone":  "+880 9610-UPASHAM (Emergency Support)",
	})
}

func (h *VideoHandler) LogAttendance(c *gin.Context) {
	var req AttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userRole := c.GetString("role")

	_ = h.videoServ.LogAttendance(req.VideoRoomID, userID, userRole, req.EventType)
	c.JSON(http.StatusOK, gin.H{"status": "logged"})
}
