# Admin and Operations User Guide - Upasham Plus

## 1. Operations Portal Guide
**URL:** Accessible via the top workspace bar -> **Operations (Dhaka)**

### 1.1 Managing the Paid Request Queue
1. When a patient books a consultation and completes payment through bKash or SSLCOMMERZ, the booking transitions to **`Paid/Awaiting Assignment`**.
2. Operations officers monitor this queue.
3. Click **"Assign Doctor"** on the triage item. Select the available on-duty licensed doctor from the verified dropdown list and confirm.
4. The booking transitions to **`Assigned`**.

### 1.2 Confirming and Provisioning Consultation Slot
1. Click **"Confirm Slot"** on the assigned triage item.
2. Select the scheduled consultation time.
3. Upon confirmation:
   - The booking transitions to **`Confirmed`**.
   - A cryptographically random Jitsi Video Room is generated.
   - Both the patient and doctor portals display the **"Join Video Call"** button.

### 1.3 Rescheduling and Cancellations
- To cancel or issue a refund, click **"Cancel"** and provide operational rationale.
- The status updates to **`Cancelled`** and triggers a refund record in the finance ledger.

---

## 2. Doctor Portal Guide
**URL:** Accessible via the top workspace bar -> **Doctor (Dr. Farhana / Dr. Tanvir)**

### 2.1 Attending Video Call
1. In the **Assigned Patient Queue**, locate the confirmed booking.
2. Click **"Launch Video Consultation"** to enter the protected in-app Jitsi telehealth room.
3. Telehealth attendance timestamps are automatically logged.

### 2.2 Issuing Official Electronic Prescription
1. Fill in the **Provisional Diagnosis** (e.g. *Seasonal Rhinitis / Viral Fever*).
2. Add multiple medicine rows using the **Rx Medications List**:
   - Brand name (e.g. *Napa 500mg*, *Fexo 120mg*)
   - Generic name
   - Dose (e.g. *1+0+1*)
   - Frequency & duration
   - Meal instructions (e.g. *After meals with water*)
3. Enter diagnostic test orders if required (e.g. *CBC with ESR*).
4. Provide special patient advice and emergency red-flag warnings.
5. Enter any internal clinical notes in the confidential physician note box (hidden from the patient PDF).
6. Click **"Finalize & Issue Prescription (PDF)"**.
7. The system:
   - Marks the consultation **`Completed`**.
   - Computes a SHA-256 verification hash.
   - Generates the official branded prescription PDF available for instant patient download.

---

## 3. Finance Portal Guide
**URL:** Accessible via the top workspace bar -> **Finance & Gateway**

1. **Dashboard Metrics:** Track gross platform volume, bKash vs SSLCOMMERZ split, and refund volume.
2. **Transactions Ledger:** View individual invoices, gateway transaction IDs, and settlement dates.
3. **Refund Authorization:** Click **"Refund"** on any paid item, verify refund amount and operational reason, and click **"Approve Refund"**.
