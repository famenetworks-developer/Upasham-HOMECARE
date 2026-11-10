package models

import (
	"time"
)

// System Roles
const (
	RolePatient    = "patient"
	RoleDoctor     = "doctor"
	RoleOperations = "operations"
	RoleFinance    = "finance"
	RoleAdmin      = "admin"
)

// Booking States per Section 6 State Machine:
// Booking: Draft -> Payment Pending -> Paid/Awaiting Assignment -> Assigned -> Confirmed -> In Progress -> Completed.
// Controlled alternate transitions: Cancelled, No-show, Refund Pending, Refunded.
const (
	BookingStatusDraft                  = "Draft"
	BookingStatusPaymentPending         = "Payment Pending"
	BookingStatusPaidAwaitingAssignment = "Paid/Awaiting Assignment"
	BookingStatusAssigned               = "Assigned"
	BookingStatusConfirmed              = "Confirmed"
	BookingStatusInProgress             = "In Progress"
	BookingStatusCompleted              = "Completed"
	BookingStatusCancelled              = "Cancelled"
	BookingStatusNoShow                 = "No-show"
	BookingStatusRefundPending          = "Refund Pending"
	BookingStatusRefunded               = "Refunded"
)

// Payment States per Section 6 State Machine:
// Payment: Initiated -> Processing -> Paid.
// Alternate states: Failed, Expired, Refund Requested, Partially Refunded, Refunded, Reconciliation Exception.
const (
	PaymentStatusInitiated               = "Initiated"
	PaymentStatusProcessing              = "Processing"
	PaymentStatusPaid                    = "Paid"
	PaymentStatusFailed                  = "Failed"
	PaymentStatusExpired                 = "Expired"
	PaymentStatusRefundRequested         = "Refund Requested"
	PaymentStatusPartiallyRefunded       = "Partially Refunded"
	PaymentStatusRefunded                = "Refunded"
	PaymentStatusReconciliationException = "Reconciliation Exception"
)

// User represents authenticated account
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"` // patient, doctor, operations, finance, admin
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type PatientAccount struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	EmergencyContactPhone string    `json:"emergency_contact_phone"`
	AddressDhaka          string    `json:"address_dhaka"`
	CreatedAt             time.Time `json:"created_at"`
}

