package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"upasham-backend/internal/models"
)

type QuotationService struct {
	db           *sql.DB
	generalPrice float64
}

func NewQuotationService(db *sql.DB, generalPrice float64) *QuotationService {
	return &QuotationService{
		db:           db,
		generalPrice: generalPrice,
	}
}

// CalculateQuotation strictly calculates pricing server-side (Rule 11: never trust client-posted amount)
func (s *QuotationService) CalculateQuotation(serviceType string, doctorID *string) (float64, string, error) {
	if serviceType == "general" {
		return s.generalPrice, "BDT", nil
	}

	if serviceType == "specialist" {
		if doctorID == nil || *doctorID == "" {
			return 1000.0, "BDT", nil // reference base specialist price
		}

		var fee float64
		err := s.db.QueryRow("SELECT consultation_fee FROM doctor_profiles WHERE id = $1", *doctorID).Scan(&fee)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 1000.0, "BDT", nil
			}
			return 0, "", fmt.Errorf("failed to fetch doctor fee: %w", err)
		}
		if fee <= 0 {
			fee = 1000.0
		}
		return fee, "BDT", nil
	}

	return 0, "", errors.New("unsupported service type")
}

// ValidateBookingTransition validates the Section 6 State Machine
// Booking: Draft -> Payment Pending -> Paid/Awaiting Assignment -> Assigned -> Confirmed -> In Progress -> Completed.
// Controlled alternate transitions: Cancelled, No-show, Refund Pending, Refunded.
func ValidateBookingTransition(currentStatus, nextStatus string, userRole string) error {
	validMap := map[string][]string{
		models.BookingStatusDraft: {
			models.BookingStatusPaymentPending,
			models.BookingStatusCancelled,
		},
		models.BookingStatusPaymentPending: {
			models.BookingStatusPaidAwaitingAssignment,
			models.BookingStatusCancelled,
		},
		models.BookingStatusPaidAwaitingAssignment: {
			models.BookingStatusAssigned,
			models.BookingStatusConfirmed,
			models.BookingStatusCancelled,
			models.BookingStatusRefundPending,
		},
		models.BookingStatusAssigned: {
			models.BookingStatusConfirmed,
			models.BookingStatusCancelled,
			models.BookingStatusPaidAwaitingAssignment, // re-assign
		},
		models.BookingStatusConfirmed: {
			models.BookingStatusInProgress,
			models.BookingStatusCancelled,
			models.BookingStatusNoShow,
			models.BookingStatusAssigned, // reschedule
		},
		models.BookingStatusInProgress: {
			models.BookingStatusCompleted,
			models.BookingStatusCancelled,
		},
		models.BookingStatusRefundPending: {
			models.BookingStatusRefunded,
			models.BookingStatusCancelled,
		},
	}

	allowed, exists := validMap[currentStatus]
	if !exists {
		// Terminal states cannot transition normally
		if currentStatus == models.BookingStatusCompleted || currentStatus == models.BookingStatusRefunded {
			return fmt.Errorf("booking in terminal status '%s' cannot be altered", currentStatus)
		}
	}

	for _, a := range allowed {
		if a == nextStatus {
			return nil
		}
	}

	// Super-admin escape hatch with audit logging
	if userRole == models.RoleAdmin {
		return nil
	}

	return fmt.Errorf("illegal state transition from '%s' to '%s'", currentStatus, nextStatus)
}

// RecordStatusHistory writes to booking_status_histories
func RecordStatusHistory(db *sql.DB, bookingID, fromStatus, toStatus, changedByUserID, notes string) error {
	id := fmt.Sprintf("bsh-%d", time.Now().UnixNano())
	_, err := db.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, id, bookingID, fromStatus, toStatus, changedByUserID, notes)
	return err
}
