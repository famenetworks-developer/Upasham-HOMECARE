# API and Webhook Contracts - Upasham Plus Platform

All endpoints are hosted under `/api`. Standard responses use JSON format with HTTP status codes.

---

## 1. Authentication Endpoints

### `POST /api/auth/register`
Creates a patient user, linked patient account, and initial primary profile.
```json
{
  "full_name": "Rashidul Islam",
  "email": "patient@upasham.com",
  "phone": "+8801711223344",
  "password": "Password@123",
  "emergency_contact_phone": "+8801711223344",
  "address_dhaka": "Dhanmondi 7, Dhaka"
}
```
**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn...",
  "user": {
    "id": "usr-01",
    "email": "patient@upasham.com",
    "role": "patient",
    "patient_account_id": "pacc-01"
  }
}
```

### `POST /api/auth/login`
Validates credentials and returns JWT Bearer token with claims.

### `POST /api/auth/demo-login`
Utility endpoint allowing 1-click test authentication across roles:
```json
{ "role": "patient" | "doctor_farhana" | "doctor_tanvir" | "operations" | "finance" | "admin" }
```

---

## 2. Public Endpoints

### `GET /api/doctors`
Returns public list of published and BMDC-verified physicians.
- Query parameters: `?specialty=spec-1&query=Farhana&language=Bangla`
- Never exposes doctor phone number or personal credentials.

### `GET /api/catalogue`
Returns medical specialties, services catalogue, and Dhaka service zones.

### `POST /api/quotation`
Server-side quotation engine (Rule 11: never trust client amounts).
```json
{
  "service_type": "general" | "specialist",
  "doctor_id": "doc-01"
}
```
**Response (200 OK):**
```json
{
  "service_type": "general",
  "amount": 20.0,
  "currency": "BDT"
}
```

---

## 3. Patient Portal Endpoints (Require `patient` Role)

### `GET /api/patient/profiles`
Returns self and family member profiles for the authenticated patient account.

### `POST /api/patient/profiles`
Adds a new family member:
```json
{
  "full_name": "Fatima Islam",
  "relationship": "child",
  "gender": "female",
  "age": 6,
  "blood_group": "O+",
  "known_allergies": "Penicillin",
  "chronic_conditions": "Seasonal Rhinitis"
}
```

### `POST /api/patient/bookings`
Initiates a consultation in `Payment Pending` status with mandatory telehealth consent:
```json
{
  "patient_profile_id": "pprof-02",
  "service_type": "general",
  "chief_complaint": "High fever and cough",
  "preferred_time": "Immediate (within 15 minutes)",
  "consent_given": true
}
```

### `POST /api/patient/payments/initiate`
Initiates payment via selected gateway adapter (`bkash` or `sslcommerz`):
```json
{
  "booking_id": "bk-001",
  "gateway": "bkash",
  "customer_phone": "+8801711223344",
  "return_url": "/patient/callback"
}
```

---

## 4. Payment Gateway Webhook / Callback

### `POST /api/payments/callback`
Idempotent webhook handler. Redacts sensitive credentials before database persistence.
```json
{
  "invoice_ref": "INV-UP-2026-001-8486",
  "gateway": "bkash",
  "status": "success",
  "trx_id": "BK99482910",
  "raw_data": {
    "status": "success",
    "trxID": "BK99482910",
    "amount": 20.0
  }
}
```
**State Mutation:**
- Transitions booking to `Paid/Awaiting Assignment`.
- Appends to `booking_status_histories`.
- Writes audit log to `audit_events`.
- Returns HTTP 200 idempotently if already verified.

---

## 5. Operations Portal Endpoints (Require `operations` or `admin` Role)

### `GET /api/operations/queue`
Returns unassigned and active triages awaiting manual action.

### `POST /api/operations/assign`
Operations officer assigns an on-duty BMDC doctor:
```json
{
  "booking_id": "bk-001",
  "doctor_id": "doc-01",
  "notes": "Assigned Dr. Farhana based on Dhaka triage availability"
}
```

### `POST /api/operations/confirm`
Schedules consultation time and provisions encrypted Jitsi room:
```json
{
  "booking_id": "bk-001",
  "scheduled_at": "2026-10-13T20:30:00Z"
}
```

---

## 6. Doctor Portal Endpoints (Require `doctor` Role)

### `GET /api/doctor/assignments`
Returns assigned patient queue with patient clinical context and allergies.

### `POST /api/doctor/clinical/save`
Concludes encounter, creates structured prescription, and generates verification hash:
```json
{
  "booking_id": "bk-001",
  "diagnosis": "Acute Allergic Bronchospasm",
  "medicines": [
    {
      "brand_name": "Montene 4mg",
      "strength": "4mg",
      "dose": "1 tablet",
      "frequency": "Nightly",
      "duration": "14 days",
      "instructions": "Chew before sleep"
    }
  ],
  "advice": "Avoid cold drinks",
  "warnings": "Report to emergency if breathing difficulty occurs"
}
```

---

## 7. Video Consultation Access

### `GET /api/video/rooms/:id/access`
Issues short-lived server tokens for either doctor (host) or patient (participant).
Validates object-level access; rejects uninvited callers.

### `POST /api/video/attendance`
Logs join and leave events for tele-consultation audit records:
```json
{
  "video_room_id": "vroom-001",
  "event_type": "join" | "leave"
}
```

---

## 8. Prescription Stream & PDF

### `GET /api/prescriptions/:id/pdf`
Streams branded PDF directly with:
- Upasham Plus green banner
- Doctor BMDC registration & hospital affiliation
- Patient demographics & booking reference
- Rx prescription table
- Red-flag warning box
- Cryptographic verification hash & digital signature notice.
