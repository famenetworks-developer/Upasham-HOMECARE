package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"upasham-backend/internal/models"
)

type FinanceHandler struct {
	db *sql.DB
}

func NewFinanceHandler(db *sql.DB) *FinanceHandler {
	return &FinanceHandler{db: db}
}

type ProcessRefundRequest struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Reason    string  `json:"reason" binding:"required"`
}

func (h *FinanceHandler) GetDashboard(c *gin.Context) {
	var totalRevenue, bkashRevenue, sslRevenue float64
	var paidCount, refundedCount int

	_ = h.db.QueryRow("SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM payments WHERE status = 'Paid'").Scan(&totalRevenue, &paidCount)
	_ = h.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'Paid' AND gateway = 'bkash'").Scan(&bkashRevenue)
	_ = h.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'Paid' AND gateway = 'sslcommerz'").Scan(&sslRevenue)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM refunds WHERE status = 'PROCESSED'").Scan(&refundedCount)

	// Fetch recent gateway events
	gwRows, err := h.db.Query(`
		SELECT ge.id, ge.payment_id, ge.gateway, ge.event_type, ge.payload_json, ge.is_verified, ge.created_at,
		       p.invoice_ref, p.amount
		FROM gateway_events ge
		JOIN payments p ON ge.payment_id = p.id
		ORDER BY ge.created_at DESC LIMIT 10
	`)
	var events []gin.H
	if err == nil {
		defer gwRows.Close()
		for gwRows.Next() {
			var id, pid, gw, evType, payload, invRef string
			var isVer bool
			var cat time.Time
			var amt float64
			if err := gwRows.Scan(&id, &pid, &gw, &evType, &payload, &isVer, &cat, &invRef, &amt); err == nil {
				events = append(events, gin.H{
					"id":          id,
					"payment_id":  pid,
					"invoice_ref": invRef,
					"gateway":     gw,
					"event_type":  evType,
					"payload":     payload,
					"is_verified": isVer,
					"amount":      amt,
					"created_at":  cat,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": gin.H{
			"total_revenue":   totalRevenue,
			"paid_count":      paidCount,
			"bkash_revenue":   bkashRevenue,
			"ssl_revenue":     sslRevenue,
			"refunded_count":  refundedCount,
			"currency":        "BDT",
		},
		"recent_gateway_events": events,
	})
}

func (h *FinanceHandler) ProcessRefund(c *gin.Context) {
	var req ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	financeUserID := c.GetString("user_id")

	var p models.Payment
	err := h.db.QueryRow(`
		SELECT id, booking_id, amount, status FROM payments WHERE id = $1
	`, req.PaymentID).Scan(&p.ID, &p.BookingID, &p.Amount, &p.Status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment record not found"})
		return
	}

	if p.Status != models.PaymentStatusPaid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only paid transactions can be refunded"})
		return
	}

	if req.Amount > p.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refund amount cannot exceed original payment amount"})
		return
	}

	refundID := fmt.Sprintf("ref-%d", time.Now().UnixNano())
	refundTxRef := fmt.Sprintf("GW-REF-%d", time.Now().UnixNano()%100000000)

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Insert refund record
	_, err = tx.Exec(`
		INSERT INTO refunds (id, payment_id, booking_id, amount, reason, status, processed_by_user_id, gateway_refund_ref, created_at)
		VALUES ($1, $2, $3, $4, $5, 'PROCESSED', $6, $7, NOW())
	`, refundID, p.ID, p.BookingID, req.Amount, req.Reason, financeUserID, refundTxRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refund record"})
		return
	}

	// Update payment status
	newPayStatus := models.PaymentStatusRefunded
	if req.Amount < p.Amount {
		newPayStatus = models.PaymentStatusPartiallyRefunded
	}
	_, _ = tx.Exec("UPDATE payments SET status = $1 WHERE id = $2", newPayStatus, p.ID)

	// Update booking status
	_, _ = tx.Exec("UPDATE bookings SET status = 'Refunded', updated_at = NOW() WHERE id = $1", p.BookingID)

	// Audit Log
	_, _ = tx.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'REFUND_PAYMENT', 'Payment', $3, $4, $5, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), financeUserID, p.ID, c.ClientIP(), fmt.Sprintf(`{"refund_id": "%s", "amount": %.2f, "reason": "%s"}`, refundID, req.Amount, req.Reason))

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Refund issued successfully",
		"refund_id":          refundID,
		"gateway_refund_ref": refundTxRef,
		"status":             "PROCESSED",
	})
}
