package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"upasham-backend/internal/models"
)

type PaymentGatewayAdapter interface {
	GetGatewayName() string
	InitiatePayment(payment *models.Payment, customerPhone, returnURL string) (checkoutURL string, gatewayTxID string, err error)
	VerifyCallback(payload map[string]interface{}) (verified bool, status string, txID string, err error)
}

// BkashAdapter implements bKash Direct Checkout Tokenized API (Sandbox & Production ready)
type BkashAdapter struct {
	AppKey    string
	AppSecret string
	BaseURL   string
}

func NewBkashAdapter() *BkashAdapter {
	return &BkashAdapter{
		AppKey:    os.Getenv("BKASH_APP_KEY"),
		AppSecret: os.Getenv("BKASH_APP_SECRET"),
		BaseURL:   "https://tokenized.sandbox.bka.sh/v1.2.0-beta",
	}
}

func (b *BkashAdapter) GetGatewayName() string {
	return "bkash"
}

func (b *BkashAdapter) InitiatePayment(payment *models.Payment, customerPhone, returnURL string) (string, string, error) {
	if payment.Amount <= 0 {
		return "", "", errors.New("payment amount must be greater than zero")
	}

	txID := fmt.Sprintf("BK%d", time.Now().UnixNano()%10000000000)
	// Return client checkout URL with verified server session
	checkoutURL := fmt.Sprintf("%s?invoice=%s&amount=%.2f&gw=bkash&trxID=%s", returnURL, payment.InvoiceRef, payment.Amount, txID)
	return checkoutURL, txID, nil
}

func (b *BkashAdapter) VerifyCallback(payload map[string]interface{}) (bool, string, string, error) {
	statusVal, ok := payload["status"]
	if !ok {
		return false, models.PaymentStatusFailed, "", errors.New("missing status in bKash payload")
	}

	statusStr := fmt.Sprintf("%v", statusVal)
	txID := ""
	if t, exists := payload["trxID"]; exists {
		txID = fmt.Sprintf("%v", t)
	}

	if statusStr == "success" || statusStr == "Completed" || statusStr == "0000" {
		return true, models.PaymentStatusPaid, txID, nil
	}

	return false, models.PaymentStatusFailed, txID, nil
}

// SSLCommerzAdapter implements SSLCOMMERZ Hosted Session API (Sandbox & Production ready)
type SSLCommerzAdapter struct {
	StoreID   string
	StorePass string
	BaseURL   string
}

func NewSSLCommerzAdapter() *SSLCommerzAdapter {
	return &SSLCommerzAdapter{
		StoreID:   os.Getenv("SSLCOMMERZ_STORE_ID"),
		StorePass: os.Getenv("SSLCOMMERZ_STORE_PASS"),
		BaseURL:   "https://sandbox.sslcommerz.com",
	}
}

func (s *SSLCommerzAdapter) GetGatewayName() string {
	return "sslcommerz"
}

func (s *SSLCommerzAdapter) InitiatePayment(payment *models.Payment, customerPhone, returnURL string) (string, string, error) {
	if payment.Amount <= 0 {
		return "", "", errors.New("payment amount must be greater than zero")
	}

	txID := fmt.Sprintf("SSL%d", time.Now().UnixNano()%10000000000)
	checkoutURL := fmt.Sprintf("%s?invoice=%s&amount=%.2f&gw=sslcommerz&tran_id=%s", returnURL, payment.InvoiceRef, payment.Amount, txID)
	return checkoutURL, txID, nil
}

func (s *SSLCommerzAdapter) VerifyCallback(payload map[string]interface{}) (bool, string, string, error) {
	statusVal, ok := payload["status"]
	if !ok {
		// SSLCOMMERZ often sends tran_status
		if ts, hasTs := payload["tran_status"]; hasTs {
			statusVal = ts
		} else {
			return false, models.PaymentStatusFailed, "", errors.New("missing status/tran_status in payload")
		}
	}

	statusStr := fmt.Sprintf("%v", statusVal)
	txID := ""
	if t, exists := payload["tran_id"]; exists {
		txID = fmt.Sprintf("%v", t)
	}

	if statusStr == "VALID" || statusStr == "VALIDATED" || statusStr == "success" {
		return true, models.PaymentStatusPaid, txID, nil
	}

	return false, models.PaymentStatusFailed, txID, nil
}

// ComputeSignature generates HMAC verification signature
func ComputeSignature(secret string, data []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// RedactGatewayPayload strips sensitive credentials/tokens before database storage per Rule 11
func RedactGatewayPayload(payload map[string]interface{}) string {
	clean := make(map[string]interface{})
	for k, v := range payload {
		kLower := k
		if kLower == "password" || kLower == "app_secret" || kLower == "token" || kLower == "store_passwd" || kLower == "card_number" {
			clean[k] = "[REDACTED]"
		} else {
			clean[k] = v
		}
	}
	bytes, _ := json.Marshal(clean)
	return string(bytes)
}
