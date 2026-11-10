# Upasham Plus Specification Compliance Matrix

Audit basis: `Upasham_Plus_Codex_Project_Prompt.pdf`, version 1.0, dated 13 October 2026.

Status meanings: **Implemented** is locally testable; **Foundation** means the schema/configuration boundary exists but the external service is not activated; **Open** requires credentials, infrastructure, policy, or additional product work.

## Product rules and surfaces

| Requirement | Status | Evidence / note |
|---|---|---|
| Configurable BDT 20 general consultation | Implemented | Server quotation service; client amount is ignored. |
| Configurable specialist price | Implemented | Stored per doctor and returned by server quotation. |
| Pay before operations confirmation | Implemented | Operations queue joins paid payments; assignment starts from Paid/Awaiting Assignment. |
| Manual operations assignment and confirmation | Implemented | Operations queue, assignment, confirmation, reschedule, cancellation routes and UI. |
| Patient family profiles and separated history | Implemented | Account-owned profiles; object checks on profile, booking and prescription access. |
| Doctor directory and profile details | Implemented | Public list/detail APIs, filters, publish/verify controls; private contacts excluded. |
| Patient, doctor, operations, finance, admin portals | Implemented | Role-protected routes and React surfaces. |
| Native mobile apps excluded | Implemented | Responsive web only. |

## Identity, authorization, and security

| Requirement | Status | Evidence / note |
|---|---|---|
| JWT authentication and server RBAC | Implemented | Bearer-token middleware and role middleware. |
| Explicit Role/Permission domain | Implemented | Roles, permissions and role-permission migration tables; current authorization remains role-based. |
| Patient object separation | Implemented | Account ownership checks on profiles, bookings, payments and prescription PDFs. |
| Doctor assignment separation | Implemented | Assignment checks for booking, encounter, prescription and meeting access. |
| Finance clinical-data restriction | Implemented | Finance routes expose financial data; prescription PDFs deny finance role. |
| Rate limiting | Implemented | Login, registration, quotation, booking, initiation and callback limiters. |
| Security headers and correlation IDs | Implemented | CSP, frame/content policies, referrer/permissions policy, HSTS behind HTTPS proxy. |
| Sensitive read/admin-write auditing | Partial | Prescription PDF reads and core admin/clinical/operations/refund actions audited. A centralized audit middleware remains desirable. |
| CSRF | Not applicable to current auth transport | API uses Authorization bearer tokens, not ambient authentication cookies. Reassess if cookies are introduced. |
| Production HTTPS and encrypted backups | Open | Deployment/infrastructure responsibility; documented in runbook. |

## Booking, payment, and finance

| Requirement | Status | Evidence / note |
|---|---|---|
| Server state-machine validation | Implemented for principal flows | Transition service plus checks on assignment, confirmation, cancellation and meeting access. |
| Database transactions | Implemented for booking, callback, clinical completion and refunds | Multi-record state changes use SQL transactions. |
| bKash and SSLCOMMERZ adapter interface | Foundation | Sandbox adapters exist and credentials are environment-only. Real provider HTTP exchange/signature contracts require merchant credentials. |
| Callback idempotency and payload redaction | Implemented for sandbox flow | Paid callbacks are idempotent; credential-like keys are redacted before persistence. |
| Reconciliation exception queue/job | Open | Finance event dashboard exists; retry-safe scheduled reconciliation worker is not implemented. |
| Refund records and authorized workflow | Implemented locally | Finance/admin-only transactional refund records. Provider settlement remains manual. |
| Finance exports | Open | Transactions are visible; CSV export is not implemented. |
| Minor-unit money representation | Partial | Package foundation uses minor units; existing consultation/payment columns use fixed-point NUMERIC and Go decimal-like values, not floats in PostgreSQL. Go models still use `float64`. |

## Video and clinical records

| Requirement | Status | Evidence / note |
|---|---|---|
| Random meeting identity | Implemented | Cryptographically random room names and PINs. |
| Object-authorized, expiring meeting access | Implemented | Correct patient/doctor checks, booking-state check, join window and expiry check. |
| Jitsi authentication strategy | Foundation | Domain is environment-configured; production Jitsi JWT/signing infrastructure is not configured. |
| Attendance logs / no media recording | Implemented | Join/leave events are stored; no recording workflow exists. |
| Jitsi outage operations notification | Partial | UI fallback exists; persistent operations notification on outage is open. |
| Structured encounter/prescription/version | Implemented | Notes, private notes, medicines, tests, advice, follow-up, warnings and snapshots. |
| Branded prescription PDF | Implemented | Doctor, BMDC, patient, booking reference and verification hash included. |
| Protected prescription files | Implemented | Route requires authentication and object-level authorization; frontend downloads with bearer token. |
| Patient reports/uploads | Open | Diagnostic-report domain exists; upload/storage workflow is not activated. |

## Expansion foundation

Lab partners, sample jobs, diagnostic reports, home-service jobs, provider assignments, packages, package items and subscriptions have dormant schema support. Services include fulfillment type, zones, base/travel price, duration, preparation notes and activation status. Unfinished flows remain unexposed.

## Engineering and delivery

| Requirement | Status | Evidence / note |
|---|---|---|
| PostgreSQL schema and seed data | Implemented | Startup migrations and synthetic Dhaka demo data. |
| API contracts and runbooks | Implemented | `API_CONTRACTS.md`, `RUNBOOK.md`, architecture/security/user guides. |
| Unit tests | Partial | Auth, state transitions, adapters, room randomness and middleware covered. |
| Integration/E2E/security/operational suite | Open | Full automated matrix described by the PDF is not yet present. |
| Production deployment | Not authorized | No production deployment, live payment activation or real patient data use performed. |

## Release decision

The repository is a strong local/staging MVP, not a production-complete medical platform. Production launch must remain blocked until real gateway verification/reconciliation, managed Jitsi authentication, notification delivery, report storage, retention/data-correction jobs, infrastructure hardening, complete security/E2E/operational tests, and legal/clinical review are completed.
