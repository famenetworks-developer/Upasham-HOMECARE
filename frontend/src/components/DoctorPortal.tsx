import React, { useState } from 'react';
import { MedicineLine, TestOrder } from '../types';
import { Stethoscope, Video, FileText, Plus, Trash2, CheckCircle, Clock, AlertTriangle, User, Download } from 'lucide-react';
import { Language, translations } from '../utils/translations';
import { downloadAuthenticatedPDF } from '../utils/download';
import confetti from 'canvas-confetti';

interface DoctorPortalProps {
  assignments: any[];
  onRefresh: () => void;
  onJoinVideo: (roomId: string) => void;
  lang: Language;
}

export const DoctorPortal: React.FC<DoctorPortalProps> = ({
  assignments,
  onRefresh,
  onJoinVideo,
  lang,
}) => {
  const t = translations[lang];
  const [activeBooking, setActiveBooking] = useState<any>(null);

  // Prescription Form State
  const [diagnosis, setDiagnosis] = useState<string>('Acute Upper Respiratory Tract Infection (Seasonal Flu)');
  const [subjective, setSubjective] = useState<string>('Patient reports 2 days of mild fever, dry cough, rhinorrhea and headache.');
  const [objective, setObjective] = useState<string>('Temperature 99.8°F, throat mildly erythematous, chest bilaterally clear.');
  const [privateInternal, setPrivateInternal] = useState<string>('Advised to report back if fever exceeds 102°F or if wheezing appears.');
  const [advice, setAdvice] = useState<string>('Drink warm fluids. Rest for 3 days. Steam inhalation twice daily.');
  const [followUpDate, setFollowUpDate] = useState<string>('After 5 days if symptoms persist');
  const [warnings, setWarnings] = useState<string>('If breathlessness, chest pain or continuous high fever occurs, seek emergency care immediately.');

  const [medicines, setMedicines] = useState<MedicineLine[]>([
    {
      brand_name: 'Napa 500mg',
      generic_name: 'Paracetamol',
      strength: '500mg',
      dosage_form: 'Tablet',
      dose: '1+0+1',
      route: 'Oral',
      frequency: 'After meals',
      duration: '3 days',
      instructions: 'Take when fever is above 100°F',
    },
    {
      brand_name: 'Fexo 120mg',
      generic_name: 'Fexofenadine HCl',
      strength: '120mg',
      dosage_form: 'Tablet',
      dose: '0+0+1',
      route: 'Oral',
      frequency: 'Nightly',
      duration: '7 days',
      instructions: 'Take after dinner',
    },
  ]);

  const [tests, setTests] = useState<TestOrder[]>([
    { test_name: 'Complete Blood Count (CBC) with ESR', instructions: 'Routine blood test if fever persists' },
  ]);

  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [completedRxRef, setCompletedRxRef] = useState<string>('');

  const addMedicineRow = () => {
    setMedicines([
      ...medicines,
      {
        brand_name: '',
        generic_name: '',
        strength: '',
        dosage_form: 'Tablet',
        dose: '1+0+1',
        route: 'Oral',
        frequency: 'Twice daily',
        duration: '5 days',
        instructions: 'After meal',
      },
    ]);
  };

  const removeMedicineRow = (index: number) => {
    setMedicines(medicines.filter((_, idx) => idx !== index));
  };

  const updateMedicine = (index: number, field: keyof MedicineLine, value: string) => {
    const updated = [...medicines];
    updated[index] = { ...updated[index], [field]: value };
    setMedicines(updated);
  };

  const addTestRow = () => {
    setTests([...tests, { test_name: '', instructions: '' }]);
  };

  const removeTestRow = (index: number) => {
    setTests(tests.filter((_, idx) => idx !== index));
  };

  const handleIssuePrescription = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeBooking || !diagnosis) return;

    setIsSubmitting(true);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/doctor/clinical/save', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          booking_id: activeBooking.booking.id,
          diagnosis,
          subjective_notes: subjective,
          objective_notes: objective,
          private_internal: privateInternal,
          medicines: medicines.filter((m) => m.brand_name.trim() !== ''),
          tests: tests.filter((t) => t.test_name.trim() !== ''),
          advice,
          follow_up_date: followUpDate,
          warnings,
        }),
      });

      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to save clinical encounter');

      confetti({ particleCount: 70, spread: 50 });
      setCompletedRxRef(data.prescription_ref);
      onRefresh();
    } catch (err: any) {
      alert(err.message || 'Error saving prescription');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div style={{ padding: '36px 0', backgroundColor: '#F7FBFC', minHeight: '80vh' }}>
      <div className="container">
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <span style={{ fontSize: '12px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
              PHYSICIAN CONSOLE
            </span>
            <h1 style={{ fontSize: '28px', fontWeight: 800, color: '#2A3F68' }}>
              {t.portal_doctor}
            </h1>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span className="badge badge-green" style={{ fontSize: '13px', padding: '6px 14px' }}>
              <CheckCircle size={14} /> BMDC Verified Tele-Practitioner
            </span>
          </div>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: activeBooking ? '360px 1fr' : '1fr', gap: '24px' }}>
          {/* Left Column: Assigned Consultations Queue */}
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '14px' }}>
              <h3 style={{ fontSize: '16px', fontWeight: 800, color: '#2A3F68' }}>
                Assigned Patient Queue ({assignments.length})
              </h3>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {assignments.length === 0 ? (
                <div style={{ backgroundColor: '#FFF', padding: '32px', borderRadius: '12px', textAlign: 'center', color: '#64748B' }}>
                  No active assignments currently.
                </div>
              ) : (
                assignments.map((item) => {
                  const b = item.booking;
                  const p = item.patient;
                  const isSelected = activeBooking?.booking?.id === b.id;

                  return (
                    <div
                      key={b.id}
                      onClick={() => {
                        setActiveBooking(item);
                        setCompletedRxRef('');
                      }}
                      className="card-safeer"
                      style={{
                        padding: '18px',
                        cursor: 'pointer',
                        borderColor: isSelected ? '#33C2DF' : 'rgba(42, 63, 104, 0.08)',
                        backgroundColor: isSelected ? '#F0FDF4' : '#FFFFFF',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                        <span style={{ fontSize: '12px', fontWeight: 800, color: '#2A3F68' }}>{b.booking_ref}</span>
                        <span
                          className={`badge ${
                            b.status === 'In Progress'
                              ? 'badge-lime'
                              : b.status === 'Confirmed'
                              ? 'badge-green'
                              : b.status === 'Completed'
                              ? 'badge-gray'
                              : 'badge-blue'
                          }`}
                          style={{ fontSize: '11px' }}
                        >
                          {b.status}
                        </span>
                      </div>

                      <h4 style={{ fontSize: '15px', fontWeight: 700, color: '#2A3F68', marginBottom: '2px' }}>
                        {p.name}
                      </h4>
                      <p style={{ fontSize: '12px', color: '#64748B', marginBottom: '8px' }}>
                        {p.age} Yrs • {p.gender} • Blood: {p.blood_group}
                      </p>

                      <p style={{ fontSize: '12.5px', color: '#334155', lineHeight: 1.4, marginBottom: '10px' }}>
                        <strong>Complaint:</strong> {b.chief_complaint}
                      </p>

                      {b.video_room_id && b.status !== 'Completed' && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            onJoinVideo(b.video_room_id);
                          }}
                          className="btn btn-lime"
                          style={{ width: '100%', padding: '8px', fontSize: '12px', fontWeight: 800 }}
                        >
                          <Video size={14} />
                          Launch Video Consultation
                        </button>
                      )}
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* Right Column: Active Patient Context & Prescription Builder */}
          {activeBooking && (
            <div className="card-safeer" style={{ padding: '28px', backgroundColor: '#FFFFFF' }}>
              {/* Patient Banner */}
              <div
                style={{
                  backgroundColor: '#EEF7F8',
                  borderRadius: '12px',
                  padding: '18px',
                  marginBottom: '24px',
                  border: '1px solid #DCEAEC',
                  display: 'flex',
                  justifyContent: 'space-between',
                  flexWrap: 'wrap',
                  gap: '12px',
                }}
              >
                <div>
                  <span style={{ fontSize: '11px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
                    Active Patient Record
                  </span>
                  <h3 style={{ fontSize: '18px', fontWeight: 800, color: '#2A3F68' }}>
                    {activeBooking.patient.name} ({activeBooking.patient.relationship})
                  </h3>
                  <p style={{ fontSize: '13px', color: '#475569' }}>
                    Age: {activeBooking.patient.age} • Gender: {activeBooking.patient.gender} • Blood: {activeBooking.patient.blood_group}
                  </p>
                </div>

                <div style={{ textAlign: 'right', fontSize: '12.5px' }}>
                  <p style={{ color: activeBooking.patient.known_allergies !== 'None' ? '#DC2626' : '#64748B' }}>
                    <strong>Allergies:</strong> {activeBooking.patient.known_allergies}
                  </p>
                  <p style={{ color: '#475569' }}>
                    <strong>Chronic History:</strong> {activeBooking.patient.chronic_conditions}
                  </p>
                </div>
              </div>

              {completedRxRef ? (
                <div style={{ textAlign: 'center', padding: '36px 0' }}>
                  <CheckCircle size={48} color="#33C2DF" style={{ margin: '0 auto 12px' }} />
                  <h3 style={{ fontSize: '20px', fontWeight: 800, marginBottom: '6px' }}>
                    Prescription Successfully Issued!
                  </h3>
                  <p style={{ fontSize: '14px', color: '#64748B', marginBottom: '20px' }}>
                    Reference: <strong>{completedRxRef}</strong> • Consultation marked Completed
                  </p>
                  <button
                    type="button"
                    onClick={() => downloadAuthenticatedPDF(completedRxRef).catch(err => alert(err.message))}
                    className="btn btn-lime"
                    style={{ fontWeight: 800, padding: '10px 24px' }}
                  >
                    <Download size={16} />
                    Download Generated Prescription PDF
                  </button>
                </div>
              ) : (
                <form onSubmit={handleIssuePrescription}>
                  <h3 style={{ fontSize: '17px', fontWeight: 800, color: '#2A3F68', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <Stethoscope size={18} color="#33C2DF" />
                    Electronic Prescription & Encounter Form
                  </h3>

                  {/* Diagnosis */}
                  <div style={{ marginBottom: '16px' }}>
                    <label style={{ fontSize: '12.5px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                      Provisional Diagnosis / Assessment: *
                    </label>
                    <input
                      type="text"
                      required
                      value={diagnosis}
                      onChange={(e) => setDiagnosis(e.target.value)}
                      style={{ width: '100%', padding: '10px 14px', borderRadius: '8px', border: '1px solid #C7D8DE', fontSize: '14px', fontWeight: 600 }}
                    />
                  </div>

                  {/* Subjective & Objective */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
                    <div>
                      <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                        Subjective (Chief Complaint & History):
                      </label>
                      <textarea
                        rows={2}
                        value={subjective}
                        onChange={(e) => setSubjective(e.target.value)}
                        style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', fontSize: '13px' }}
                      />
                    </div>
                    <div>
                      <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                        Objective (Clinical Findings & Vitals):
                      </label>
                      <textarea
                        rows={2}
                        value={objective}
                        onChange={(e) => setObjective(e.target.value)}
                        style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', fontSize: '13px' }}
                      />
                    </div>
                  </div>

                  {/* Medicines Table */}
                  <div style={{ marginBottom: '20px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
                      <span style={{ fontSize: '13px', fontWeight: 800, color: '#9B468A', textTransform: 'uppercase' }}>
                        Rx Medications List
                      </span>
                      <button
                        type="button"
                        onClick={addMedicineRow}
                        className="btn btn-outline"
                        style={{ padding: '4px 10px', fontSize: '12px' }}
                      >
                        <Plus size={13} /> Add Medicine
                      </button>
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                      {medicines.map((m, idx) => (
                        <div
                          key={idx}
                          style={{
                            display: 'grid',
                            gridTemplateColumns: '1.5fr 1fr 1fr 1fr 1fr 32px',
                            gap: '6px',
                            alignItems: 'center',
                            backgroundColor: '#EEF7F8',
                            padding: '8px 12px',
                            borderRadius: '8px',
                            border: '1px solid #DCEAEC',
                          }}
                        >
                          <input
                            type="text"
                            placeholder="Brand Name (e.g. Napa)"
                            value={m.brand_name}
                            onChange={(e) => updateMedicine(idx, 'brand_name', e.target.value)}
                            style={{ padding: '6px 8px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                          />
                          <input
                            type="text"
                            placeholder="Strength (500mg)"
                            value={m.strength || ''}
                            onChange={(e) => updateMedicine(idx, 'strength', e.target.value)}
                            style={{ padding: '6px 8px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                          />
                          <input
                            type="text"
                            placeholder="Dose (1+0+1)"
                            value={m.dose}
                            onChange={(e) => updateMedicine(idx, 'dose', e.target.value)}
                            style={{ padding: '6px 8px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                          />
                          <input
                            type="text"
                            placeholder="Duration (5 days)"
                            value={m.duration}
                            onChange={(e) => updateMedicine(idx, 'duration', e.target.value)}
                            style={{ padding: '6px 8px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                          />
                          <input
                            type="text"
                            placeholder="Instructions (After meal)"
                            value={m.instructions || ''}
                            onChange={(e) => updateMedicine(idx, 'instructions', e.target.value)}
                            style={{ padding: '6px 8px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                          />
                          <button
                            type="button"
                            onClick={() => removeMedicineRow(idx)}
                            style={{ border: 'none', background: 'none', color: '#EF4444', cursor: 'pointer' }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Diagnostic Tests */}
                  <div style={{ marginBottom: '18px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '6px' }}>
                      <span style={{ fontSize: '12.5px', fontWeight: 700, color: '#334155' }}>
                        Diagnostic Investigations / Tests:
                      </span>
                      <button
                        type="button"
                        onClick={addTestRow}
                        style={{ border: 'none', background: 'none', color: '#33C2DF', fontSize: '12px', fontWeight: 700, cursor: 'pointer' }}
                      >
                        + Add Test
                      </button>
                    </div>

                    {tests.map((t, idx) => (
                      <div key={idx} style={{ display: 'flex', gap: '8px', marginBottom: '6px' }}>
                        <input
                          type="text"
                          placeholder="Test Name (e.g. CBC, Serum Creatinine)"
                          value={t.test_name}
                          onChange={(e) => {
                            const updated = [...tests];
                            updated[idx].test_name = e.target.value;
                            setTests(updated);
                          }}
                          style={{ flex: 1, padding: '6px 10px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                        />
                        <input
                          type="text"
                          placeholder="Instructions (Fasting / Routine)"
                          value={t.instructions}
                          onChange={(e) => {
                            const updated = [...tests];
                            updated[idx].instructions = e.target.value;
                            setTests(updated);
                          }}
                          style={{ flex: 1, padding: '6px 10px', borderRadius: '6px', border: '1px solid #C7D8DE', fontSize: '12.5px' }}
                        />
                        <button
                          type="button"
                          onClick={() => removeTestRow(idx)}
                          style={{ border: 'none', background: 'none', color: '#EF4444', cursor: 'pointer' }}
                        >
                          <Trash2 size={15} />
                        </button>
                      </div>
                    ))}
                  </div>

                  {/* Advice & Warnings */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
                    <div>
                      <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                        Patient Advice & Lifestyle Guidance:
                      </label>
                      <textarea
                        rows={2}
                        value={advice}
                        onChange={(e) => setAdvice(e.target.value)}
                        style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', fontSize: '13px' }}
                      />
                    </div>
                    <div>
                      <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                        Emergency Red-Flags & Warnings:
                      </label>
                      <textarea
                        rows={2}
                        value={warnings}
                        onChange={(e) => setWarnings(e.target.value)}
                        style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', fontSize: '13px' }}
                      />
                    </div>
                  </div>

                  {/* Private Internal Clinical Notes */}
                  <div style={{ marginBottom: '24px' }}>
                    <label style={{ fontSize: '12px', fontWeight: 700, color: '#B45309', display: 'block', marginBottom: '4px' }}>
                      🔒 Internal Physician Notes (Confidential, hidden from patient PDF):
                    </label>
                    <input
                      type="text"
                      value={privateInternal}
                      onChange={(e) => setPrivateInternal(e.target.value)}
                      style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #FDE68A', backgroundColor: '#FFFDF5', fontSize: '13px' }}
                    />
                  </div>

                  {/* Submit Action */}
                  <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px' }}>
                    <button
                      type="button"
                      onClick={() => setActiveBooking(null)}
                      className="btn btn-outline"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="btn btn-lime"
                      style={{ padding: '12px 28px', fontSize: '15px', fontWeight: 800 }}
                    >
                      {isSubmitting ? 'Generating Official Prescription...' : 'Finalize & Issue Prescription (PDF)'}
                    </button>
                  </div>
                </form>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

