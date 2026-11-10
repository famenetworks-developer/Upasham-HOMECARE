# উপশম প্লাস (Upasham Plus) - Connected Care Platform

> **"Doctors from far, not so far. A connected healthcare platform for Dhaka, designed to grow across Bangladesh."**  
> *"শহর থেকে গ্রামে, ডাক্তার এখন কাছে"*

---

## 🌟 Executive Summary

Upasham Plus is a full-stack, connected-care telemedicine and clinical governance platform built for Dhaka city. It implements:
- **BDT 20 General Consultation Triage**: Instant access to on-duty licensed doctors.
- **Specialist Consultations (BDT 1,000+ fee)**: Verified BMDC specialist physicians across Dhaka institutions.
- **Protected Jitsi Video Consultations**: Encrypted in-app tele-consultations with short-lived tokens and attendance tracking.
- **Electronic Medical Prescriptions & Branded PDFs**: Official prescriptions with diagnosis, structured medicine lines, diagnostic test orders, and cryptographic SHA-256 verification.
- **Separated Family Health History**: Manage self, children, parents, and spouse under a single account.
- **Operations Queue**: Operations manually assigns doctors and confirms video consultation slots.
- **Provider Adapters**: Direct bKash and SSLCOMMERZ checkout adapters with server-calculated quotations and idempotent webhooks.
- **Safeer Editorial Design System**: Inspired by [safeertemplate.webflow.io](https://safeertemplate.webflow.io/) using `#D5FF55` lime, `#FEBD1A` gold accent, `#0F0A01` dark slate, combined with Upasham brand colors (`#16A34A` green, `#0284C7` blue).
- **Bilingual Interface**: Seamless switching between বাংলা (Bengali) and English.

---

## 🏗️ Architecture Stack

- **Backend:** Go (Golang 1.26) + Gin HTTP Framework
- **Database:** PostgreSQL 17 (`upasham_db`)
- **Frontend:** React 19 + TypeScript + Vite
- **Styling:** CSS Design Tokens based on Safeer layout
- **PDF Engine:** Go native `gofpdf`
- **Video:** Jitsi Meet WebRTC integration

---

## 🚀 Quick Start Instructions

### 1. Database
PostgreSQL is running on localhost port 5433 with user `postgres` and database `upasham_db`.

### 2. Backend (Go Server)
```powershell
cd backend
go build -o server.exe ./cmd/server
.\server.exe
# Server starts on http://localhost:8088
```

### 3. Frontend (React + Vite)
```powershell
cd frontend
npm install
npm run dev
# Frontend starts on http://localhost:3005
```

---

## 👥 Demo Pre-Seeded Accounts

Use the sticky **"LIVE WORKSPACE"** toolbar at the top of the webpage to instantly test every role with 1 click:

1. **Patient (রাশিদুল ইসলাম):** `patient@upasham.com` / `Pass@123456`
2. **Doctor (Dr. Jony Chowdhury):** `dr.jony@upasham.com` / `Pass@123456`
3. **Doctor (Dr. Shuvojit Paik):** `dr.shuvojit@upasham.com` / `Pass@123456`
4. **Operations (ঢাকা অপারেশন্স):** `ops@upasham.com` / `Pass@123456`
5. **Finance (ফাইন্যান্স ও অডিট):** `finance@upasham.com` / `Pass@123456`
6. **Administrator (সুপার এডমিন):** `admin@upasham.com` / `Pass@123456`

---

## 🧪 Automated Test Execution

## Standalone Content Management Dashboard

Open `http://localhost:3005/#admin-cms` to enter the dedicated administrator CMS in the local demo environment. The dashboard is separate from the clinical governance portal and provides:

- Homepage editing with bilingual draft preview, publishing, revision history, and restore.
- CRUD collections for pages, articles, navigation items, and categories.
- Search, status filters, ordering, draft/published/archived states, and confirmed soft deletion.
- An image-only media library (10 MB upload limit) served from `/uploads`.
- Doctor directory, user/role inventory, audit activity, and CMS health information.

All mutating routes require an authenticated `admin` JWT and create audit records. CMS data is stored in PostgreSQL tables `cms_entries`, `media_assets`, `site_content`, and `site_content_revisions`.

### CMS API namespace

The protected endpoints are under `/api/admin/cms/*`. Uploaded files are stored in `backend/uploads`; persistent deployments should mount this directory or replace it with object storage.

### Go Unit & Integration Tests:
```powershell
cd backend
go test ./...
# Verified: state machine transitions, payment adapters, JWT, video token logic
```

### E2E Journey Verification Script:
```powershell
powershell -ExecutionPolicy Bypass -File scratch/verify_full_journey.ps1
# All 12 vertical slices executed and verified with 100% success!
```
