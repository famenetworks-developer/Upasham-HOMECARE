import React, { useState } from 'react';
import { PatientProfile, Booking, Prescription } from '../types';
import { User, Users, Calendar, FileText, Video, Download, Plus, AlertCircle, Clock, CheckCircle, X } from 'lucide-react';
import { Language, translations } from '../utils/translations';
import { downloadAuthenticatedPDF } from '../utils/download';

interface PatientPortalProps {
  profiles: PatientProfile[];
  bookings: Booking[];
  prescriptions: Prescription[];
  onRefresh: () => void;
  onJoinVideo: (roomId: string) => void;
  onOpenBooking: () => void;
  lang: Language;
}

export const PatientPortal: React.FC<PatientPortalProps> = ({
  profiles,
  bookings,
  prescriptions,
  onRefresh,
  onJoinVideo,
  onOpenBooking,
  lang,
}) => {
  const t = translations[lang];
  const [activeTab, setActiveTab] = useState<'appointments' | 'family' | 'prescriptions'>('appointments');
  const [showAddProfileModal, setShowAddProfileModal] = useState<boolean>(false);

  // New Profile Form
  const [fullName, setFullName] = useState<string>('');
  const [relationship, setRelationship] = useState<string>('child');
  const [gender, setGender] = useState<string>('female');
  const [age, setAge] = useState<number>(5);
  const [bloodGroup, setBloodGroup] = useState<string>('B+');
  const [knownAllergies, setKnownAllergies] = useState<string>('');
  const [chronicConditions, setChronicConditions] = useState<string>('');
  const [notes, setNotes] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const handleAddProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fullName) return;

    setIsSubmitting(true);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/patient/profiles', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          full_name: fullName,
          relationship,
          gender,
          age: Number(age),
          blood_group: bloodGroup,
          known_allergies: knownAllergies || 'None',
          chronic_conditions: chronicConditions || 'None',
          notes,
        }),
      });

      if (res.ok) {
        setShowAddProfileModal(false);
        setFullName('');
        setKnownAllergies('');
        setChronicConditions('');
        onRefresh();
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'Paid/Awaiting Assignment':
        return <span className="badge badge-yellow"><Clock size={11} /> {t.status_awaiting_assignment}</span>;
      case 'Assigned':
        return <span className="badge badge-blue"><User size={11} /> {t.status_assigned}</span>;
      case 'Confirmed':
        return <span className="badge badge-green"><CheckCircle size={11} /> {t.status_confirmed}</span>;
      case 'In Progress':
        return <span className="badge badge-lime animate-pulse-live"><Video size={11} /> {t.status_in_progress}</span>;
      case 'Completed':
        return <span className="badge badge-gray"><CheckCircle size={11} /> {t.status_completed}</span>;
      case 'Refunded':
        return <span className="badge badge-red">{t.status_refunded}</span>;
      default:
        return <span className="badge badge-gray">{status}</span>;
    }
  };

  return (
    <div style={{ padding: '36px 0', backgroundColor: '#F7FBFC', minHeight: '80vh' }}>
      <div className="container">
        {/* Portal Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <span style={{ fontSize: '12px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
              PATIENT DASHBOARD
            </span>
            <h1 style={{ fontSize: '28px', fontWeight: 800, color: '#2A3F68' }}>
              {t.portal_patient}
            </h1>
          </div>

          <div style={{ display: 'flex', gap: '10px' }}>
            <button
              onClick={() => setShowAddProfileModal(true)}
              className="btn btn-outline"
              style={{ fontSize: '13px' }}
            >
              <Plus size={15} />
              {t.add_family_btn}
            </button>
            <button
              onClick={onOpenBooking}
              className="btn btn-lime"
              style={{ fontSize: '13px', fontWeight: 800 }}
            >
              + নতুন ডাক্তার পরামর্শ বুক করুন
            </button>
          </div>
        </div>

        {/* Tab Buttons */}
        <div style={{ display: 'flex', gap: '8px', borderBottom: '1px solid rgba(15,10,1,0.1)', marginBottom: '28px' }}>
          <button
            onClick={() => setActiveTab('appointments')}
            style={{
              padding: '10px 18px',
              border: 'none',
              background: 'none',
              borderBottom: activeTab === 'appointments' ? '2px solid #33C2DF' : '2px solid transparent',
              color: activeTab === 'appointments' ? '#33C2DF' : '#64748B',
              fontWeight: activeTab === 'appointments' ? 700 : 500,
              fontSize: '14.5px',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
            }}
          >
            <Calendar size={16} />
            <span>{t.nav_appointments} ({bookings.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('family')}
            style={{
              padding: '10px 18px',
              border: 'none',
              background: 'none',
              borderBottom: activeTab === 'family' ? '2px solid #33C2DF' : '2px solid transparent',
              color: activeTab === 'family' ? '#33C2DF' : '#64748B',
              fontWeight: activeTab === 'family' ? 700 : 500,
              fontSize: '14.5px',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
            }}
          >
            <Users size={16} />
            <span>{t.family_manager_title} ({profiles.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('prescriptions')}
            style={{
              padding: '10px 18px',
              border: 'none',
              background: 'none',
              borderBottom: activeTab === 'prescriptions' ? '2px solid #33C2DF' : '2px solid transparent',
              color: activeTab === 'prescriptions' ? '#33C2DF' : '#64748B',
              fontWeight: activeTab === 'prescriptions' ? 700 : 500,
              fontSize: '14.5px',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
            }}
          >
            <FileText size={16} />
            <span>{t.nav_prescriptions} ({prescriptions.length})</span>
          </button>
        </div>

        {/* TAB 1: APPOINTMENTS */}
        {activeTab === 'appointments' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {bookings.length === 0 ? (
              <div style={{ backgroundColor: '#FFF', padding: '40px', borderRadius: '12px', textAlign: 'center' }}>
                <p style={{ color: '#64748B' }}>No appointments booked yet.</p>
              </div>
            ) : (
              bookings.map((b) => (
                <div
                  key={b.id}
                  className="card-safeer"
                  style={{
                    padding: '24px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    flexWrap: 'wrap',
                    gap: '16px',
                  }}
                >
                  <div style={{ flex: 1, minWidth: '280px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '8px' }}>
                      <span style={{ fontSize: '13px', fontWeight: 800, color: '#2A3F68' }}>{b.booking_ref}</span>
                      {getStatusBadge(b.status)}
                      <span style={{ fontSize: '12px', color: '#64748B' }}>
                        Paid ৳{b.quoted_amount} ({b.service_type === 'general' ? 'General Triage' : 'Specialist'})
                      </span>
                    </div>

                    <h3 style={{ fontSize: '17px', fontWeight: 800, color: '#2A3F68', marginBottom: '4px' }}>
                      {b.doctor_name ? b.doctor_name : 'Upasham On-Duty Physician (Dhaka Center)'}
                    </h3>

                    <p style={{ fontSize: '13.5px', color: '#475569', marginBottom: '4px' }}>
                      <strong>Patient:</strong> {b.patient_name} ({b.relationship})
                    </p>

                    <p style={{ fontSize: '13px', color: '#64748B' }}>
                      <strong>Complaint:</strong> {b.chief_complaint}
                    </p>

                    {b.scheduled_at && (
                      <p style={{ fontSize: '12.5px', color: '#33C2DF', fontWeight: 600, marginTop: '4px' }}>
                        Scheduled Time: {new Date(b.scheduled_at).toLocaleString('en-US', { timeZone: 'Asia/Dhaka' })}
                      </p>
                    )}
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    {(b.status === 'Confirmed' || b.status === 'In Progress') && b.video_room_id && (
                      <button
                        onClick={() => onJoinVideo(b.video_room_id!)}
                        className="btn btn-lime"
                        style={{ padding: '10px 20px', fontWeight: 800 }}
                      >
                        <Video size={16} />
                        {t.join_call_btn}
                      </button>
                    )}

                    {b.status === 'Paid/Awaiting Assignment' && (
                      <div
                        style={{
                          backgroundColor: '#FEF3C7',
                          color: '#92400E',
                          padding: '8px 14px',
                          borderRadius: '8px',
                          fontSize: '12px',
                          fontWeight: 600,
                          maxWidth: '240px',
                        }}
                      >
                        ঢাকা অপারেশন্স সেন্টার ডাক্তার নির্ধারণ করছে।
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {/* TAB 2: FAMILY PROFILES */}
        {activeTab === 'family' && (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '20px' }}>
            {profiles.map((prof) => (
              <div key={prof.id} className="card-safeer" style={{ padding: '24px' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '14px' }}>
                  <span
                    style={{
                      backgroundColor: prof.relationship === 'self' ? '#D3F8FF' : '#F6E8F3',
                      color: prof.relationship === 'self' ? '#168CA5' : '#79366C',
                      padding: '4px 10px',
                      borderRadius: '999px',
                      fontSize: '11px',
                      fontWeight: 800,
                      textTransform: 'uppercase',
                    }}
                  >
                    {prof.relationship}
                  </span>
                  <span style={{ fontSize: '13px', fontWeight: 700, color: '#2A3F68' }}>
                    Blood: <strong style={{ color: '#DC2626' }}>{prof.blood_group}</strong>
                  </span>
                </div>

                <h3 style={{ fontSize: '18px', fontWeight: 800, color: '#2A3F68', marginBottom: '4px' }}>
                  {prof.full_name}
                </h3>
                <p style={{ fontSize: '13px', color: '#64748B', marginBottom: '14px' }}>
                  {prof.age} Years Old • Gender: {prof.gender}
                </p>

                <div style={{ borderTop: '1px solid rgba(42,63,104,0.08)', paddingTop: '12px', fontSize: '12.5px', color: '#475569' }}>
                  <p style={{ marginBottom: '4px' }}>
                    <strong>Known Allergies:</strong>{' '}
                    <span style={{ color: prof.known_allergies !== 'None' ? '#DC2626' : '#64748B' }}>
                      {prof.known_allergies}
                    </span>
                  </p>
                  <p style={{ marginBottom: '4px' }}>
                    <strong>Chronic Conditions:</strong> {prof.chronic_conditions}
                  </p>
                  {prof.notes && (
                    <p style={{ color: '#64748B', fontStyle: 'italic', marginTop: '6px' }}>
                      Notes: {prof.notes}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* TAB 3: PRESCRIPTIONS */}
        {activeTab === 'prescriptions' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            {prescriptions.length === 0 ? (
              <div style={{ backgroundColor: '#FFF', padding: '40px', borderRadius: '12px', textAlign: 'center' }}>
                <p style={{ color: '#64748B' }}>No prescriptions issued yet.</p>
              </div>
            ) : (
              prescriptions.map((rx) => (
                <div key={rx.id} className="card-safeer" style={{ padding: '28px' }}>
                  <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '16px', flexWrap: 'wrap', gap: '12px' }}>
                    <div>
                      <span className="badge badge-green" style={{ marginBottom: '6px' }}>
                        DIGITAL PRESCRIPTION {rx.prescription_ref}
                      </span>
                      <h3 style={{ fontSize: '18px', fontWeight: 800, color: '#2A3F68' }}>
                        {rx.doctor_name}
                      </h3>
                      <p style={{ fontSize: '13px', color: '#64748B' }}>
                        BMDC: {rx.doctor_bmdc} • {rx.doctor_degrees} • {rx.doctor_workplace}
                      </p>
                      <p style={{ fontSize: '13px', color: '#2A3F68', fontWeight: 600, marginTop: '4px' }}>
                        Patient: {rx.patient_name} ({rx.patient_age} yrs, {rx.patient_gender})
                      </p>
                    </div>

                    <button
                      type="button"
                      onClick={() => downloadAuthenticatedPDF(rx.id).catch(err => alert(err.message))}
                      className="btn btn-lime"
                      style={{ padding: '10px 18px', fontSize: '13px', fontWeight: 800 }}
                    >
                      <Download size={15} />
                      {t.download_pdf_btn}
                    </button>
                  </div>

                  <div style={{ backgroundColor: '#EEF7F8', borderRadius: '10px', padding: '14px', marginBottom: '16px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 700, color: '#33C2DF', textTransform: 'uppercase' }}>
                      Clinical Assessment / Diagnosis
                    </span>
                    <p style={{ fontSize: '14.5px', fontWeight: 600, color: '#2A3F68', marginTop: '2px' }}>
                      {rx.diagnosis}
                    </p>
                  </div>

                  {/* Medicines Table */}
                  {rx.medicines && rx.medicines.length > 0 && (
                    <div style={{ marginBottom: '16px' }}>
                      <h4 style={{ fontSize: '13px', fontWeight: 800, color: '#9B468A', textTransform: 'uppercase', marginBottom: '8px' }}>
                        Rx Prescribed Medicines
                      </h4>
                      <div style={{ border: '1px solid #DCEAEC', borderRadius: '8px', overflow: 'hidden' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
                          <thead style={{ backgroundColor: '#EEF7F8', color: '#334155' }}>
                            <tr>
                              <th style={{ padding: '8px 12px' }}>#</th>
                              <th style={{ padding: '8px 12px' }}>Medicine</th>
                              <th style={{ padding: '8px 12px' }}>Dose</th>
                              <th style={{ padding: '8px 12px' }}>Frequency</th>
                              <th style={{ padding: '8px 12px' }}>Duration & Advice</th>
                            </tr>
                          </thead>
                          <tbody>
                            {rx.medicines.map((m, idx) => (
                              <tr key={idx} style={{ borderTop: '1px solid #DCEAEC' }}>
                                <td style={{ padding: '8px 12px', fontWeight: 600 }}>{idx + 1}</td>
                                <td style={{ padding: '8px 12px', fontWeight: 700, color: '#2A3F68' }}>
                                  {m.brand_name} {m.strength && `(${m.strength})`}
                                  {m.generic_name && (
                                    <div style={{ fontSize: '11px', color: '#64748B', fontWeight: 400 }}>
                                      {m.generic_name}
                                    </div>
                                  )}
                                </td>
                                <td style={{ padding: '8px 12px' }}>{m.dose}</td>
                                <td style={{ padding: '8px 12px' }}>{m.frequency}</td>
                                <td style={{ padding: '8px 12px', color: '#475569' }}>
                                  {m.duration} • {m.instructions}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Advice & Warnings */}
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '12px', fontSize: '12.5px' }}>
                    {rx.advice && (
                      <div style={{ backgroundColor: '#F0FDF4', padding: '12px', borderRadius: '8px', border: '1px solid #BBF7D0' }}>
                        <strong style={{ color: '#166534' }}>Doctor's Advice:</strong>
                        <p style={{ color: '#14532D', marginTop: '2px' }}>{rx.advice}</p>
                      </div>
                    )}
                    {rx.warnings && (
                      <div style={{ backgroundColor: '#FEF2F2', padding: '12px', borderRadius: '8px', border: '1px solid #FECACA' }}>
                        <strong style={{ color: '#991B1B' }}>Emergency Red Flags:</strong>
                        <p style={{ color: '#7F1D1D', marginTop: '2px' }}>{rx.warnings}</p>
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {/* Modal: Add Family Member */}
        {showAddProfileModal && (
          <div
            style={{
              position: 'fixed',
              inset: 0,
              backgroundColor: 'rgba(42,63,104,0.62)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              zIndex: 300,
              padding: '20px',
            }}
          >
            <div
              className="card-safeer animate-fade-in"
              style={{ width: '100%', maxWidth: '520px', padding: '28px', backgroundColor: '#FFF' }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '18px' }}>
                <h3 style={{ fontSize: '18px', fontWeight: 800 }}>{t.family_add_title}</h3>
                <button
                  onClick={() => setShowAddProfileModal(false)}
                  style={{ border: 'none', background: 'none', cursor: 'pointer' }}
                >
                  <X size={18} />
                </button>
              </div>

              <form onSubmit={handleAddProfile} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                <div>
                  <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                    {t.name_label}
                  </label>
                  <input
                    type="text"
                    required
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                  />
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                  <div>
                    <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                      {t.relationship_label}
                    </label>
                    <select
                      value={relationship}
                      onChange={(e) => setRelationship(e.target.value)}
                      style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', backgroundColor: '#FFF' }}
                    >
                      <option value="child">Child (সন্তান)</option>
                      <option value="parent">Parent (পিতামাতা)</option>
                      <option value="spouse">Spouse (জীবনসঙ্গী)</option>
                      <option value="other">Other (অন্যান্য)</option>
                    </select>
                  </div>

                  <div>
                    <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                      {t.age_label}
                    </label>
                    <input
                      type="number"
                      required
                      min={1}
                      max={120}
                      value={age}
                      onChange={(e) => setAge(Number(e.target.value))}
                      style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                    />
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                  <div>
                    <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                      Gender
                    </label>
                    <select
                      value={gender}
                      onChange={(e) => setGender(e.target.value)}
                      style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', backgroundColor: '#FFF' }}
                    >
                      <option value="female">Female</option>
                      <option value="male">Male</option>
                      <option value="other">Other</option>
                    </select>
                  </div>

                  <div>
                    <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                      {t.blood_group_label}
                    </label>
                    <select
                      value={bloodGroup}
                      onChange={(e) => setBloodGroup(e.target.value)}
                      style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', backgroundColor: '#FFF' }}
                    >
                      <option value="A+">A+</option>
                      <option value="A-">A-</option>
                      <option value="B+">B+</option>
                      <option value="B-">B-</option>
                      <option value="O+">O+</option>
                      <option value="O-">O-</option>
                      <option value="AB+">AB+</option>
                      <option value="AB-">AB-</option>
                      <option value="Unknown">Unknown</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                    {t.allergies_label}
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. Dust, Penicillin, Seafood"
                    value={knownAllergies}
                    onChange={(e) => setKnownAllergies(e.target.value)}
                    style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                  />
                </div>

                <div>
                  <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                    {t.chronic_label}
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. Asthma, Diabetes, Hypertension"
                    value={chronicConditions}
                    onChange={(e) => setChronicConditions(e.target.value)}
                    style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                  />
                </div>

                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="btn btn-lime"
                  style={{ width: '100%', padding: '12px', fontWeight: 800, marginTop: '8px' }}
                >
                  {isSubmitting ? 'Saving Profile...' : 'Save Profile'}
                </button>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

