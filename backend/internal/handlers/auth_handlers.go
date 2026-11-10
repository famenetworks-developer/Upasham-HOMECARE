package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"upasham-backend/internal/auth"
	"upasham-backend/internal/models"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type RegisterRequest struct {
	FullName              string `json:"full_name" binding:"required"`
	Email                 string `json:"email" binding:"required,email"`
	Phone                 string `json:"phone" binding:"required"`
	Password              string `json:"password" binding:"required,min=6"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	AddressDhaka          string `json:"address_dhaka"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type DemoLoginRequest struct {
	Role string `json:"role" binding:"required"` // patient, doctor_jony, doctor_shuvojit, operations, finance, admin
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check existing email
	var count int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists"})
		return
	}

	passHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	userID := "usr-" + uuid.New().String()[:8]
	accountID := "pacc-" + uuid.New().String()[:8]
	profileID := "pprof-" + uuid.New().String()[:8]

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}
	defer tx.Rollback()

	// Insert User
	_, err = tx.Exec(`
		INSERT INTO users (id, email, phone, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'patient', true, NOW(), NOW())
	`, userID, req.Email, req.Phone, passHash, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Insert Patient Account
	_, err = tx.Exec(`
		INSERT INTO patient_accounts (id, user_id, emergency_contact_phone, address_dhaka, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, accountID, userID, req.EmergencyContactPhone, req.AddressDhaka)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient account"})
		return
	}

	// Insert Primary Self Profile
	_, err = tx.Exec(`
		INSERT INTO patient_profiles (id, account_id, full_name, relationship, gender, age, blood_group, known_allergies, chronic_conditions, notes, created_at)
		VALUES ($1, $2, $3, 'self', 'other', 30, 'Unknown', 'None reported', 'None reported', 'Primary account holder', NOW())
	`, profileID, accountID, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create primary profile"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error"})
		return
	}

	token, _ := auth.GenerateToken(userID, req.Email, "patient", req.FullName, accountID, "")

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":                 userID,
			"email":              req.Email,
			"phone":              req.Phone,
			"full_name":          req.FullName,
			"role":               "patient",
			"patient_account_id": accountID,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var u models.User
	err := h.db.QueryRow(`
		SELECT id, email, phone, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`, req.Email).Scan(&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !u.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is deactivated"})
		return
	}

	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	var patientAccountID, doctorID string
	if u.Role == models.RolePatient {
		_ = h.db.QueryRow("SELECT id FROM patient_accounts WHERE user_id = $1", u.ID).Scan(&patientAccountID)
	} else if u.Role == models.RoleDoctor {
		_ = h.db.QueryRow("SELECT id FROM doctor_profiles WHERE user_id = $1", u.ID).Scan(&doctorID)
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Role, u.FullName, patientAccountID, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation error"})
		return
	}

	// Audit login
	_, _ = h.db.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'LOGIN', 'User', $3, $4, '{"auth_method": "password"}', NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), u.ID, u.ID, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":                 u.ID,
			"email":              u.Email,
			"phone":              u.Phone,
			"full_name":          u.FullName,
			"role":               u.Role,
			"patient_account_id": patientAccountID,
			"doctor_id":          doctorID,
		},
	})
}

// DemoLogin provides seamless 1-click test access across Patient, Doctor, Operations, Finance, and Admin
func (h *AuthHandler) DemoLogin(c *gin.Context) {
	var req DemoLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetEmail := ""
	switch req.Role {
	case "patient":
		targetEmail = "patient@upasham.com"
	case "doctor_jony", "doctor":
		targetEmail = "dr.jony@upasham.com"
	case "doctor_shuvojit":
		targetEmail = "dr.shuvojit@upasham.com"
	case "operations":
		targetEmail = "ops@upasham.com"
	case "finance":
		targetEmail = "finance@upasham.com"
	case "admin":
		targetEmail = "admin@upasham.com"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown demo role"})
		return
	}

	var u models.User
	err := h.db.QueryRow(`
		SELECT id, email, phone, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`, targetEmail).Scan(&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Demo user not found: " + targetEmail})
		return
	}

	var patientAccountID, doctorID string
	if u.Role == models.RolePatient {
		_ = h.db.QueryRow("SELECT id FROM patient_accounts WHERE user_id = $1", u.ID).Scan(&patientAccountID)
	} else if u.Role == models.RoleDoctor {
		_ = h.db.QueryRow("SELECT id FROM doctor_profiles WHERE user_id = $1", u.ID).Scan(&doctorID)
	}

	token, _ := auth.GenerateToken(u.ID, u.Email, u.Role, u.FullName, patientAccountID, doctorID)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":                 u.ID,
			"email":              u.Email,
			"phone":              u.Phone,
			"full_name":          u.FullName,
			"role":               u.Role,
			"patient_account_id": patientAccountID,
			"doctor_id":          doctorID,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	var u models.User
	err := h.db.QueryRow(`
		SELECT id, email, phone, full_name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(&u.ID, &u.Email, &u.Phone, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	patientAccountID := c.GetString("patient_account_id")
	doctorID := c.GetString("doctor_id")

	c.JSON(http.StatusOK, gin.H{
		"id":                 u.ID,
		"email":              u.Email,
		"phone":              u.Phone,
		"full_name":          u.FullName,
		"role":               u.Role,
		"patient_account_id": patientAccountID,
		"doctor_id":          doctorID,
	})
}