type PatientProfile struct {
	ID                string    `json:"id"`
	AccountID         string    `json:"account_id"`
	FullName          string    `json:"full_name"`
	Relationship      string    `json:"relationship"` // self, child, spouse, parent, other
	Gender            string    `json:"gender"`       // male, female, other
	DateOfBirth       string    `json:"date_of_birth,omitempty"`
	Age               int       `json:"age"`
	BloodGroup        string    `json:"blood_group"`
	KnownAllergies    string    `json:"known_allergies"`
	ChronicConditions string    `json:"chronic_conditions"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

type DoctorProfile struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	FullName        string    `json:"full_name"`
	BMDCNumber      string    `json:"bmdc_number"`
	Specialties     []string  `json:"specialties"`
	Qualifications  string    `json:"qualifications"`
	ExperienceYears int       `json:"experience_years"`
	Workplace       string    `json:"workplace"`
	Languages       []string  `json:"languages"`
	Bio             string    `json:"bio"`
	BioBn           string    `json:"bio_bn,omitempty"`
	ConsultationFee float64   `json:"consultation_fee"`
	IsVerified      bool      `json:"is_verified"`
	IsPublished     bool      `json:"is_published"`
	AvatarURL       string    `json:"avatar_url"`
	Rating          float64   `json:"rating"`
	TotalReviews    int       `json:"total_reviews"`
	Availability    string    `json:"availability"` // e.g. "Available today 6 PM - 10 PM"
	CreatedAt       time.Time `json:"created_at"`
}

type Specialty struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	NameBn        string `json:"name_bn"`
	Slug          string `json:"slug"`
	Icon          string `json:"icon"`
	Description   string `json:"description"`
	DescriptionBn string `json:"description_bn"`
	IsActive      bool   `json:"is_active"`
}

type Credential struct {
	ID             string    `json:"id"`
	DoctorID       string    `json:"doctor_id"`
	CredentialType string    `json:"credential_type"` // BMDC_REGISTRATION, MBBS_CERTIFICATE, SPECIALIZATION
	DocumentURL    string    `json:"document_url"`
	IssueDate      string    `json:"issue_date"`
	ExpiryDate     string    `json:"expiry_date"`
	IsVerified     bool      `json:"is_verified"`
	CreatedAt      time.Time `json:"created_at"`
}

type Service struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	FulfillmentType string  `json:"fulfillment_type"` // digital_consultation, internal_test, partner_test, sample_collection, nurse_visit, home_doctor_visit
	Description     string  `json:"description"`
	BasePrice       float64 `json:"base_price"`
	IsActive        bool    `json:"is_active"`
}

type ServicePrice struct {
	ID        string  `json:"id"`
	ServiceID string  `json:"service_id"`
	ZoneID    string  `json:"zone_id"`
	BasePrice float64 `json:"base_price"`
	Currency  string  `json:"currency"`
}

type ServiceZone struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	City     string `json:"city"` // Dhaka
	IsActive bool   `json:"is_active"`
}

type Booking struct {
	ID               string    `json:"id"`
	BookingRef       string    `json:"booking_ref"`
	PatientAccountID string    `json:"patient_account_id"`
	PatientProfileID string    `json:"patient_profile_id"`
	DoctorID         *string   `json:"doctor_id,omitempty"` // nullable for general triage
	SpecialtyID      *string   `json:"specialty_id,omitempty"`
	ServiceType      string    `json:"service_type"` // general (BDT 20), specialist (BDT 1,000+)
	Status           string    `json:"status"`
	ChiefComplaint   string    `json:"chief_complaint"`
	TriageNotes      string    `json:"triage_notes"`
	PreferredTime    string    `json:"preferred_time"`
	ScheduledAt      *string   `json:"scheduled_at,omitempty"`
	QuotedAmount     float64   `json:"quoted_amount"`
	Currency         string    `json:"currency"`
	PaymentID        *string   `json:"payment_id,omitempty"`
	VideoRoomID      *string   `json:"video_room_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Joins for response convenience
	PatientName    string `json:"patient_name,omitempty"`
	DoctorName     *string `json:"doctor_name,omitempty"`
	SpecialtyName  *string `json:"specialty_name,omitempty"`
	Relationship   string `json:"relationship,omitempty"`
	DoctorWorkplace *string `json:"doctor_workplace,omitempty"`
}

type BookingStatusHistory struct {
	ID              string    `json:"id"`
	BookingID       string    `json:"booking_id"`
	FromStatus      string    `json:"from_status"`
	ToStatus        string    `json:"to_status"`
	ChangedByUserID string    `json:"changed_by_user_id"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}

type Assignment struct {
	ID               string    `json:"id"`
	BookingID        string    `json:"booking_id"`
	DoctorID         string    `json:"doctor_id"`
	AssignedByUserID string    `json:"assigned_by_user_id"`
	Status           string    `json:"status"` // PENDING, ACCEPTED, DECLINED
	AssignedAt       time.Time `json:"assigned_at"`
}

type ScheduleEvent struct {
	ID        string    `json:"id"`
	BookingID string    `json:"booking_id"`
	EventType string    `json:"event_type"` // APPOINTMENT_CONFIRMED, RESCHEDULED, CALL_STARTED, CALL_ENDED
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type Payment struct {
	ID               string     `json:"id"`
	BookingID        string     `json:"booking_id"`
	InvoiceRef       string     `json:"invoice_ref"`
	Gateway          string     `json:"gateway"` // bkash, sslcommerz
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	GatewayTxID      string     `json:"gateway_tx_id"`
	GatewaySignature string     `json:"gateway_signature"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`

	// Enriched info
	PatientName string  `json:"patient_name,omitempty"`
	ServiceType string  `json:"service_type,omitempty"`
	BookingRef  string  `json:"booking_ref,omitempty"`
}

type GatewayEvent struct {
	ID          string    `json:"id"`
	PaymentID   string    `json:"payment_id"`
	Gateway     string    `json:"gateway"`
	EventType   string    `json:"event_type"`
	PayloadJSON string    `json:"payload_json"`
	Signature   string    `json:"signature"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
}

