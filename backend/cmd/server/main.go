package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"upasham-backend/internal/auth"
	"upasham-backend/internal/config"
	"upasham-backend/internal/db"
	"upasham-backend/internal/handlers"
	"upasham-backend/internal/services"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("[SERVER] Starting Upasham Plus Connected Care Backend in %s mode", cfg.Env)

	// Initialize Database
	database, err := db.InitDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("[DB] Fatal error initializing PostgreSQL database: %v", err)
	}
	defer database.Close()

	// Auth secret
	auth.SetJWTSecret(cfg.JWTSecret)

	// Initialize Services
	quotationServ := services.NewQuotationService(database, cfg.GeneralPrice)
	videoServ := services.NewVideoService(database, cfg.JitsiDomain)
	pdfServ := services.NewPDFService()

	// Initialize Handlers
	authH := handlers.NewAuthHandler(database)
	patientH := handlers.NewPatientHandler(database)
	doctorH := handlers.NewDoctorHandler(database)
	bookingH := handlers.NewBookingHandler(database, quotationServ)
	paymentH := handlers.NewPaymentHandler(database)
	opsH := handlers.NewOperationsHandler(database, videoServ)
	clinicalH := handlers.NewClinicalHandler(database, pdfServ)
	videoH := handlers.NewVideoHandler(database, videoServ)
	financeH := handlers.NewFinanceHandler(database)
	adminH := handlers.NewAdminHandler(database)
	contentH := handlers.NewContentHandler(database)
	cmsH := handlers.NewCMSHandler(database)

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(auth.SecurityHeaders())

	// Configure CORS for React frontend
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))
	router.Static("/uploads", "./uploads")

	// Health Check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"platform":  "Upasham Plus Connected Care Dhaka",
			"version":   "1.0.0",
			"database":  "PostgreSQL 17 (Connected)",
			"gateways":  []string{"bKash", "SSLCOMMERZ"},
			"tele_mode": "Jitsi Protected Consultation",
		})
	})

	api := router.Group("/api")

	// ================= PUBLIC ROUTES =================
	authLimiter := auth.NewRateLimiter(10, time.Minute)
	consultLimiter := auth.NewRateLimiter(20, time.Minute)
	paymentLimiter := auth.NewRateLimiter(30, time.Minute)
	api.POST("/auth/register", authLimiter.Middleware("register"), authH.Register)
	api.POST("/auth/login", authLimiter.Middleware("login"), authH.Login)
	api.POST("/auth/demo-login", authH.DemoLogin)

	api.GET("/doctors", doctorH.PublicList)
	api.GET("/doctors/:id", doctorH.PublicDetail)
	api.GET("/catalogue", adminH.GetCatalogue)
	api.GET("/content", contentH.GetPublished)
	api.POST("/quotation", consultLimiter.Middleware("quotation"), bookingH.GetQuotation)

	// Payment Callbacks & Simulation
	api.POST("/payments/callback", paymentLimiter.Middleware("payment_callback"), paymentH.Callback)
	if cfg.Env != "production" {
		api.GET("/payments/simulate", paymentH.SimulatePayment)
	}

	// ================= PROTECTED ROUTES =================
	protected := api.Group("")
	protected.Use(auth.RequireAuth())
	{
		protected.GET("/prescriptions/:id/pdf", clinicalH.DownloadPrescriptionPDF)
		protected.GET("/auth/me", authH.Me)

		// Patient Portal Routes
		patientGroup := protected.Group("/patient")
		patientGroup.Use(auth.RequireRole("patient"))
		{
			patientGroup.GET("/profiles", patientH.GetProfiles)
			patientGroup.POST("/profiles", patientH.CreateProfile)
			patientGroup.GET("/bookings", patientH.GetMyBookings)
			patientGroup.GET("/prescriptions", patientH.GetMyPrescriptions)
			patientGroup.POST("/bookings", consultLimiter.Middleware("create_booking"), bookingH.CreateBooking)
			patientGroup.POST("/payments/initiate", paymentLimiter.Middleware("payment_initiate"), paymentH.InitiatePayment)
		}

		// Generic booking lookup (with object-level security in handler)
		protected.GET("/bookings/:id", bookingH.GetBooking)

		// Doctor Portal Routes
		doctorGroup := protected.Group("/doctor")
		doctorGroup.Use(auth.RequireRole("doctor"))
		{
			doctorGroup.GET("/assignments", doctorH.GetMyAssignments)
			doctorGroup.POST("/clinical/save", clinicalH.SaveEncounterAndPrescription)
		}

		// Video Consultation Routes (Patient & Doctor)
		protected.GET("/video/rooms/:id/access", videoH.GetRoomAccess)
		protected.POST("/video/attendance", videoH.LogAttendance)

		// Operations Portal Routes
		opsGroup := protected.Group("/operations")
		opsGroup.Use(auth.RequireRole("operations", "admin"))
		{
			opsGroup.GET("/queue", opsH.GetPaidQueue)
			opsGroup.POST("/assign", opsH.AssignDoctor)
			opsGroup.POST("/confirm", opsH.ConfirmAppointment)
			opsGroup.POST("/reschedule", opsH.Reschedule)
			opsGroup.POST("/cancel", opsH.Cancel)
		}

		// Finance Portal Routes
		finGroup := protected.Group("/finance")
		finGroup.Use(auth.RequireRole("finance", "admin"))
		{
			finGroup.GET("/dashboard", financeH.GetDashboard)
			finGroup.GET("/transactions", paymentH.ListTransactions)
			finGroup.POST("/refund", financeH.ProcessRefund)
		}

		// Administration Routes
		admGroup := protected.Group("/admin")
		admGroup.Use(auth.RequireRole("admin"))
		{
			admGroup.GET("/stats", adminH.GetStats)
			admGroup.GET("/doctors", adminH.ListAllDoctors)
			admGroup.PATCH("/doctors/:id/verify", adminH.ToggleVerifyDoctor)
			admGroup.PATCH("/doctors/:id/publish", adminH.TogglePublishDoctor)
			admGroup.GET("/audit-logs", adminH.GetAuditLogs)
			admGroup.GET("/content", contentH.GetAdmin)
			admGroup.PUT("/content", contentH.SaveDraft)
			admGroup.POST("/content/publish", contentH.Publish)
			admGroup.GET("/content/revisions", contentH.ListRevisions)
			admGroup.POST("/content/revisions/:id/restore", contentH.RestoreRevision)
			admGroup.GET("/cms/overview", cmsH.Overview)
			admGroup.GET("/cms/entries/:type", cmsH.ListEntries)
			admGroup.POST("/cms/entries/:type", cmsH.CreateEntry)
			admGroup.PUT("/cms/entries/:type/:id", cmsH.UpdateEntry)
			admGroup.DELETE("/cms/entries/:type/:id", cmsH.DeleteEntry)
			admGroup.GET("/cms/media", cmsH.ListMedia)
			admGroup.POST("/cms/media", cmsH.UploadMedia)
			admGroup.DELETE("/cms/media/:id", cmsH.DeleteMedia)
			admGroup.GET("/cms/users", cmsH.ListUsers)
		}
	}

	log.Printf("[SERVER] Listening on http://localhost:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("[SERVER] Failed to run server: %v", err)
	}
}
