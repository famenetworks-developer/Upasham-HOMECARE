# Architecture Decision Record (ADR) - Upasham Plus Connected Care Platform

**Document Version:** 1.0  
**Date:** October 2026  
**Status:** Approved & Implemented  
**Platform:** Upasham Plus (উপশম প্লাস) — Connected Healthcare for Dhaka  

---

## 1. System Overview & Mission

Upasham Plus is a secure, responsive patient, doctor, and operations web platform tailored for Dhaka city. It enables:
1. **BDT 20 General Consultation Triage** with available on-duty licensed doctors.
2. **Specialist Consultations (BDT 1,000 reference base)** with BMDC-verified physicians across Dhaka's premier institutions (Dhaka Medical College, BIRDEM, Square Hospital, NICVD).
3. **Protected Jitsi Video Consultations** with unguessable, short-lived room authorization.
4. **Official Electronic Prescriptions** with clinical diagnosis, structured medicine lines, diagnostic orders, advice, red-flag emergency warnings, cryptographic verification hashes, and branded PDF generation.
5. **Separated Family Health History** under a single account (self, child, spouse, parent).
6. **Operations-governed Manual Assignment & Confirmation**: Patients pay before operational confirmation; operations officers assign doctors and schedule video slots.
7. **Bilingual User Experience**: Complete Bengali (বাংলা) and English interface.
8. **Safeer Editorial Design System**: Editorial typography, high contrast badges, rounded corners (3px, 12px, 999px), card elevation, and responsive layouts.

---

## 2. Technical Stack Decision

| Layer | Selected Technology | Rationale |
|---|---|---|
| **Backend** | Go (Golang 1.26) + Gin | High throughput, sub-millisecond response times, type safety, memory efficiency, single-binary deployment. |
| **Database** | PostgreSQL 17 | Relational ACID compliance for transactional payments, booking state machines, and relational patient history. |
| **Frontend** | React 19 + TypeScript + Vite | Componentized responsive UI, type safety, instant hot reload, and modular surface rendering. |
| **Styling** | Vanilla CSS + Design Tokens | Custom tokenized styling inspired by Safeer template (`#D5FF55` lime, `#0F0A01` dark text, `#FEBD1A` accent) combined with Upasham brand colors (`#16A34A` green, `#0284C7` blue). |
| **Telehealth** | Jitsi Meet WebRTC | Open standards, environment-configurable, short-lived room identity tokens, zero server media recording in initial release. |
| **PDF Engine** | `github.com/jung-kurt/gofpdf` | Native Go PDF generation producing compliant, digitally signed medical prescription documents. |
| **Payments** | Provider Adapters (bKash & SSLCOMMERZ) | Server-persisted quotations, idempotent callbacks, HMAC signatures, redacted storage. |

---

## 3. Booking State Machine

```
   [ Draft ]
       │
       ▼
[ Payment Pending ] ────────► [ Cancelled ]
       │
       ▼ (Callback Verified)
[ Paid/Awaiting Assignment ] ──► [ Refund Pending ] ──► [ Refunded ]
       │
       ▼ (Operations Manual Assignment)
   [ Assigned ]
       │
       ▼ (Operations Confirms Slot & Jitsi Room)
  [ Confirmed ] ────────────► [ No-show / Rescheduled ]
       │
       ▼ (Join Video Consultation)
  [ In Progress ]
       │
       ▼ (Prescription Issued)
  [ Completed ]
```

---

## 4. Payment Lifecycle State Machine

```
[ Initiated ] ──► [ Processing ] ──► [ Paid ] ──► [ Refund Requested ] ──► [ Refunded ]
       │                                     └──► [ Partially Refunded ]
       └──► [ Failed / Expired ]
       └──► [ Reconciliation Exception ]
```

---

## 5. Domain Entities & Database Schema

The database schema (`upasham_db`) contains explicit normalized tables:
- `users`: Core authentication identity & RBAC role (`patient`, `doctor`, `operations`, `finance`, `admin`).
- `patient_accounts`: Account container managing billing address and emergency contacts.
- `patient_profiles`: Separated patient medical history records (self, children, parents, spouse) with age, gender, blood group, allergies, and chronic conditions.
- `doctor_profiles`: BMDC license, qualifications, workplace, consultation fee, verification flag, publication flag.
- `specialties`: Medical specialties with bilingual names and slugs.
- `credentials`: Doctor verification certificates and BMDC registration records.
- `service_zones`: Dhaka North, South, and Central zones.
- `services`: Digital consultations (BDT 20 / 1000) and dormant home services.
- `bookings`: Central booking lifecycle entity.
- `booking_status_histories`: Immutable status audit trail.
- `assignments`: Operational doctor-to-booking assignment record.
- `payments`: Transactions, invoice references, and gateway references.
- `gateway_events`: Redacted webhook payloads and signature verification statuses.
- `refunds`: Refund authorization and gateway transaction references.
- `video_rooms`: Random unguessable room identifiers and token lifetimes.
- `attendance_events`: Join/leave telemetry for telehealth audit compliance.
- `encounters`: Doctor-patient clinical consultation sessions.
- `clinical_notes`: Subjective, objective, and private physician notes.
- `prescriptions`: Official medical prescriptions with version numbers and verification hashes.
- `medicine_lines`: Structured dosage lines (brand, generic, dose, route, freq, duration, instructions).
- `test_orders`: Recommended diagnostic investigations.
- `audit_events`: Security, financial, and clinical mutation audit trail.
- Dormant models: `lab_partners`, `sample_jobs`, `diagnostic_reports`, `home_service_jobs`.
