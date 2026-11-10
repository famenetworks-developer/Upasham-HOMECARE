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

type PaymentHandler struct {
	db                *sql.DB
	bkashAdapter      *services.BkashAdapter
	sslcommerzAdapter *services.SSLCommerzAdapter
}

func NewPaymentHandler(db *sql.DB) *PaymentHandler {
	return &PaymentHandler{
		db:                db,
		bkashAdapter:      services.NewBkashAdapter(),
		sslcommerzAdapter: services.NewSSLCommerzAdapter(),
	}
}

type InitiatePaymentRequest struct {
	BookingID     string `json:"booking_id" binding:"required"`
	Gateway       string `json:"gateway" binding:"required"` // bkash, sslcommerz
	CustomerPhone string `json:"customer_phone"`
	ReturnURL     string `json:"return_url"`
}

type CallbackPayload struct {
	InvoiceRef string                 `json:"invoice_ref" binding:"required"`
	Gateway    string                 `json:"gateway" binding:"required"`
	Status     string                 `json:"status" binding:"required"` // success, failed
	TrxID      string                 `json:"trx_id"`
	RawData    map[string]interface{} `json:"raw_data"`
}

func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
	accountID := c.GetString("patient_account_id")
	var req InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch booking
	var b models.Booking
	err := h.db.QueryRow(`
		SELECT id, booking_ref, quoted_amount, currency, status
		FROM bookings WHERE id = $1 AND patient_account_id = $2
	`, req.BookingID, accountID).Scan(&b.ID, &b.BookingRef, &b.QuotedAmount, &b.Currency, &b.Status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found for this patient account"})
		return
	}

	if b.Status != models.BookingStatusPaymentPending && b.Status != models.BookingStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking is not in payable status: " + b.Status})
		return
	}

	paymentID := "pay-" + uuid.New().String()[:8]
	invoiceRef := fmt.Sprintf("INV-%s-%d", b.BookingRef, time.Now().Unix()%10000)

	payment := models.Payment{
		ID:         paymentID,
		BookingID:  b.ID,
		InvoiceRef: invoiceRef,
		Gateway:    req.Gateway,
		Amount:     b.QuotedAmount,
		Currency:   b.Currency,
		Status:     models.PaymentStatusInitiated,
	}

	var checkoutURL, gatewayTxID string
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = "/patient/payment-callback"
	}

	if req.Gateway == "bkash" {
		checkoutURL, gatewayTxID, err = h.bkashAdapter.InitiatePayment(&payment, req.CustomerPhone, returnURL)
	} else if req.Gateway == "sslcommerz" {
		checkoutURL, gatewayTxID, err = h.sslcommerzAdapter.InitiatePayment(&payment, req.CustomerPhone, returnURL)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported payment gateway"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gateway initialization failed: " + err.Error()})
		return
	}

	// Persist payment record
	_, err = h.db.Exec(`
		INSERT INTO payments (id, booking_id, invoice_ref, gateway, amount, currency, status, gateway_tx_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`, paymentID, b.ID, invoiceRef, req.Gateway, b.QuotedAmount, b.Currency, models.PaymentStatusInitiated, gatewayTxID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist payment"})
		return
	}

	// Update booking payment_id
	_, _ = h.db.Exec("UPDATE bookings SET payment_id = $1 WHERE id = $2", paymentID, b.ID)

	c.JSON(http.StatusOK, gin.H{
		"payment_id":    paymentID,
		"invoice_ref":   invoiceRef,
		"gateway":       req.Gateway,
		"amount":        b.QuotedAmount,
		"currency":      b.Currency,
		"checkout_url":  checkoutURL,
		"gateway_tx_id": gatewayTxID,
	})
}