type Refund struct {
	ID                string    `json:"id"`
	PaymentID         string    `json:"payment_id"`
	BookingID         string    `json:"booking_id"`
	Amount            float64   `json:"amount"`
	Reason            string    `json:"reason"`
	Status            string    `json:"status"` // REQUESTED, APPROVED, PROCESSED, REJECTED
	ProcessedByUserID string    `json:"processed_by_user_id"`
	GatewayRefundRef  string    `json:"gateway_refund_ref"`
	CreatedAt         time.Time `json:"created_at"`
}

type VideoRoom struct {
	ID               string    `json:"id"`
	BookingID        string    `json:"booking_id"`
	RoomName         string    `json:"room_name"`
	RoomPIN          string    `json:"room_pin"`
	HostToken        string    `json:"host_token"`
	ParticipantToken string    `json:"participant_token"`
	JitsiDomain      string    `json:"jitsi_domain"`
	ExpiresAt        time.Time `json:"expires_at"`
	Status           string    `json:"status"` // ACTIVE, EXPIRED, CLOSED
	CreatedAt        time.Time `json:"created_at"`
}

type AttendanceEvent struct {
	ID          string    `json:"id"`
	VideoRoomID string    `json:"video_room_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	EventType   string    `json:"event_type"` // join, leave
	Timestamp   time.Time `json:"timestamp"`
}

type Encounter struct {
	ID               string     `json:"id"`
	BookingID        string     `json:"booking_id"`
	DoctorID         string     `json:"doctor_id"`
	PatientProfileID string     `json:"patient_profile_id"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	Status           string     `json:"status"` // IN_PROGRESS, COMPLETED
}

type ClinicalNote struct {
	ID          string    `json:"id"`
	EncounterID string    `json:"encounter_id"`
	Subjective  string    `json:"subjective"`
	Objective   string    `json:"objective"`
	Assessment  string    `json:"assessment"`
	Plan        string    `json:"plan"`
	IsPrivate   bool      `json:"is_private"` // true = internal notes, false = patient-visible
	CreatedAt   time.Time `json:"created_at"`
}

type Prescription struct {
	ID               string    `json:"id"`
	BookingID        string    `json:"booking_id"`
	EncounterID      string    `json:"encounter_id"`
	PrescriptionRef  string    `json:"prescription_ref"`
	DoctorID         string    `json:"doctor_id"`
	PatientProfileID string    `json:"patient_profile_id"`
	Diagnosis        string    `json:"diagnosis"`
	Advice           string    `json:"advice"`
	FollowUpDate     string    `json:"follow_up_date"`
	Warnings         string    `json:"warnings"`
	VerificationHash string    `json:"verification_hash"`
	Version          int       `json:"version"`
	CreatedAt        time.Time `json:"created_at"`

	DoctorName       string         `json:"doctor_name,omitempty"`
	DoctorBMDC       string         `json:"doctor_bmdc,omitempty"`
	DoctorDegrees    string         `json:"doctor_degrees,omitempty"`
	DoctorWorkplace  string         `json:"doctor_workplace,omitempty"`
	PatientName      string         `json:"patient_name,omitempty"`
	PatientAge       int            `json:"patient_age,omitempty"`
	PatientGender    string         `json:"patient_gender,omitempty"`
	BookingRef       string         `json:"booking_ref,omitempty"`
	Medicines        []MedicineLine `json:"medicines,omitempty"`
	Tests            []TestOrder    `json:"tests,omitempty"`
}

type PrescriptionVersion struct {
	ID             string    `json:"id"`
	PrescriptionID string    `json:"prescription_id"`
	VersionNum     int       `json:"version_num"`
	SnapshotJSON   string    `json:"snapshot_json"`
	CreatedAt      time.Time `json:"created_at"`
}

