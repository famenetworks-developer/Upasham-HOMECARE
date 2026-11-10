package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"upasham-backend/internal/models"
)

type VideoService struct {
	db          *sql.DB
	jitsiDomain string
}

func NewVideoService(db *sql.DB, jitsiDomain string) *VideoService {
	return &VideoService{
		db:          db,
		jitsiDomain: jitsiDomain,
	}
}

// GenerateRandomRoom creates an unguessable cryptographic room name
func GenerateRandomRoom() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("upasham-care-%s", hex.EncodeToString(b))
}

// GeneratePIN creates a 6-digit room entry PIN
func GeneratePIN() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	pin := (int(b[0])<<16|int(b[1])<<8|int(b[2]))%900000 + 100000
	return fmt.Sprintf("%d", pin)
}

// CreateRoomForBooking creates or returns active VideoRoom for a confirmed booking
func (v *VideoService) CreateRoomForBooking(bookingID string) (*models.VideoRoom, error) {
	// Check if already created
	var room models.VideoRoom
	err := v.db.QueryRow(`
		SELECT id, booking_id, room_name, room_pin, host_token, participant_token, jitsi_domain, expires_at, status, created_at
		FROM video_rooms WHERE booking_id = $1
	`, bookingID).Scan(
		&room.ID, &room.BookingID, &room.RoomName, &room.RoomPIN,
		&room.HostToken, &room.ParticipantToken, &room.JitsiDomain,
		&room.ExpiresAt, &room.Status, &room.CreatedAt,
	)

	if err == nil {
		return &room, nil
	}

	// Create new
	roomID := fmt.Sprintf("vroom-%d", time.Now().UnixNano())
	roomName := GenerateRandomRoom()
	pin := GeneratePIN()

	hostTokenBytes := make([]byte, 16)
	_, _ = rand.Read(hostTokenBytes)
	hostToken := hex.EncodeToString(hostTokenBytes)

	partTokenBytes := make([]byte, 16)
	_, _ = rand.Read(partTokenBytes)
	partToken := hex.EncodeToString(partTokenBytes)

	expiresAt := time.Now().Add(4 * time.Hour)

	_, err = v.db.Exec(`
		INSERT INTO video_rooms (id, booking_id, room_name, room_pin, host_token, participant_token, jitsi_domain, expires_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ACTIVE', NOW())
	`, roomID, bookingID, roomName, pin, hostToken, partToken, v.jitsiDomain, expiresAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert video room: %w", err)
	}

	// Update booking with video_room_id
	_, _ = v.db.Exec("UPDATE bookings SET video_room_id = $1 WHERE id = $2", roomID, bookingID)

	room = models.VideoRoom{
		ID:               roomID,
		BookingID:        bookingID,
		RoomName:         roomName,
		RoomPIN:          pin,
		HostToken:        hostToken,
		ParticipantToken: partToken,
		JitsiDomain:      v.jitsiDomain,
		ExpiresAt:        expiresAt,
		Status:           "ACTIVE",
		CreatedAt:        time.Now(),
	}

	return &room, nil
}

// LogAttendance records attendance event for HIPAA/BMDC tele-consult audits
func (v *VideoService) LogAttendance(roomID, userID, role, eventType string) error {
	id := fmt.Sprintf("att-%d", time.Now().UnixNano())
	_, err := v.db.Exec(`
		INSERT INTO attendance_events (id, video_room_id, user_id, role, event_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, id, roomID, userID, role, eventType)
	return err
}

// ValidateMeetingAccess ensures user is either assigned doctor or booking patient
func (v *VideoService) ValidateMeetingAccess(roomID, userID, userRole string) (*models.VideoRoom, bool, error) {
	var room models.VideoRoom
	var patientUserID, doctorUserID sql.NullString
	var scheduledAt sql.NullString
	var bookingStatus string

	err := v.db.QueryRow(`
		SELECT vr.id, vr.booking_id, vr.room_name, vr.room_pin, vr.host_token, vr.participant_token, vr.jitsi_domain, vr.expires_at, vr.status,
		       pa.user_id, dp.user_id, b.scheduled_at, b.status
		FROM video_rooms vr
		JOIN bookings b ON vr.booking_id = b.id
		LEFT JOIN patient_accounts pa ON b.patient_account_id = pa.id
		LEFT JOIN doctor_profiles dp ON b.doctor_id = dp.id
		WHERE vr.id = $1
	`, roomID).Scan(
		&room.ID, &room.BookingID, &room.RoomName, &room.RoomPIN,
		&room.HostToken, &room.ParticipantToken, &room.JitsiDomain,
		&room.ExpiresAt, &room.Status,
		&patientUserID, &doctorUserID, &scheduledAt, &bookingStatus,
	)

	if err != nil {
		return nil, false, errors.New("meeting room not found")
	}

	if time.Now().After(room.ExpiresAt) {
		return nil, false, errors.New("meeting session has expired")
	}
	if bookingStatus != models.BookingStatusConfirmed && bookingStatus != models.BookingStatusInProgress {
		return nil, false, errors.New("consultation is not currently joinable")
	}
	if scheduledAt.Valid && scheduledAt.String != "" {
		if scheduled, parseErr := time.Parse(time.RFC3339, scheduledAt.String); parseErr == nil {
			now := time.Now()
			if now.Before(scheduled.Add(-15*time.Minute)) || now.After(scheduled.Add(2*time.Hour)) {
				return nil, false, errors.New("outside the allowed consultation join window")
			}
		}
	}

	// Super-admin or Operations monitor
	if userRole == models.RoleAdmin || userRole == models.RoleOperations {
		return &room, true, nil
	}

	// Doctor check (Host)
	if doctorUserID.Valid && doctorUserID.String == userID {
		return &room, true, nil
	}

	// Patient check (Participant)
	if patientUserID.Valid && patientUserID.String == userID {
		return &room, false, nil
	}

	return nil, false, errors.New("unauthorized to join this consultation room")
}
