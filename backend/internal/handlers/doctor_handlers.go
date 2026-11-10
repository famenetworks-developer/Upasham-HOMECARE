package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"upasham-backend/internal/models"
)

type DoctorHandler struct {
	db *sql.DB
}

func NewDoctorHandler(db *sql.DB) *DoctorHandler {
	return &DoctorHandler{db: db}
}

func (h *DoctorHandler) PublicList(c *gin.Context) {
	specialty := c.Query("specialty")
	query := c.Query("query")
	language := c.Query("language")

	sqlQuery := `
		SELECT id, full_name, bmdc_number, specialties, qualifications, experience_years, workplace,
		       languages, bio, bio_bn, consultation_fee, is_verified, is_published, avatar_url,
		       rating, total_reviews, availability, created_at
		FROM doctor_profiles
		WHERE is_published = TRUE
	`
	var args []interface{}
	argIdx := 1

	if specialty != "" {
		sqlQuery += ` AND $` + string(rune('0'+argIdx)) + ` = ANY(specialties)`
		args = append(args, specialty)
		argIdx++
	}

	if query != "" {
		sqlQuery += ` AND (full_name ILIKE $` + string(rune('0'+argIdx)) + ` OR qualifications ILIKE $` + string(rune('0'+argIdx)) + ` OR workplace ILIKE $` + string(rune('0'+argIdx)) + `)`
		args = append(args, "%"+query+"%")
		argIdx++
	}

	if language != "" {
		sqlQuery += ` AND $` + string(rune('0'+argIdx)) + ` = ANY(languages)`
		args = append(args, language)
		argIdx++
	}

	sqlQuery += ` ORDER BY rating DESC, total_reviews DESC`

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
		return
	}
	defer rows.Close()

	var doctors []models.DoctorProfile
	for rows.Next() {
		var d models.DoctorProfile
		var bioBn sql.NullString
		var avail sql.NullString
		var specs, langs []string

		if err := rows.Scan(
			&d.ID, &d.FullName, &d.BMDCNumber, pq.Array(&specs), &d.Qualifications, &d.ExperienceYears, &d.Workplace,
			pq.Array(&langs), &d.Bio, &bioBn, &d.ConsultationFee, &d.IsVerified, &d.IsPublished, &d.AvatarURL,
			&d.Rating, &d.TotalReviews, &avail, &d.CreatedAt,
		); err == nil {
			d.Specialties = specs
			d.Languages = langs
			if bioBn.Valid {
				d.BioBn = bioBn.String
			}
			if avail.Valid {
				d.Availability = avail.String
			}
			doctors = append(doctors, d)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"doctors": doctors,
		"total":   len(doctors),
		"meta": gin.H{
			"title":       "Verified Doctors in Dhaka | Upasham Plus Connected Care",
			"description": "Book video consultations with BMDC-verified specialist physicians across Dhaka from BDT 20.",
		},
	})
}

func (h *DoctorHandler) PublicDetail(c *gin.Context) {
	id := c.Param("id")

	var d models.DoctorProfile
	var bioBn, avail sql.NullString
	var specs, langs []string

	err := h.db.QueryRow(`
		SELECT id, full_name, bmdc_number, specialties, qualifications, experience_years, workplace,
		       languages, bio, bio_bn, consultation_fee, is_verified, is_published, avatar_url,
		       rating, total_reviews, availability, created_at
		FROM doctor_profiles
		WHERE id = $1 AND is_published = TRUE
	`, id).Scan(
		&d.ID, &d.FullName, &d.BMDCNumber, pq.Array(&specs), &d.Qualifications, &d.ExperienceYears, &d.Workplace,
		pq.Array(&langs), &d.Bio, &bioBn, &d.ConsultationFee, &d.IsVerified, &d.IsPublished, &d.AvatarURL,
		&d.Rating, &d.TotalReviews, &avail, &d.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor profile not found"})
		return
	}

	d.Specialties = specs
	d.Languages = langs
	if bioBn.Valid {
		d.BioBn = bioBn.String
	}
	if avail.Valid {
		d.Availability = avail.String
	}

	c.JSON(http.StatusOK, d)
}

func (h *DoctorHandler) GetMyAssignments(c *gin.Context) {
	doctorID := c.GetString("doctor_id")
	if doctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor profile not linked"})
		return
	}

	rows, err := h.db.Query(`
		SELECT b.id, b.booking_ref, b.patient_account_id, b.patient_profile_id, b.doctor_id, b.specialty_id,
		       b.service_type, b.status, b.chief_complaint, b.triage_notes, b.preferred_time, b.scheduled_at,
		       b.quoted_amount, b.currency, b.video_room_id, b.created_at, b.updated_at,
		       pp.full_name, pp.relationship, pp.age, pp.gender, pp.blood_group, pp.known_allergies, pp.chronic_conditions
		FROM bookings b
		JOIN patient_profiles pp ON b.patient_profile_id = pp.id
		WHERE b.doctor_id = $1
		ORDER BY CASE WHEN b.status = 'In Progress' THEN 0 WHEN b.status = 'Confirmed' THEN 1 WHEN b.status = 'Assigned' THEN 2 ELSE 3 END, b.created_at DESC
	`, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}
	defer rows.Close()

	var assignments []gin.H
	for rows.Next() {
		var b models.Booking
		var docID, specID, schedAt, vroomID sql.NullString
		var patName, rel, gender, blood, allergies, conditions string
		var age int

		if err := rows.Scan(
			&b.ID, &b.BookingRef, &b.PatientAccountID, &b.PatientProfileID, &docID, &specID,
			&b.ServiceType, &b.Status, &b.ChiefComplaint, &b.TriageNotes, &b.PreferredTime, &schedAt,
			&b.QuotedAmount, &b.Currency, &vroomID, &b.CreatedAt, &b.UpdatedAt,
			&patName, &rel, &age, &gender, &blood, &allergies, &conditions,
		); err == nil {
			if schedAt.Valid {
				b.ScheduledAt = &schedAt.String
			}
			if vroomID.Valid {
				b.VideoRoomID = &vroomID.String
			}

			assignments = append(assignments, gin.H{
				"booking": b,
				"patient": gin.H{
					"name":               patName,
					"relationship":       rel,
					"age":                age,
					"gender":             gender,
					"blood_group":        blood,
					"known_allergies":    allergies,
					"chronic_conditions": conditions,
				},
			})
		}
	}

	c.JSON(http.StatusOK, assignments)
}
