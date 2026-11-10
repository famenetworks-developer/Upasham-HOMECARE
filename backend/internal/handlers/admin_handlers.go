package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"upasham-backend/internal/models"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	var totalUsers, totalDoctors, totalBookings int
	var totalRevenue float64

	_ = h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM doctor_profiles").Scan(&totalDoctors)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM bookings").Scan(&totalBookings)
	_ = h.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'Paid'").Scan(&totalRevenue)

	c.JSON(http.StatusOK, gin.H{
		"total_users":    totalUsers,
		"total_doctors":  totalDoctors,
		"total_bookings": totalBookings,
		"total_revenue":  totalRevenue,
		"currency":       "BDT",
	})
}

func (h *AdminHandler) ListAllDoctors(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT dp.id, dp.full_name, dp.bmdc_number, dp.specialties, dp.qualifications, dp.experience_years, dp.workplace,
		       dp.consultation_fee, dp.is_verified, dp.is_published, dp.rating, dp.total_reviews, u.email, u.phone
		FROM doctor_profiles dp
		JOIN users u ON dp.user_id = u.id
		ORDER BY dp.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors: " + err.Error()})
		return
	}
	defer rows.Close()

	var doctors []gin.H
	for rows.Next() {
		var d models.DoctorProfile
		var specs []string
		var email, phone string
		if err := rows.Scan(
			&d.ID, &d.FullName, &d.BMDCNumber, pq.Array(&specs), &d.Qualifications, &d.ExperienceYears, &d.Workplace,
			&d.ConsultationFee, &d.IsVerified, &d.IsPublished, &d.Rating, &d.TotalReviews, &email, &phone,
		); err == nil {
			doctors = append(doctors, gin.H{
				"id":               d.ID,
				"full_name":        d.FullName,
				"bmdc_number":      d.BMDCNumber,
				"specialties":      specs,
				"qualifications":   d.Qualifications,
				"experience_years": d.ExperienceYears,
				"workplace":        d.Workplace,
				"consultation_fee": d.ConsultationFee,
				"is_verified":      d.IsVerified,
				"is_published":     d.IsPublished,
				"rating":           d.Rating,
				"total_reviews":    d.TotalReviews,
				"email":            email,
				"phone":            phone,
			})
		}
	}

	c.JSON(http.StatusOK, doctors)
}

func (h *AdminHandler) ToggleVerifyDoctor(c *gin.Context) {
	doctorID := c.Param("id")
	adminUserID := c.GetString("user_id")

	var current bool
	err := h.db.QueryRow("SELECT is_verified FROM doctor_profiles WHERE id = $1", doctorID).Scan(&current)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	newVal := !current
	_, _ = h.db.Exec("UPDATE doctor_profiles SET is_verified = $1 WHERE id = $2", newVal, doctorID)

	// Audit Log
	_, _ = h.db.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'TOGGLE_VERIFY_DOCTOR', 'DoctorProfile', $3, $4, $5, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), adminUserID, doctorID, c.ClientIP(), fmt.Sprintf(`{"verified": %t}`, newVal))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Doctor verification updated",
		"doctor_id":   doctorID,
		"is_verified": newVal,
	})
}

func (h *AdminHandler) TogglePublishDoctor(c *gin.Context) {
	doctorID := c.Param("id")
	adminUserID := c.GetString("user_id")

	var current bool
	err := h.db.QueryRow("SELECT is_published FROM doctor_profiles WHERE id = $1", doctorID).Scan(&current)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	newVal := !current
	_, _ = h.db.Exec("UPDATE doctor_profiles SET is_published = $1 WHERE id = $2", newVal, doctorID)

	// Audit Log
	_, _ = h.db.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at)
		VALUES ($1, $2, 'TOGGLE_PUBLISH_DOCTOR', 'DoctorProfile', $3, $4, $5, NOW())
	`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), adminUserID, doctorID, c.ClientIP(), fmt.Sprintf(`{"published": %t}`, newVal))

	c.JSON(http.StatusOK, gin.H{
		"message":      "Doctor publication status updated",
		"doctor_id":    doctorID,
		"is_published": newVal,
	})
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, user_id, action, entity_type, entity_id, ip_address, details_json, created_at
		FROM audit_events
		ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}
	defer rows.Close()

	var logs []models.AuditEvent
	for rows.Next() {
		var a models.AuditEvent
		var uid, entID, ip sql.NullString
		if err := rows.Scan(&a.ID, &uid, &a.Action, &a.EntityType, &entID, &ip, &a.DetailsJSON, &a.CreatedAt); err == nil {
			if uid.Valid {
				a.UserID = uid.String
			}
			if entID.Valid {
				a.EntityID = entID.String
			}
			if ip.Valid {
				a.IPAddress = ip.String
			}
			logs = append(logs, a)
		}
	}

	c.JSON(http.StatusOK, logs)
}

func (h *AdminHandler) GetCatalogue(c *gin.Context) {
	specRows, _ := h.db.Query("SELECT id, name, name_bn, slug, icon, description, is_active FROM specialties ORDER BY name ASC")
	var specialties []models.Specialty
	if specRows != nil {
		defer specRows.Close()
		for specRows.Next() {
			var s models.Specialty
			_ = specRows.Scan(&s.ID, &s.Name, &s.NameBn, &s.Slug, &s.Icon, &s.Description, &s.IsActive)
			specialties = append(specialties, s)
		}
	}

	srvRows, _ := h.db.Query("SELECT id, name, slug, fulfillment_type, description, base_price, is_active FROM services ORDER BY base_price ASC")
	var services []models.Service
	if srvRows != nil {
		defer srvRows.Close()
		for srvRows.Next() {
			var s models.Service
			_ = srvRows.Scan(&s.ID, &s.Name, &s.Slug, &s.FulfillmentType, &s.Description, &s.BasePrice, &s.IsActive)
			services = append(services, s)
		}
	}

	zoneRows, _ := h.db.Query("SELECT id, name, code, city, is_active FROM service_zones")
	var zones []models.ServiceZone
	if zoneRows != nil {
		defer zoneRows.Close()
		for zoneRows.Next() {
			var z models.ServiceZone
			_ = zoneRows.Scan(&z.ID, &z.Name, &z.Code, &z.City, &z.IsActive)
			zones = append(zones, z)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"specialties": specialties,
		"services":    services,
		"zones":       zones,
	})
}