// Callback processes gateway response idempotently
func (h *PaymentHandler) Callback(c *gin.Context) {
	var req CallbackPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p models.Payment
	err := h.db.QueryRow(`
		SELECT id, booking_id, invoice_ref, gateway, amount, currency, status
		FROM payments WHERE invoice_ref = $1
	`, req.InvoiceRef).Scan(&p.ID, &p.BookingID, &p.InvoiceRef, &p.Gateway, &p.Amount, &p.Currency, &p.Status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment reference not found"})
		return
	}
	if req.Gateway != p.Gateway {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gateway does not match the initiated payment"})
		return
	}

	// IDEMPOTENCY CHECK: If already paid, return 200 immediately
	if p.Status == models.PaymentStatusPaid {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ALREADY_PAID",
			"message": "Payment has already been confirmed idempotently",
		})
		return
	}

	// Verify gateway response
	verified := false
	paymentStatus := models.PaymentStatusFailed
	txID := req.TrxID
	if req.RawData == nil {
		req.RawData = map[string]interface{}{"status": req.Status, "trxID": req.TrxID}
	}

	if req.Gateway == "bkash" {
		verified, paymentStatus, txID, _ = h.bkashAdapter.VerifyCallback(req.RawData)
	} else if req.Gateway == "sslcommerz" {
		verified, paymentStatus, txID, _ = h.sslcommerzAdapter.VerifyCallback(req.RawData)
	}

	redactedJSON := services.RedactGatewayPayload(req.RawData)
	eventID := fmt.Sprintf("gev-%d", time.Now().UnixNano())
	_, _ = h.db.Exec(`
		INSERT INTO gateway_events (id, payment_id, gateway, event_type, payload_json, signature, is_verified, created_at)
		VALUES ($1, $2, $3, 'CALLBACK', $4, 'sig_verified', $5, NOW())
	`, eventID, p.ID, req.Gateway, redactedJSON, verified)

	if !verified || paymentStatus != models.PaymentStatusPaid {
		_, _ = h.db.Exec("UPDATE payments SET status = $1 WHERE id = $2", models.PaymentStatusFailed, p.ID)
		c.JSON(http.StatusOK, gin.H{"status": "FAILED", "message": "Payment verification rejected"})
		return
	}

	// Execute State Transition: Payment Pending -> Paid/Awaiting Assignment
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Update Payment
	_, err = tx.Exec(`
		UPDATE payments 
		SET status = $1, gateway_tx_id = $2, paid_at = NOW() 
		WHERE id = $3
	`, models.PaymentStatusPaid, txID, p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Update Booking status to "Paid/Awaiting Assignment"
	_, err = tx.Exec(`
		UPDATE bookings 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
	`, models.BookingStatusPaidAwaitingAssignment, p.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking status"})
		return
	}

	// Record Status History
	_, _ = tx.Exec(`
		INSERT INTO booking_status_histories (id, booking_id, from_status, to_status, changed_by_user_id, notes, created_at)
		VALUES ($1, $2, 'Payment Pending', 'Paid/Awaiting Assignment', 'GATEWAY_CALLBACK', 'Successful payment confirmation via ' || $3, NOW())
	`, fmt.Sprintf("bsh-%d", time.Now().UnixNano()), p.BookingID, req.Gateway)

	// Audit Log
	_, _ = tx.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, 'SYSTEM', 'PAYMENT_VERIFIED', 'Payment', $2, $3, $4, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), p.ID, c.ClientIP(), fmt.Sprintf(`{"amount": %.2f, "gateway": "%s", "tx_id": "%s"}`, p.Amount, req.Gateway, txID))

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "SUCCESS",
		"booking_id":  p.BookingID,
		"new_status":  models.BookingStatusPaidAwaitingAssignment,
		"amount_paid": p.Amount,
		"currency":    p.Currency,
		"gateway":     req.Gateway,
		"tx_id":       txID,
	})
}

// SimulatePayment enables seamless interactive test execution for demo & evaluation
func (h *PaymentHandler) SimulatePayment(c *gin.Context) {
	bookingID := c.Query("booking_id")
	gateway := c.Query("gateway")
	if gateway == "" {
		gateway = "bkash"
	}

	var p models.Payment
	err := h.db.QueryRow(`
		SELECT id, booking_id, invoice_ref, gateway, amount, currency, status
		FROM payments WHERE booking_id = $1 ORDER BY created_at DESC LIMIT 1
	`, bookingID).Scan(&p.ID, &p.BookingID, &p.InvoiceRef, &p.Gateway, &p.Amount, &p.Currency, &p.Status)

	if err != nil {
		// If payment record not created yet, create one
		var quotedAmount float64
		var bRef string
		_ = h.db.QueryRow("SELECT quoted_amount, booking_ref FROM bookings WHERE id = $1", bookingID).Scan(&quotedAmount, &bRef)
		if quotedAmount <= 0 {
			quotedAmount = 20.0
		}
		p.ID = "pay-" + uuid.New().String()[:8]
		p.BookingID = bookingID
		p.InvoiceRef = fmt.Sprintf("INV-%s-%d", bRef, time.Now().Unix()%10000)
		p.Gateway = gateway
		p.Amount = quotedAmount
		p.Currency = "BDT"
		p.Status = models.PaymentStatusInitiated

		_, _ = h.db.Exec(`
			INSERT INTO payments (id, booking_id, invoice_ref, gateway, amount, currency, status, gateway_tx_id, created_at)
			VALUES ($1, $2, $3, $4, $5, 'BDT', 'Initiated', 'SIM-TX', NOW())
		`, p.ID, p.BookingID, p.InvoiceRef, p.Gateway, p.Amount)
		_, _ = h.db.Exec("UPDATE bookings SET payment_id = $1 WHERE id = $2", p.ID, bookingID)
	}

	txID := fmt.Sprintf("SIM%d", time.Now().UnixNano()%1000000000)
	c.Request.Body = http.NoBody // reuse callback
	c.JSON(http.StatusOK, gin.H{
		"invoice_ref": p.InvoiceRef,
		"gateway":     p.Gateway,
		"status":      "success",
		"trx_id":      txID,
		"amount":      p.Amount,
		"message":     "Sandbox simulation payment ready. Send to /api/payments/callback to confirm.",
	})
}

func (h *PaymentHandler) ListTransactions(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT p.id, p.booking_id, p.invoice_ref, p.gateway, p.amount, p.currency, p.status, p.gateway_tx_id, p.paid_at, p.created_at,
		       pp.full_name, b.service_type, b.booking_ref
		FROM payments p
		JOIN bookings b ON p.booking_id = b.id
		JOIN patient_profiles pp ON b.patient_profile_id = pp.id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var p models.Payment
		var txID, sig sql.NullString
		var paidAt *time.Time
		var patName, srvType, bkRef string

		if err := rows.Scan(
			&p.ID, &p.BookingID, &p.InvoiceRef, &p.Gateway, &p.Amount, &p.Currency, &p.Status, &txID, &paidAt, &p.CreatedAt,
			&patName, &srvType, &bkRef,
		); err == nil {
			if txID.Valid {
				p.GatewayTxID = txID.String
			}
			if sig.Valid {
				p.GatewaySignature = sig.String
			}
			p.PaidAt = paidAt
			p.PatientName = patName
			p.ServiceType = srvType
			p.BookingRef = bkRef
			payments = append(payments, p)
		}
	}

	c.JSON(http.StatusOK, payments)
}
