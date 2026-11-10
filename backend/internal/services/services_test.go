package services_test

import (
	"strings"
	"testing"

	"upasham-backend/internal/models"
	"upasham-backend/internal/services"
)

func TestStateTransitions(t *testing.T) {
	// Valid flow: Draft -> Payment Pending
	err := services.ValidateBookingTransition(models.BookingStatusDraft, models.BookingStatusPaymentPending, models.RolePatient)
	if err != nil {
		t.Fatalf("expected valid transition from Draft to Payment Pending, got %v", err)
	}

	// Valid flow: Payment Pending -> Paid/Awaiting Assignment
	err = services.ValidateBookingTransition(models.BookingStatusPaymentPending, models.BookingStatusPaidAwaitingAssignment, models.RolePatient)
	if err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}

	// Valid flow: Paid/Awaiting Assignment -> Assigned
	err = services.ValidateBookingTransition(models.BookingStatusPaidAwaitingAssignment, models.BookingStatusAssigned, models.RoleOperations)
	if err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}

	// Valid flow: Assigned -> Confirmed
	err = services.ValidateBookingTransition(models.BookingStatusAssigned, models.BookingStatusConfirmed, models.RoleOperations)
	if err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}

	// Valid flow: Confirmed -> In Progress
	err = services.ValidateBookingTransition(models.BookingStatusConfirmed, models.BookingStatusInProgress, models.RoleDoctor)
	if err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}

	// Valid flow: In Progress -> Completed
	err = services.ValidateBookingTransition(models.BookingStatusInProgress, models.BookingStatusCompleted, models.RoleDoctor)
	if err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}

	// ILLEGAL TRANSITION: Draft -> Completed directly
	err = services.ValidateBookingTransition(models.BookingStatusDraft, models.BookingStatusCompleted, models.RolePatient)
	if err == nil {
		t.Fatalf("expected error on illegal transition Draft -> Completed, but passed")
	}

	// ILLEGAL TRANSITION: Completed -> In Progress
	err = services.ValidateBookingTransition(models.BookingStatusCompleted, models.BookingStatusInProgress, models.RoleDoctor)
	if err == nil {
		t.Fatalf("expected error on modifying terminal state Completed, but passed")
	}
}

func TestPaymentAdapters(t *testing.T) {
	bkash := services.NewBkashAdapter()
	ssl := services.NewSSLCommerzAdapter()

	payment := &models.Payment{
		Amount:     20.0,
		InvoiceRef: "INV-TEST-001",
	}

	// bKash initiation
	bkUrl, bkTrx, err := bkash.InitiatePayment(payment, "+8801700000000", "/return")
	if err != nil || bkUrl == "" || bkTrx == "" {
		t.Fatalf("failed bkash payment initiate: %v", err)
	}

	// bKash callback verification
	ok, status, trx, err := bkash.VerifyCallback(map[string]interface{}{
		"status": "success",
		"trxID":  bkTrx,
	})
	if !ok || status != models.PaymentStatusPaid || trx != bkTrx {
		t.Fatalf("bkash callback verification failed: ok=%v, status=%s", ok, status)
	}

	// SSLCOMMERZ initiation
	sslUrl, sslTrx, err := ssl.InitiatePayment(payment, "+8801700000000", "/return")
	if err != nil || sslUrl == "" || sslTrx == "" {
		t.Fatalf("failed sslcommerz payment initiate: %v", err)
	}

	// SSLCOMMERZ callback verification
	ok, status, trx, err = ssl.VerifyCallback(map[string]interface{}{
		"tran_status": "VALID",
		"tran_id":     sslTrx,
	})
	if !ok || status != models.PaymentStatusPaid || trx != sslTrx {
		t.Fatalf("sslcommerz callback verification failed: ok=%v, status=%s", ok, status)
	}
}

func TestPayloadRedaction(t *testing.T) {
	raw := map[string]interface{}{
		"password":     "superSecret123",
		"store_passwd": "storeSecret999",
		"card_number":  "4111111111111111",
		"amount":       20.0,
		"trxID":        "BK999123",
	}
	redacted := services.RedactGatewayPayload(raw)
	if len(redacted) == 0 {
		t.Fatalf("redacted payload is empty")
	}
	if strings.Contains(redacted, "superSecret123") || strings.Contains(redacted, "storeSecret999") {
		t.Fatalf("redacted payload still contains secrets!")
	}
}

func TestVideoRoomGeneration(t *testing.T) {
	room1 := services.GenerateRandomRoom()
	room2 := services.GenerateRandomRoom()
	if room1 == room2 {
		t.Fatalf("rooms must be cryptographically random and unique")
	}
	pin := services.GeneratePIN()
	if len(pin) != 6 {
		t.Fatalf("expected 6-digit pin, got %s", pin)
	}
}
