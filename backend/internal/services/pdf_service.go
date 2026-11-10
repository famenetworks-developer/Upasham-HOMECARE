package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"upasham-backend/internal/models"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GeneratePrescriptionPDF(rx *models.Prescription) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 12, 15)
	pdf.AddPage()

	// 1. Top Brand Banner
	pdf.SetFillColor(21, 128, 61) // Upasham Green
	pdf.Rect(0, 0, 210, 8, "F")

	// Header: Upasham Plus Connected Care
	pdf.SetY(14)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(15, 10, 1) // Safeer Dark
	pdf.CellFormat(100, 8, "UPASHAM PLUS", "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(21, 128, 61)
	pdf.CellFormat(80, 8, "TELEHEALTH PRESCRIPTION", "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(100, 5, "Connected Care Platform Dhaka | www.upasham.com", "", 0, "L", false, 0, "")
	pdf.CellFormat(80, 5, fmt.Sprintf("Prescription ID: %s", rx.PrescriptionRef), "", 1, "R", false, 0, "")

	// Horizontal rule
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.4)
	pdf.Line(15, 30, 195, 30)

	// 2. Doctor Info & Patient Info Box
	pdf.SetY(34)
	// Left: Doctor Info
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(15, 10, 1)
	pdf.CellFormat(90, 6, rx.DoctorName, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(71, 85, 105)
	pdf.CellFormat(90, 5, rx.DoctorDegrees, "", 1, "L", false, 0, "")
	pdf.CellFormat(90, 5, fmt.Sprintf("BMDC Reg No: %s", rx.DoctorBMDC), "", 1, "L", false, 0, "")
	pdf.CellFormat(90, 5, rx.DoctorWorkplace, "", 1, "L", false, 0, "")

	// Right: Patient Info Box
	pdf.SetXY(110, 34)
	pdf.SetFillColor(248, 250, 252)
	pdf.Rect(110, 34, 85, 26, "F")
	pdf.SetXY(112, 36)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(15, 10, 1)
	pdf.CellFormat(80, 4, fmt.Sprintf("Patient: %s", rx.PatientName), "", 1, "L", false, 0, "")

	pdf.SetX(112)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.CellFormat(80, 4, fmt.Sprintf("Age: %d Years  |  Gender: %s", rx.PatientAge, rx.PatientGender), "", 1, "L", false, 0, "")
	pdf.SetX(112)
	pdf.CellFormat(80, 4, fmt.Sprintf("Booking Ref: %s", rx.BookingRef), "", 1, "L", false, 0, "")
	pdf.SetX(112)
	pdf.CellFormat(80, 4, fmt.Sprintf("Date: %s", rx.CreatedAt.Format("02 Jan 2006, 03:04 PM")), "", 1, "L", false, 0, "")

	// Line separator
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(15, 64, 195, 64)

	// 3. Clinical Assessment & Diagnosis
	pdf.SetY(68)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(21, 128, 61)
	pdf.CellFormat(180, 5, "PROVISIONAL DIAGNOSIS / CLINICAL ASSESSMENT:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(15, 10, 1)
	pdf.MultiCell(180, 5, rx.Diagnosis, "", "L", false)

	pdf.Ln(4)

	// 4. Rx Symbol & Medicines Section
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(2, 132, 199) // Stethoscope blue
	pdf.CellFormat(15, 8, "Rx", "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(15, 10, 1)
	pdf.CellFormat(165, 8, "PRESCRIBED MEDICATIONS", "", 1, "L", false, 0, "")

	// Table Header
	pdf.SetFillColor(241, 245, 249)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.SetTextColor(30, 41, 59)
	pdf.CellFormat(10, 6, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(65, 6, "Medicine / Generic / Strength", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 6, "Dose / Route", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 6, "Frequency", "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 6, "Duration & Instructions", "1", 1, "L", true, 0, "")

	// Table Rows
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(15, 10, 1)
	for i, m := range rx.Medicines {
		medDesc := fmt.Sprintf("%s (%s)", m.BrandName, m.Strength)
		if m.GenericName != "" {
			medDesc = fmt.Sprintf("%s\n[%s]", medDesc, m.GenericName)
		}
		doseDesc := fmt.Sprintf("%s\n%s", m.Dose, m.Route)
		durationDesc := fmt.Sprintf("%s\n%s", m.Duration, m.Instructions)

		pdf.CellFormat(10, 10, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(65, 10, medDesc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, doseDesc, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 10, m.Frequency, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 10, durationDesc, "1", 1, "L", false, 0, "")
	}

	pdf.Ln(4)

	// 5. Recommended Diagnostic Tests (if any)
	if len(rx.Tests) > 0 {
		pdf.SetFont("Arial", "B", 9.5)
		pdf.SetTextColor(21, 128, 61)
		pdf.CellFormat(180, 5, "DIAGNOSTIC INVESTIGATIONS:", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8.5)
		pdf.SetTextColor(15, 10, 1)
		for _, t := range rx.Tests {
			pdf.CellFormat(180, 4, fmt.Sprintf("• %s (%s)", t.TestName, t.Instructions), "", 1, "L", false, 0, "")
		}
		pdf.Ln(3)
	}

	// 6. Advice & Instructions
	if rx.Advice != "" {
		pdf.SetFont("Arial", "B", 9.5)
		pdf.SetTextColor(21, 128, 61)
		pdf.CellFormat(180, 5, "SPECIAL ADVICE & LIFESTYLE GUIDANCE:", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8.5)
		pdf.SetTextColor(51, 65, 85)
		pdf.MultiCell(180, 4.5, rx.Advice, "", "L", false)
		pdf.Ln(3)
	}

	// 7. Follow-up & Red Flags / Emergency Warnings
	if rx.FollowUpDate != "" {
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(15, 10, 1)
		pdf.CellFormat(180, 5, fmt.Sprintf("Next Follow-Up: %s", rx.FollowUpDate), "", 1, "L", false, 0, "")
	}

	if rx.Warnings != "" {
		pdf.SetFillColor(254, 242, 242) // Light red warning box
		pdf.Rect(15, pdf.GetY(), 180, 14, "F")
		pdf.SetFont("Arial", "B", 8)
		pdf.SetTextColor(185, 28, 28)
		pdf.CellFormat(180, 4, "EMERGENCY WARNING & CAUTION:", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 7.5)
		pdf.MultiCell(180, 3.5, rx.Warnings, "", "L", false)
	}

	// 8. Footer with Verification Hash & Digital Signature Notice
	pdf.SetY(268)
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(15, 267, 195, 267)

	pdf.SetFont("Arial", "", 7.5)
	pdf.SetTextColor(148, 163, 184)
	pdf.CellFormat(120, 4, fmt.Sprintf("Digitally generated on %s | Security Hash: %s", time.Now().Format("2006-01-02 15:04:05"), rx.VerificationHash), "", 0, "L", false, 0, "")
	pdf.CellFormat(60, 4, "BMDC Verified Digital Prescription", "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("pdf generation failed: %w", err)
	}

	return buf.Bytes(), nil
}