type MedicineLine struct {
	ID             string `json:"id"`
	PrescriptionID string `json:"prescription_id"`
	BrandName      string `json:"brand_name"`
	GenericName    string `json:"generic_name"`
	Strength       string `json:"strength"`      // e.g. 500mg, 20mg
	DosageForm     string `json:"dosage_form"`   // Tablet, Capsule, Syrup, Drop
	Dose           string `json:"dose"`          // e.g. 1+0+1, 1 tablet
	Route          string `json:"route"`         // Oral, Topical, IV
	Frequency      string `json:"frequency"`     // Twice daily, After meals
	Duration       string `json:"duration"`      // 5 days, 1 month
	Instructions   string `json:"instructions"`  // Take after meal with water
}

type TestOrder struct {
	ID             string `json:"id"`
	PrescriptionID string `json:"prescription_id"`
	TestName       string `json:"test_name"`
	Instructions   string `json:"instructions"`
	Status         string `json:"status"` // PENDING, SAMPLE_REQUESTED, COMPLETED
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Link      string    `json:"link"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type Consent struct {
	ID               string    `json:"id"`
	BookingID        string    `json:"booking_id"`
	PatientAccountID string    `json:"patient_account_id"`
	ConsentType      string    `json:"consent_type"` // TELEHEALTH_CONSENT, DATA_PROCESSING_CONSENT
	AcceptedAt       time.Time `json:"accepted_at"`
}

type AuditEvent struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Action      string    `json:"action"` // LOGIN, VIEW_PATIENT, ASSIGN_DOCTOR, CONFIRM_BOOKING, ISSUE_PRESCRIPTION, REFUND_PAYMENT, VERIFY_DOCTOR
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	IPAddress   string    `json:"ip_address"`
	DetailsJSON string    `json:"details_json"`
	CreatedAt   time.Time `json:"created_at"`
}

// Dormant / Configurable domain support as specified in Section 5
type LabPartner struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ContactPhone string    `json:"contact_phone"`
	Address      string    `json:"address"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type SampleJob struct {
	ID              string     `json:"id"`
	BookingID       string     `json:"booking_id"`
	PatientID       string     `json:"patient_id"`
	LabPartnerID    *string    `json:"lab_partner_id,omitempty"`
	SampleType      string     `json:"sample_type"` // Blood, Urine, Swab
	Status          string     `json:"status"`      // PENDING, COLLECTOR_ASSIGNED, COLLECTED, DELIVERED_LAB, RESULT_READY
	ScheduledPickup *time.Time `json:"scheduled_pickup,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DiagnosticReport struct {
	ID          string    `json:"id"`
	SampleJobID *string   `json:"sample_job_id,omitempty"`
	PatientID   string    `json:"patient_id"`
	ReportTitle string    `json:"report_title"`
	DocumentURL string    `json:"document_url"`
	ReportJSON  string    `json:"report_json"`
	VerifiedBy  string    `json:"verified_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type HomeServiceJob struct {
	ID             string     `json:"id"`
	ServiceType    string     `json:"service_type"` // NURSE_VISIT, CAREGIVER_SHIFT, HOME_DOCTOR
	PatientID      string     `json:"patient_id"`
	DhakaZone      string     `json:"dhaka_zone"`
	AssignedStaff  *string    `json:"assigned_staff,omitempty"`
	Status         string     `json:"status"`
	ScheduledTime  *time.Time `json:"scheduled_time,omitempty"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
}

type ProviderAssignment struct {
	ID           string    `json:"id"`
	JobID        string    `json:"job_id"`
	ProviderID   string    `json:"provider_id"`
	ProviderType string    `json:"provider_type"` // DOCTOR, NURSE, LAB_TECHNICIAN, CAREGIVER
	Status       string    `json:"status"`
	AssignedAt   time.Time `json:"assigned_at"`
}

type Package struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type PackageItem struct {
	ID          string `json:"id"`
	PackageID   string `json:"package_id"`
	ServiceID   string `json:"service_id"`
	Quantity    int    `json:"quantity"`
}

type Subscription struct {
	ID               string    `json:"id"`
	PatientAccountID string    `json:"patient_account_id"`
	PackageID        string    `json:"package_id"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Status           string    `json:"status"` // ACTIVE, EXPIRED, CANCELLED
}
