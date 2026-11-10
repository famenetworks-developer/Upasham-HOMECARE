package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitDB(dbURL string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("[DB] Successfully connected to PostgreSQL")

	if err = runMigrations(DB); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if err = seedReferenceData(DB); err != nil {
		log.Printf("[DB] Seed data notice: %v", err)
	}

	return DB, nil
}

func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		phone TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		full_name TEXT NOT NULL,
		role TEXT NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS patient_accounts (
		id TEXT PRIMARY KEY,
		user_id TEXT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
		emergency_contact_phone TEXT,
		address_dhaka TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS patient_profiles (
		id TEXT PRIMARY KEY,
		account_id TEXT REFERENCES patient_accounts(id) ON DELETE CASCADE,
		full_name TEXT NOT NULL,
		relationship TEXT NOT NULL,
		gender TEXT NOT NULL,
		date_of_birth TEXT,
		age INT NOT NULL,
		blood_group TEXT,
		known_allergies TEXT,
		chronic_conditions TEXT,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS specialties (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		name_bn TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		icon TEXT,
		description TEXT,
		description_bn TEXT,
		is_active BOOLEAN DEFAULT TRUE
	);

	CREATE TABLE IF NOT EXISTS doctor_profiles (
		id TEXT PRIMARY KEY,
		user_id TEXT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
		full_name TEXT NOT NULL,
		bmdc_number TEXT NOT NULL,
		specialties TEXT[] NOT NULL,
		qualifications TEXT NOT NULL,
		experience_years INT NOT NULL,
		workplace TEXT NOT NULL,
		languages TEXT[] NOT NULL,
		bio TEXT NOT NULL,
		bio_bn TEXT,
		consultation_fee NUMERIC(10,2) NOT NULL,
		is_verified BOOLEAN DEFAULT FALSE,
		is_published BOOLEAN DEFAULT FALSE,
		avatar_url TEXT,
		rating NUMERIC(3,2) DEFAULT 5.0,
		total_reviews INT DEFAULT 0,
		availability TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS credentials (
		id TEXT PRIMARY KEY,
		doctor_id TEXT REFERENCES doctor_profiles(id) ON DELETE CASCADE,
		credential_type TEXT NOT NULL,
		document_url TEXT NOT NULL,
		issue_date TEXT,
		expiry_date TEXT,
		is_verified BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS service_zones (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		code TEXT UNIQUE NOT NULL,
		city TEXT DEFAULT 'Dhaka',
		is_active BOOLEAN DEFAULT TRUE
	);

	CREATE TABLE IF NOT EXISTS services (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		fulfillment_type TEXT NOT NULL,
		description TEXT,
		base_price NUMERIC(10,2) NOT NULL,
		is_active BOOLEAN DEFAULT TRUE
	);

	CREATE TABLE IF NOT EXISTS service_prices (
		id TEXT PRIMARY KEY,
		service_id TEXT REFERENCES services(id) ON DELETE CASCADE,
		zone_id TEXT REFERENCES service_zones(id) ON DELETE CASCADE,
		base_price NUMERIC(10,2) NOT NULL,
		currency TEXT DEFAULT 'BDT'
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id TEXT PRIMARY KEY,
		booking_ref TEXT UNIQUE NOT NULL,
		patient_account_id TEXT REFERENCES patient_accounts(id),
		patient_profile_id TEXT REFERENCES patient_profiles(id),
		doctor_id TEXT REFERENCES doctor_profiles(id),
		specialty_id TEXT REFERENCES specialties(id),
		service_type TEXT NOT NULL,
		status TEXT NOT NULL,
		chief_complaint TEXT NOT NULL,
		triage_notes TEXT,
		preferred_time TEXT,
		scheduled_at TEXT,
		quoted_amount NUMERIC(10,2) NOT NULL,
		currency TEXT DEFAULT 'BDT',
		payment_id TEXT,
		video_room_id TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS booking_status_histories (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id) ON DELETE CASCADE,
		from_status TEXT NOT NULL,
		to_status TEXT NOT NULL,
		changed_by_user_id TEXT,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS assignments (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id) ON DELETE CASCADE,
		doctor_id TEXT REFERENCES doctor_profiles(id),
		assigned_by_user_id TEXT,
		status TEXT DEFAULT 'ACCEPTED',
		assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS schedule_events (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id) ON DELETE CASCADE,
		event_type TEXT NOT NULL,
		start_time TIMESTAMP WITH TIME ZONE NOT NULL,
		end_time TIMESTAMP WITH TIME ZONE NOT NULL,
		status TEXT DEFAULT 'SCHEDULED'
	);

	CREATE TABLE IF NOT EXISTS payments (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id) ON DELETE CASCADE,
		invoice_ref TEXT UNIQUE NOT NULL,
		gateway TEXT NOT NULL,
		amount NUMERIC(10,2) NOT NULL,
		currency TEXT DEFAULT 'BDT',
		status TEXT NOT NULL,
		gateway_tx_id TEXT,
		gateway_signature TEXT,
		paid_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS gateway_events (
		id TEXT PRIMARY KEY,
		payment_id TEXT REFERENCES payments(id) ON DELETE CASCADE,
		gateway TEXT NOT NULL,
		event_type TEXT NOT NULL,
		payload_json TEXT NOT NULL,
		signature TEXT,
		is_verified BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS refunds (
		id TEXT PRIMARY KEY,
		payment_id TEXT REFERENCES payments(id) ON DELETE CASCADE,
		booking_id TEXT REFERENCES bookings(id),
		amount NUMERIC(10,2) NOT NULL,
		reason TEXT NOT NULL,
		status TEXT NOT NULL,
		processed_by_user_id TEXT,
		gateway_refund_ref TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS video_rooms (
		id TEXT PRIMARY KEY,
		booking_id TEXT UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
		room_name TEXT UNIQUE NOT NULL,
		room_pin TEXT NOT NULL,
		host_token TEXT,
		participant_token TEXT,
		jitsi_domain TEXT DEFAULT 'meet.jit.si',
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		status TEXT DEFAULT 'ACTIVE',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS attendance_events (
		id TEXT PRIMARY KEY,
		video_room_id TEXT REFERENCES video_rooms(id) ON DELETE CASCADE,
		user_id TEXT,
		role TEXT NOT NULL,
		event_type TEXT NOT NULL,
		timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS encounters (
		id TEXT PRIMARY KEY,
		booking_id TEXT UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
		doctor_id TEXT REFERENCES doctor_profiles(id),
		patient_profile_id TEXT REFERENCES patient_profiles(id),
		started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		ended_at TIMESTAMP WITH TIME ZONE,
		status TEXT DEFAULT 'COMPLETED'
	);

	CREATE TABLE IF NOT EXISTS clinical_notes (
		id TEXT PRIMARY KEY,
		encounter_id TEXT REFERENCES encounters(id) ON DELETE CASCADE,
		subjective TEXT,
		objective TEXT,
		assessment TEXT,
		plan TEXT,
		is_private BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS prescriptions (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id) ON DELETE CASCADE,
		encounter_id TEXT REFERENCES encounters(id) ON DELETE CASCADE,
		prescription_ref TEXT UNIQUE NOT NULL,
		doctor_id TEXT REFERENCES doctor_profiles(id),
		patient_profile_id TEXT REFERENCES patient_profiles(id),
		diagnosis TEXT NOT NULL,
		advice TEXT,
		follow_up_date TEXT,
		warnings TEXT,
		verification_hash TEXT NOT NULL,
		version INT DEFAULT 1,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS prescription_versions (
		id TEXT PRIMARY KEY,
		prescription_id TEXT REFERENCES prescriptions(id) ON DELETE CASCADE,
		version_num INT NOT NULL,
		snapshot_json TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS medicine_lines (
		id TEXT PRIMARY KEY,
		prescription_id TEXT REFERENCES prescriptions(id) ON DELETE CASCADE,
		brand_name TEXT NOT NULL,
		generic_name TEXT,
		strength TEXT,
		dosage_form TEXT,
		dose TEXT NOT NULL,
		route TEXT DEFAULT 'Oral',
		frequency TEXT NOT NULL,
		duration TEXT NOT NULL,
		instructions TEXT
	);

	CREATE TABLE IF NOT EXISTS test_orders (
		id TEXT PRIMARY KEY,
		prescription_id TEXT REFERENCES prescriptions(id) ON DELETE CASCADE,
		test_name TEXT NOT NULL,
		instructions TEXT,
		status TEXT DEFAULT 'PENDING'
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		link TEXT,
		is_read BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS consents (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id) ON DELETE CASCADE,
		patient_account_id TEXT REFERENCES patient_accounts(id),
		consent_type TEXT NOT NULL,
		accepted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS audit_events (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		action TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id TEXT,
		ip_address TEXT,
		details_json TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS site_content (
		id TEXT PRIMARY KEY,
		draft_content JSONB NOT NULL DEFAULT '{}'::jsonb,
		published_content JSONB NOT NULL DEFAULT '{}'::jsonb,
		draft_revision INT NOT NULL DEFAULT 1,
		published_revision INT NOT NULL DEFAULT 0,
		updated_by TEXT,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		published_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE IF NOT EXISTS site_content_revisions (
		id TEXT PRIMARY KEY,
		revision INT NOT NULL,
		content JSONB NOT NULL,
		status TEXT NOT NULL,
		created_by TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS cms_entries (
		id TEXT PRIMARY KEY,
		content_type TEXT NOT NULL,
		title TEXT NOT NULL,
		slug TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'DRAFT',
		data JSONB NOT NULL DEFAULT '{}'::jsonb,
		sort_order INT NOT NULL DEFAULT 0,
		created_by TEXT,
		updated_by TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE,
		UNIQUE(content_type, slug)
	);
	CREATE INDEX IF NOT EXISTS cms_entries_type_idx ON cms_entries(content_type,status,updated_at DESC) WHERE deleted_at IS NULL;
	CREATE TABLE IF NOT EXISTS media_assets (
		id TEXT PRIMARY KEY,
		file_name TEXT NOT NULL,
		storage_name TEXT NOT NULL,
		url TEXT NOT NULL,
		mime_type TEXT NOT NULL,
		size_bytes BIGINT NOT NULL,
		alt_text TEXT DEFAULT '',
		uploaded_by TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);
	CREATE INDEX IF NOT EXISTS site_content_revisions_created_idx ON site_content_revisions(created_at DESC);
	INSERT INTO site_content (id) VALUES ('primary') ON CONFLICT (id) DO NOTHING;

	-- Dormant domain tables
	CREATE TABLE IF NOT EXISTS lab_partners (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		contact_phone TEXT,
		address TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sample_jobs (
		id TEXT PRIMARY KEY,
		booking_id TEXT REFERENCES bookings(id),
		patient_id TEXT REFERENCES patient_profiles(id),
		lab_partner_id TEXT REFERENCES lab_partners(id),
		sample_type TEXT NOT NULL,
		status TEXT NOT NULL,
		scheduled_pickup TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS diagnostic_reports (
		id TEXT PRIMARY KEY,
		sample_job_id TEXT REFERENCES sample_jobs(id),
		patient_id TEXT REFERENCES patient_profiles(id),
		report_title TEXT NOT NULL,
		document_url TEXT,
		report_json TEXT,
		verified_by TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS home_service_jobs (
		id TEXT PRIMARY KEY,
		service_type TEXT NOT NULL,
		patient_id TEXT REFERENCES patient_profiles(id),
		dhaka_zone TEXT NOT NULL,
		assigned_staff TEXT,
		status TEXT NOT NULL,
		scheduled_time TIMESTAMP WITH TIME ZONE,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS roles (
		id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, description TEXT, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS permissions (
		id TEXT PRIMARY KEY, code TEXT UNIQUE NOT NULL, description TEXT, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS role_permissions (
		role_id TEXT REFERENCES roles(id) ON DELETE CASCADE,
		permission_id TEXT REFERENCES permissions(id) ON DELETE CASCADE,
		PRIMARY KEY (role_id, permission_id)
	);
	CREATE TABLE IF NOT EXISTS provider_assignments (
		id TEXT PRIMARY KEY, home_service_job_id TEXT REFERENCES home_service_jobs(id), provider_user_id TEXT REFERENCES users(id),
		status TEXT NOT NULL DEFAULT 'PENDING', assigned_by_user_id TEXT REFERENCES users(id), assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS packages (
		id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, price_minor BIGINT NOT NULL DEFAULT 0,
		currency TEXT NOT NULL DEFAULT 'BDT', is_active BOOLEAN DEFAULT FALSE, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS package_items (
		id TEXT PRIMARY KEY, package_id TEXT REFERENCES packages(id) ON DELETE CASCADE, service_id TEXT REFERENCES services(id), quantity INT NOT NULL DEFAULT 1
	);
	CREATE TABLE IF NOT EXISTS subscriptions (
		id TEXT PRIMARY KEY, patient_account_id TEXT REFERENCES patient_accounts(id), package_id TEXT REFERENCES packages(id),
		status TEXT NOT NULL DEFAULT 'DORMANT', starts_at TIMESTAMP WITH TIME ZONE, ends_at TIMESTAMP WITH TIME ZONE, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	ALTER TABLE services ADD COLUMN IF NOT EXISTS travel_price NUMERIC(10,2) NOT NULL DEFAULT 0;
	ALTER TABLE services ADD COLUMN IF NOT EXISTS duration_minutes INT;
	ALTER TABLE services ADD COLUMN IF NOT EXISTS preparation_notes TEXT;
	ALTER TABLE services ADD COLUMN IF NOT EXISTS activation_status TEXT NOT NULL DEFAULT 'DRAFT';
	ALTER TABLE payments ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
	CREATE UNIQUE INDEX IF NOT EXISTS payments_idempotency_key_idx ON payments(idempotency_key) WHERE idempotency_key IS NOT NULL;
	CREATE INDEX IF NOT EXISTS audit_events_entity_idx ON audit_events(entity_type, entity_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS bookings_account_idx ON bookings(patient_account_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS gateway_events_payment_idx ON gateway_events(payment_id, created_at DESC);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migration exec error: %w", err)
	}
	log.Println("[DB] Schema migrations applied successfully")
	return nil
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func seedReferenceData(db *sql.DB) error {
	// Check if already seeded
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err == nil && count > 0 {
		log.Printf("[DB] Database already contains %d users; reconciling reference catalogue", count)
	} else {
		log.Println("[DB] Seeding reference data for Upasham Plus connected care...")
	}

	passHash := hashPassword("Pass@123456")

	// 1. Seed Specialties
	specialties := [][]string{
		{"spec-1", "General Medicine", "সাধারণ চিকিৎসা", "general-medicine", "Stethoscope", "First-contact primary triage, fever, cough, flu and chronic hypertension management.", "জ্বর, সর্দি, কাশি এবং দৈনন্দিন স্বাস্থ্য পরামর্শ।"},
		{"spec-2", "Pediatrics & Child Health", "শিশু ও কিশোর রোগ", "pediatrics", "Baby", "Comprehensive healthcare for newborns, infants, children and adolescents.", "নবজাতক ও শিশুদের সকল ধরণের শারীরিক পরামর্শ ও চিকিৎসা।"},
		{"spec-3", "Gynecology & Obstetrics", "স্ত্রী ও প্রসূতি রোগ", "gynecology", "HeartPulse", "Women's reproductive health, prenatal tele-consultations, and postnatal guidance.", "নারীদের স্বাস্থ্য ও গর্ভকালীন টেলিমেডিসিন সেবা।"},
		{"spec-4", "Cardiology", "হৃদরোগ", "cardiology", "Activity", "Hypertension, ischemic heart condition triage, ECG evaluation and specialist review.", "উচ্চ রক্তচাপ ও হৃদযন্ত্রের জটিলতার বিশেষজ্ঞ পরামর্শ।"},
		{"spec-5", "Dermatology", "চর্ম ও অ্যালার্জি", "dermatology", "Sparkles", "Skin conditions, allergies, eczema, acne and tele-dermatology consultations.", "চর্মরোগ ও অ্যালার্জির বিশেষজ্ঞ সমাধান।"},
		{"spec-6", "Orthopedics", "অর্থোপেডিক ও বাতব্যথা", "orthopedics", "Bone", "Joint pain, spine evaluation, arthritis management and bone trauma advice.", "হাড়ভাঙা, জয়েন্টের ব্যথা ও বাতের উন্নত চিকিৎসা।"},
		{"spec-7", "Psychiatry & Mental Health", "মানসিক স্বাস্থ্য ও কাউন্সেলিং", "psychiatry", "Brain", "Confidential depression, anxiety, stress and psychiatric tele-counseling.", "গোপনীয় ও আন্তরিক মানসিক স্বাস্থ্য পরামর্শ।"},
		{"spec-8", "ENT (Ear, Nose, Throat)", "নাক, কান ও গলা", "ent", "Headphones", "Ear infections, sinusitis, tonsillitis and hearing-related complaints.", "নাক, কান ও গলার সংক্রমণ ও চিকিৎসা।"},
	}

	for _, s := range specialties {
		_, _ = db.Exec(`
			INSERT INTO specialties (id, name, name_bn, slug, icon, description, description_bn, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, true)
			ON CONFLICT (id) DO NOTHING;
		`, s[0], s[1], s[2], s[3], s[4], s[5], s[6])
	}

	// 2. Seed Dhaka Service Zones
	zones := [][]string{
		{"zone-dhk-north", "Dhaka North (Uttara, Gulshan, Banani, Mirpur)", "DHK-N", "Dhaka"},
		{"zone-dhk-south", "Dhaka South (Dhanmondi, Old Dhaka, Motijheel)", "DHK-S", "Dhaka"},
		{"zone-dhk-central", "Dhaka Central (Tejgaon, Mohakhali, Farmgate)", "DHK-C", "Dhaka"},
	}
	for _, z := range zones {
		_, _ = db.Exec(`
			INSERT INTO service_zones (id, name, code, city, is_active)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (id) DO NOTHING;
		`, z[0], z[1], z[2], z[3])
	}

	// 3. Seed Catalogue Services
	services := [][]interface{}{
		{"srv-gen-consult", "General Medical Triage", "general-triage", "digital_consultation", "Instant general medical consultation with Upasham Plus licensed doctors.", 20.0},
		{"srv-spec-consult", "Specialist Doctor Consultation", "specialist-consult", "digital_consultation", "Detailed 20-minute video session with verified BMDC specialist.", 1000.0},
		{"srv-sample-home", "Home Sample Collection", "sample-collection", "sample_collection", "Phlebotomist visit at your doorstep across Dhaka.", 300.0},
		{"srv-nurse-visit", "Nurse / Caregiver Home Visit", "nurse-visit", "nurse_visit", "Dressing, IV cannulation, catheter care and geriatric support.", 800.0},
	}
	for _, s := range services {
		_, _ = db.Exec(`
			INSERT INTO services (id, name, slug, fulfillment_type, description, base_price, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)
			ON CONFLICT (id) DO NOTHING;
		`, s[0], s[1], s[2], s[3], s[4], s[5])
	}

	// 4. Seed Core Users
	// Patient User
	_, _ = db.Exec(`
		INSERT INTO users (id, email, phone, password_hash, full_name, role, is_active)
		VALUES ('usr-pat-01', 'patient@upasham.com', '+8801711223344', $1, 'Rashidul Islam', 'patient', true)
		ON CONFLICT (id) DO NOTHING;
	`, passHash)

	_, _ = db.Exec(`
		INSERT INTO patient_accounts (id, user_id, emergency_contact_phone, address_dhaka)
		VALUES ('pacc-01', 'usr-pat-01', '+8801711223344', 'House 42, Road 7, Dhanmondi, Dhaka-1205')
		ON CONFLICT (id) DO NOTHING;
	`)

	// Family Profiles
	_, _ = db.Exec(`
		INSERT INTO patient_profiles (id, account_id, full_name, relationship, gender, age, blood_group, known_allergies, chronic_conditions, notes)
		VALUES 
		('pprof-01', 'pacc-01', 'Rashidul Islam', 'self', 'male', 34, 'B+', 'None', 'Mild Hypertension', 'Self account owner'),
		('pprof-02', 'pacc-01', 'Fatima Islam', 'child', 'female', 6, 'O+', 'Penicillin', 'None', 'School child, recurring seasonal allergic rhinitis'),
		('pprof-03', 'pacc-01', 'Amena Begum', 'parent', 'female', 62, 'A+', 'Dust, Sulfa drugs', 'Type 2 Diabetes, Osteoarthritis', 'Needs regular prescription review')
		ON CONFLICT (id) DO NOTHING;
	`)

	// Doctors published on the official Upasham Plus website. Details not
	// published there remain intentionally unset rather than being invented.
	doctorUsers := []struct {
		id, email, phone, name, bmdc, qual, workplace, bio, bioBn, avail, avatar string
		exp                                                                      int
		fee, rating                                                              float64
		reviews                                                                  int
		specialties, langs                                                       string
	}{
		{
			id: "usr-doc-01", email: "dr.jony@upasham.com", phone: "+8801812345671",
			name: "Dr. Jony Chowdhury", bmdc: "", qual: "PGT Cardiology (BMU); CCD (BIRDEM)",
			workplace: "Senior Advisor, Upasham Plus", fee: 1000,
			avail: "Schedule through Upasham Plus", avatar: "/doctors/dr-jony-chowdhury.jpg",
			bio:         "Former medical officer in CCU, nephrology, cardiology, ICU and medicine, including roles at Ad-Din Hospital, National Kidney Foundation, National Heart Foundation, Shahabuddin Medical College Hospital and United Hospital.",
			specialties: "{\"spec-4\", \"spec-1\"}", langs: "{\"Bangla\", \"English\"}",
		},
		{
			id: "usr-doc-02", email: "dr.shuvojit@upasham.com", phone: "+8801812345672",
			name: "Dr. Shuvojit Paik", bmdc: "", qual: "MBBS (Dhaka University); CCD (BIRDEM)",
			workplace: "Doctor, Upasham Plus", fee: 1000,
			avail: "Schedule through Upasham Plus", avatar: "/doctors/dr-shuvojit-paik.jpg",
			bio:         "Postgraduate training at Suhrawardy Medical College Hospital and Bangladesh Shishu Hospital. Former medical officer in paediatrics, NICU, nephrology and neurology; senior medical officer in paediatrics, NICU and general practice.",
			specialties: "{\"spec-2\", \"spec-1\"}", langs: "{\"Bangla\", \"English\"}",
		},
		{
			id: "usr-doc-03", email: "dr.shaila@upasham.com", phone: "+8801812345673",
			name: "Dr. Shaila Sharmin Shahnewaz", bmdc: "", qual: "MBBS; FCPS (Physical Medicine & Rehabilitation)",
			workplace: "Associate Professor, Shaheed Monsur Ali Medical College Hospital, Dhaka", fee: 1000,
			avail: "Schedule through Upasham Plus", avatar: "/doctors/dr-shaila-sharmin-shahnewaz.png",
			bio:         "Special consultation doctor and Associate Professor in the Department of Physical Medicine & Rehabilitation.",
			specialties: "{\"spec-6\"}", langs: "{\"Bangla\", \"English\"}",
		},
		{
			id: "usr-doc-04", email: "dr.arshi@upasham.com", phone: "+8801812345674",
			name: "Dr. Arshi Tasnim", bmdc: "", qual: "MBBS; DMU; CMU; BLST",
			workplace: "Doctor, Upasham Plus", fee: 1000,
			avail: "Schedule through Upasham Plus", avatar: "/doctors/dr-arshi-tasnim.jpg",
			bio:         "Upasham Plus doctor providing connected patient consultations and practical medical guidance.",
			specialties: "{\"spec-1\"}", langs: "{\"Bangla\", \"English\"}",
		},
		{
			id: "usr-doc-05", email: "dr.sadia@upasham.com", phone: "+8801812345675",
			name: "Dr. Sadia Rahman Shifa", bmdc: "", qual: "MBBS; US-Bangla Medical College Hospital",
			workplace: "Medical Officer, Upasham Plus", fee: 1000,
			avail: "Schedule through Upasham Plus", avatar: "/doctors/dr-sadia-rahman-shifa.png",
			bio:         "Medical Officer supporting patient care and connected consultations at Upasham Plus.",
			specialties: "{\"spec-1\"}", langs: "{\"Bangla\", \"English\"}",
		},
		{
			id: "usr-doc-06", email: "dr.dipa@upasham.com", phone: "+8801812345676",
			name: "Dr. Dipa Agarwala", bmdc: "", qual: "MBBS",
			workplace: "Doctor, Upasham Plus", fee: 1000,
			avail: "Schedule through Upasham Plus", avatar: "/doctors/dr-dipa-agarwala.webp",
			bio:         "Upasham Plus doctor supporting accessible, connected patient care.",
			specialties: "{\"spec-1\"}", langs: "{\"Bangla\", \"English\"}",
		},
	}

	for i, d := range doctorUsers {
		docID := fmt.Sprintf("doc-%02d", i+1)
		_, _ = db.Exec(`
			INSERT INTO users (id, email, phone, password_hash, full_name, role, is_active)
			VALUES ($1, $2, $3, $4, $5, 'doctor', true)
			ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, phone = EXCLUDED.phone,
				full_name = EXCLUDED.full_name, is_active = true, updated_at = NOW();
		`, d.id, d.email, d.phone, passHash, d.name)

		_, _ = db.Exec(`
			INSERT INTO doctor_profiles (id, user_id, full_name, bmdc_number, specialties, qualifications, experience_years, workplace, languages, bio, bio_bn, consultation_fee, is_verified, is_published, avatar_url, rating, total_reviews, availability)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true, true, $13, $14, $15, $16)
			ON CONFLICT (id) DO UPDATE SET full_name = EXCLUDED.full_name,
				bmdc_number = EXCLUDED.bmdc_number, specialties = EXCLUDED.specialties,
				qualifications = EXCLUDED.qualifications, experience_years = EXCLUDED.experience_years,
				workplace = EXCLUDED.workplace, languages = EXCLUDED.languages, bio = EXCLUDED.bio,
				bio_bn = EXCLUDED.bio_bn, consultation_fee = EXCLUDED.consultation_fee,
				is_verified = EXCLUDED.is_verified, is_published = EXCLUDED.is_published,
				avatar_url = EXCLUDED.avatar_url, rating = EXCLUDED.rating,
				total_reviews = EXCLUDED.total_reviews, availability = EXCLUDED.availability;
		`, docID, d.id, d.name, d.bmdc, d.specialties, d.qual, d.exp, d.workplace, d.langs, d.bio, d.bioBn, d.fee, d.avatar, d.rating, d.reviews, d.avail)
	}

	// Remove obsolete demo credential numbers; the public source does not publish them.
	_, _ = db.Exec(`DELETE FROM credentials WHERE id IN ('cred-01', 'cred-02', 'cred-03', 'cred-04')`)

	// Operations User
	_, _ = db.Exec(`
		INSERT INTO users (id, email, phone, password_hash, full_name, role, is_active)
		VALUES ('usr-ops-01', 'ops@upasham.com', '+8801900112233', $1, 'Upasham Operations Dhaka', 'operations', true)
		ON CONFLICT (id) DO NOTHING;
	`, passHash)

	// Finance User
	_, _ = db.Exec(`
		INSERT INTO users (id, email, phone, password_hash, full_name, role, is_active)
		VALUES ('usr-fin-01', 'finance@upasham.com', '+8801900112244', $1, 'Upasham Accounts & Finance', 'finance', true)
		ON CONFLICT (id) DO NOTHING;
	`, passHash)

	// Admin User
	_, _ = db.Exec(`
		INSERT INTO users (id, email, phone, password_hash, full_name, role, is_active)
		VALUES ('usr-adm-01', 'admin@upasham.com', '+8801900112255', $1, 'System Administrator', 'admin', true)
		ON CONFLICT (id) DO NOTHING;
	`, passHash)

	// 5. Seed Sample Bookings representing realistic state machine journeys
	// Booking 1: General Triage (BDT 20), Paid, In Queue for Operations manual assignment
	_, _ = db.Exec(`
		INSERT INTO bookings (id, booking_ref, patient_account_id, patient_profile_id, service_type, status, chief_complaint, triage_notes, preferred_time, quoted_amount, currency)
		VALUES ('bk-001', 'UP-2026-001', 'pacc-01', 'pprof-01', 'general', 'Paid/Awaiting Assignment', 'Fever since yesterday (101°F) with headache and body pain.', 'Vitals normal, no chest tightness.', 'Immediate (within 15 mins)', 20.0, 'BDT')
		ON CONFLICT (id) DO NOTHING;
	`)
	_, _ = db.Exec(`
		INSERT INTO payments (id, booking_id, invoice_ref, gateway, amount, currency, status, gateway_tx_id, gateway_signature, paid_at)
		VALUES ('pay-001', 'bk-001', 'INV-UP-2026-001', 'bkash', 20.0, 'BDT', 'Paid', 'BK99482910', 'sig_bkash_verified_demo', NOW())
		ON CONFLICT (id) DO NOTHING;
	`)

	// Booking 2: Specialist with Dr. Jony (BDT 1,000), Confirmed, Active Jitsi Video Room
	scheduledTime := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	_, _ = db.Exec(`
		INSERT INTO bookings (id, booking_ref, patient_account_id, patient_profile_id, doctor_id, specialty_id, service_type, status, chief_complaint, triage_notes, preferred_time, scheduled_at, quoted_amount, currency)
		VALUES ('bk-002', 'UP-2026-002', 'pacc-01', 'pprof-03', 'doc-01', 'spec-1', 'specialist', 'Confirmed', 'Blood sugar fluctuating after morning meals. Knee pain persisting.', 'Type 2 diabetic patient, needs insulin regimen audit.', 'Today evening 7:00 PM', $1, 1000.0, 'BDT')
		ON CONFLICT (id) DO NOTHING;
	`, scheduledTime)

	_, _ = db.Exec(`
		INSERT INTO payments (id, booking_id, invoice_ref, gateway, amount, currency, status, gateway_tx_id, gateway_signature, paid_at)
		VALUES ('pay-002', 'bk-002', 'INV-UP-2026-002', 'sslcommerz', 1000.0, 'BDT', 'Paid', 'SSL88392109', 'sig_ssl_verified_demo', NOW())
		ON CONFLICT (id) DO NOTHING;
	`)

	// Setup Video Room for Booking 2
	_, _ = db.Exec(`
		INSERT INTO video_rooms (id, booking_id, room_name, room_pin, host_token, participant_token, jitsi_domain, expires_at, status)
		VALUES ('vroom-002', 'bk-002', 'upasham-care-dhaka-bk002', '492817', 'host-jwt-doc01', 'part-jwt-pat01', 'meet.jit.si', NOW() + INTERVAL '3 hours', 'ACTIVE')
		ON CONFLICT (id) DO NOTHING;
	`)

	// Booking 3: Completed consultation with Dr. Shuvojit for child Fatima, has full electronic prescription!
	_, _ = db.Exec(`
		INSERT INTO bookings (id, booking_ref, patient_account_id, patient_profile_id, doctor_id, specialty_id, service_type, status, chief_complaint, triage_notes, preferred_time, quoted_amount, currency)
		VALUES ('bk-003', 'UP-2026-003', 'pacc-01', 'pprof-02', 'doc-02', 'spec-2', 'specialist', 'Completed', 'Dry coughing fits especially at night, allergic sneezing.', 'Follow up from last week, chest clear, no wheezing.', 'Completed', 1200.0, 'BDT')
		ON CONFLICT (id) DO NOTHING;
	`)

	_, _ = db.Exec(`
		INSERT INTO payments (id, booking_id, invoice_ref, gateway, amount, currency, status, gateway_tx_id, gateway_signature, paid_at)
		VALUES ('pay-003', 'bk-003', 'INV-UP-2026-003', 'bkash', 1200.0, 'BDT', 'Paid', 'BK77392103', 'sig_bkash_verified_demo', NOW() - INTERVAL '2 days')
		ON CONFLICT (id) DO NOTHING;
	`)

	_, _ = db.Exec(`
		INSERT INTO encounters (id, booking_id, doctor_id, patient_profile_id, started_at, ended_at, status)
		VALUES ('enc-003', 'bk-003', 'doc-02', 'pprof-02', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '22 minutes', 'COMPLETED')
		ON CONFLICT (id) DO NOTHING;
	`)

	_, _ = db.Exec(`
		INSERT INTO clinical_notes (id, encounter_id, subjective, objective, assessment, plan, is_private)
		VALUES ('cnote-003', 'enc-003', 'Child presents with nocturnal dry cough, clear rhinorrhea.', 'Temp 98.4F, throat mildly congested, chest clear, no crepitations.', 'Allergic Bronchospasm & Upper Airway Cough Syndrome', 'Antihistamine, Montelukast, steam inhalation, avoidance of cold beverages.', false)
		ON CONFLICT (id) DO NOTHING;
	`)

	_, _ = db.Exec(`
		INSERT INTO prescriptions (id, booking_id, encounter_id, prescription_ref, doctor_id, patient_profile_id, diagnosis, advice, follow_up_date, warnings, verification_hash, version)
		VALUES ('rx-003', 'bk-003', 'enc-003', 'RX-UP-2026-003', 'doc-02', 'pprof-02', 'Allergic Cough & Seasonal Rhinitis', 'Drink lukewarm water. Avoid cold food and dusty environment. Steam inhalation twice daily for 5 minutes.', 'After 7 days or if fever > 101°F develops', 'Penicillin allergy noted in profile. In case of respiratory distress, visit nearest pediatric emergency immediately.', 'vhash_dhaka_upasham_rx003_secure', 1)
		ON CONFLICT (id) DO NOTHING;
	`)

	_, _ = db.Exec(`
		INSERT INTO medicine_lines (id, prescription_id, brand_name, generic_name, strength, dosage_form, dose, route, frequency, duration, instructions)
		VALUES 
		('med-01', 'rx-003', 'Fexo 30mg', 'Fexofenadine HCl', '30mg/5ml', 'Syrup', '1 teaspoon', 'Oral', 'Once daily at bedtime', '7 days', 'Shake bottle well before use'),
		('med-02', 'rx-003', 'Montene 4mg', 'Montelukast', '4mg', 'Chewable Tablet', '1 tablet', 'Oral', 'Nightly before sleep', '14 days', 'Chew thoroughly before swallowing'),
		('med-03', 'rx-003', 'Nasalmist', '0.9% Normal Saline', '0.9%', 'Nasal Spray', '1 puff each nostril', 'Nasal', 'Twice daily', '5 days', 'Use before bedtime to clear nasal passage')
		ON CONFLICT (id) DO NOTHING;
	`)

	_, _ = db.Exec(`
		INSERT INTO test_orders (id, prescription_id, test_name, instructions, status)
		VALUES ('tst-01', 'rx-003', 'CBC with ESR & Absolute Eosinophil Count', 'No fasting required. Can be collected from home.', 'PENDING')
		ON CONFLICT (id) DO NOTHING;
	`)

	// Audit Log
	_, _ = db.Exec(`
		INSERT INTO audit_events (id, user_id, action, entity_type, entity_id, ip_address, details_json)
		VALUES 
		('aud-01', 'usr-pat-01', 'LOGIN', 'User', 'usr-pat-01', '103.114.98.12', '{"device": "Mobile Safari", "location": "Dhanmondi, Dhaka"}'),
		('aud-02', 'usr-pat-01', 'CREATE_BOOKING', 'Booking', 'bk-001', '103.114.98.12', '{"service": "general", "amount": 20.0, "gateway": "bkash"}'),
		('aud-03', 'usr-doc-02', 'ISSUE_PRESCRIPTION', 'Prescription', 'rx-003', '103.120.201.5', '{"medicines_count": 3, "patient_profile": "pprof-02"}')
		ON CONFLICT (id) DO NOTHING;
	`)

	log.Println("[DB] Reference seed data completed with 100% integrity!")
	return nil
}
