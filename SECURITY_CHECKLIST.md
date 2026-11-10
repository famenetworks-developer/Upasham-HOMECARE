# Security and Privacy Checklist - Upasham Plus Platform

**Assessment Date:** October 2026  
**Status:** Implemented & Verified  

---

## 1. Authentication & Session Security
- [x] **Password Hashing:** Passwords hashed with bcrypt (`cost=10`). Plaintext passwords never stored.
- [x] **Stateless JWT Tokens:** 24-hour expiration tokens signed with HMAC-SHA256 (`HS256`).
- [x] **Bearer Token Authorization:** Extracted via `Authorization: Bearer <token>` middleware.
- [x] **Session Deactivation:** Inactive users immediately barred at login.

---

## 2. Role-Based Access Control (RBAC) & Object Authorization
- [x] **Server-Side Enforcement:** Route-level RBAC middleware for `patient`, `doctor`, `operations`, `finance`, and `admin`.
- [x] **Patient Family Separation:** Object-level check verifies that patient accounts can ONLY query or book for their own linked `patient_profiles`.
- [x] **Doctor Scope of Practice:** Doctors can ONLY access encounters and clinical notes assigned to their doctor profile ID.
- [x] **Finance Isolation:** Finance users receive financial ledger and gateway metadata only, with zero unrestricted access to patient clinical notes.
- [x] **Supervisory Admin Logging:** Administrative modifications (e.g. verifying doctor BMDC credentials or toggling directory publication) are logged to the immutable `audit_events` table.

---

## 3. Video Meeting Privacy & Telehealth Controls
- [x] **Unguessable Room Identifiers:** Rooms generated via 24-character cryptographic random hex strings (`upasham-care-<hex>`); never derived from patient phone numbers or sequential IDs.
- [x] **Short-Lived Authorization:** Meeting tokens expire within a 4-hour window from scheduling.
- [x] **Object Access Validation:** Users joining `/api/video/rooms/:id/access` must be either the assigned doctor or the booking patient. Uninvited users receive HTTP 403 Forbidden.
- [x] **No Recording in Initial Release:** In accordance with Section 10 rules, consultation media is not recorded on server storage.
- [x] **Audit Telemetry:** Join and leave attendance timestamps recorded in `attendance_events`.

---

## 4. Payment & Financial Data Controls
- [x] **Server-Authoritative Pricing:** Client-posted pricing amounts are ignored. Quotations are calculated strictly on the server (`/api/quotation`).
- [x] **Payload Redaction:** Gateway webhooks strip passwords, API keys, and credit card numbers (`[REDACTED]`) before database insertion.
- [x] **Idempotent Webhooks:** Repeated gateway callbacks return HTTP 200 without duplicating revenue or corrupting state transitions.
- [x] **Controlled Refund Workflow:** Transactions are never silently deleted or overwritten; refunds require explicit justification and are persisted to `refunds` table.

---

## 5. Audit Logging Architecture
- [x] **Immutable Trail:** The `audit_events` table captures user ID, action (`LOGIN`, `PAYMENT_VERIFIED`, `ASSIGN_DOCTOR`, `CONFIRM_BOOKING`, `ISSUE_PRESCRIPTION`, `REFUND_PAYMENT`), IP address, entity references, and timestamps.
