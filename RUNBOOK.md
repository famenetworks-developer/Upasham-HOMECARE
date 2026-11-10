# Operational Runbook - Upasham Plus Connected Care Platform

This document describes how to deploy, manage, back up, and troubleshoot the Upasham Plus web platform.

---

## 1. Prerequisites
- **Go:** 1.24+ (Tested on Go 1.26.2)
- **Node.js:** v20+ (Tested on Node v24.14.1, npm 11.11.0)
- **PostgreSQL:** 16+ or 17 (Tested on native PostgreSQL 17 on Windows / port 5433)

---

## 2. Environment Configuration

Copy the template to your deployment environment:
```bash
cp .env.example .env
```
Key variables:
- `PORT`: Backend HTTP port (default `8088`)
- `DATABASE_URL`: `postgres://postgres:postgres@127.0.0.1:5433/upasham_db?sslmode=disable`
- `JWT_SECRET`: 32+ byte cryptographic secret
- `JITSI_DOMAIN`: `meet.jit.si` (or private Jitsi cluster)
- `BKASH_APP_KEY`: bKash sandbox/production merchant key
- `SSLCOMMERZ_STORE_ID`: SSLCOMMERZ merchant store ID

---

## 3. Database Initialization & Migrations

The Go backend automatically applies schema migrations and baseline reference seed data on startup.
To verify or reset the database manually using PostgreSQL `psql`:

```powershell
# Create database
& "C:\Program Files\PostgreSQL\17\bin\psql.exe" -h 127.0.0.1 -p 5433 -U postgres -c "CREATE DATABASE upasham_db;"

# Verify tables
& "C:\Program Files\PostgreSQL\17\bin\psql.exe" -h 127.0.0.1 -p 5433 -U postgres -d upasham_db -c "\dt"
```

---

## 4. Building & Running Locally

### Backend (Go)
```powershell
cd backend
go build -o server.exe ./cmd/server
.\server.exe
# Output: [SERVER] Listening on http://localhost:8088
```

### Automated Backend Tests
```powershell
cd backend
go test ./...
# Runs unit tests for state machines, payment adapters, JWT, video token logic
```

### Frontend (React + Vite)
```powershell
cd frontend
npm install
npm run build   # Validates TypeScript & bundles assets
npm run dev     # Runs on http://localhost:3005
```

---

## 5. Seed Users for Immediate Evaluation

| Role | Email | Password | Pre-seeded Context |
|---|---|---|---|
| **Patient** | `patient@upasham.com` | `Pass@123456` | Rashidul Islam (Self + Daughter Fatima + Mother Amena) |
| **Doctor (Medicine)** | `dr.farhana@upasham.com` | `Pass@123456` | Dr. Farhana Rahman (BMDC A-84920, DMC) |
| **Doctor (Pediatrics)** | `dr.tanvir@upasham.com` | `Pass@123456` | Dr. Tanvir Ahmed (BMDC A-72195, BIRDEM) |
| **Operations** | `ops@upasham.com` | `Pass@123456` | Dhaka Operations Center Officer |
| **Finance** | `finance@upasham.com` | `Pass@123456` | Accounts & Reconciliation Manager |
| **Admin** | `admin@upasham.com` | `Pass@123456` | System Super-Administrator |

*Note: The top sticky "LIVE WORKSPACE" toolbar allows switching between all these roles in 1 click.*

---

## 6. Backup & Recovery Runbook

### Database Backup
```powershell
& "C:\Program Files\PostgreSQL\17\bin\pg_dump.exe" -h 127.0.0.1 -p 5433 -U postgres -d upasham_db -F c -b -v -f "upasham_db_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss').dump"
```

### Database Restore
```powershell
& "C:\Program Files\PostgreSQL\17\bin\pg_restore.exe" -h 127.0.0.1 -p 5433 -U postgres -d upasham_db -v -c "upasham_db_backup.dump"
```
