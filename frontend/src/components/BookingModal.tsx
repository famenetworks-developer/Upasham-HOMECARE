import React, { useState, useEffect } from 'react';
import { Doctor, PatientProfile } from '../types';
import { X, CheckCircle, ShieldAlert, CreditCard, ArrowRight, Loader2, User } from 'lucide-react';
import { Language, translations } from '../utils/translations';
import confetti from 'canvas-confetti';

interface BookingModalProps {
  isOpen: boolean;
  onClose: () => void;
  selectedDoctor: Doctor | null;
  serviceType: 'general' | 'specialist';
  patientProfiles: PatientProfile[];
  onBookingSuccess: (bookingId: string) => void;
  lang: Language;
}

export const BookingModal: React.FC<BookingModalProps> = ({
  isOpen,
  onClose,
  selectedDoctor,
  serviceType,
  patientProfiles,
  onBookingSuccess,
  lang,
}) => {
  const t = translations[lang];

  const [step, setStep] = useState<number>(1);
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');
  const [complaint, setComplaint] = useState<string>('');
  const [preferredTime, setPreferredTime] = useState<string>('Immediate (within 15 minutes)');
  const [consentGiven, setConsentGiven] = useState<boolean>(true);
  const [gateway, setGateway] = useState<'bkash' | 'sslcommerz'>('bkash');

  const [quotedAmount, setQuotedAmount] = useState<number>(serviceType === 'general' ? 20 : (selectedDoctor?.consultation_fee || 1000));
  const [loading, setLoading] = useState<boolean>(false);
  const [createdBookingId, setCreatedBookingId] = useState<string>('');
  const [errorMsg, setErrorMsg] = useState<string>('');

  useEffect(() => {
    if (!isOpen) return;
    if (patientProfiles.length > 0 && !selectedProfileId) {
      setSelectedProfileId(patientProfiles[0].id);
    }
  }, [isOpen, patientProfiles, selectedProfileId]);

  // Fetch official server quotation (Rule 11: never trust client-posted amount)
  useEffect(() => {
    if (!isOpen) return;
    const fetchQuote = async () => {
      try {
        const res = await fetch('/api/quotation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            service_type: serviceType,
            doctor_id: selectedDoctor?.id || null,
          }),
        });
        const data = await res.json();
        if (data.amount) {
          setQuotedAmount(data.amount);
        }
      } catch (e) {
        console.error('Error fetching quotation', e);
      }
    };
    fetchQuote();
  }, [isOpen, serviceType, selectedDoctor]);

  if (!isOpen) return null;

  const handleCreateBooking = async () => {
    if (!complaint.trim()) {
      setErrorMsg(lang === 'bn' ? 'অনুগ্রহ করে আপনার প্রধান স্বাস্থ্য সমস্যা লিখুন' : 'Please provide your chief health complaint');
      return;
    }
    if (!consentGiven) {
      setErrorMsg(lang === 'bn' ? 'টেলিমেডিসিন সেবা গ্রহণের সম্মতি প্রয়োজন' : 'Consent is required to proceed');
      return;
    }

    setLoading(true);
    setErrorMsg('');

    try {
      // Step A: Create Booking on server
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/patient/bookings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          patient_profile_id: selectedProfileId,
          service_type: serviceType,
          doctor_id: selectedDoctor ? selectedDoctor.id : null,
          specialty_id: selectedDoctor && selectedDoctor.specialties.length > 0 ? selectedDoctor.specialties[0] : null,
          chief_complaint: complaint,
          preferred_time: preferredTime,
          consent_given: consentGiven,
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Failed to create booking');
      }

      setCreatedBookingId(data.booking_id);
      setStep(2); // move to payment step
    } catch (err: any) {
      setErrorMsg(err.message || 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  const handleSimulatePayment = async () => {
    setLoading(true);
    setErrorMsg('');

    try {
      // Initiate payment
      const token = localStorage.getItem('upasham_token') || '';
      const initRes = await fetch('/api/patient/payments/initiate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          booking_id: createdBookingId,
          gateway: gateway,
          customer_phone: '+8801711223344',
          return_url: '/patient/callback',
        }),
      });
      const initData = await initRes.json();
      if (!initRes.ok) throw new Error(initData.error || 'Payment initiation failed');

      // Execute simulated verified callback
      const callbackRes = await fetch('/api/payments/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          invoice_ref: initData.invoice_ref,
          gateway: gateway,
          status: 'success',
          trx_id: initData.gateway_tx_id || 'BK-TEST-TX',
          raw_data: {
            status: 'success',
            trxID: initData.gateway_tx_id,
            amount: initData.amount,
          },
        }),
      });
      const cbData = await callbackRes.json();
      if (!callbackRes.ok) throw new Error(cbData.error || 'Callback verification failed');

      // Confetti celebration
      confetti({
        particleCount: 80,
        spread: 60,
        origin: { y: 0.6 },
      });

      setStep(3); // success state
      onBookingSuccess(createdBookingId);
    } catch (err: any) {
      setErrorMsg(err.message || 'Payment processing error');
    } finally {
      setLoading(false);
    }
  };

  const selectedPatient = patientProfiles.find((p) => p.id === selectedProfileId);

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(42,63,104,0.68)',
        backdropFilter: 'blur(4px)',
        zIndex: 250,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px',
      }}
      onClick={onClose}
    >
      <div
        className="card-safeer animate-fade-in"
        style={{
          width: '100%',
          maxWidth: '620px',
          maxHeight: '92vh',
          overflowY: 'auto',
          padding: '32px',
          backgroundColor: '#FFFFFF',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
          <div>
            <span style={{ fontSize: '12px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
              STEP {step} OF 3
            </span>
            <h2 style={{ fontSize: '22px', fontWeight: 800, color: '#2A3F68' }}>
              {step === 1 && (serviceType === 'general' ? 'জেনারেল মেডিকেল ট্রায়াজ বুকিং' : 'বিশেষজ্ঞ ডাক্তার সিরিয়াল বুকিং')}
              {step === 2 && 'পেমেন্ট গেটওয়ে নির্বাচন'}
              {step === 3 && 'পেমেন্ট সফল ও অপেক্ষমান তালিকা'}
            </h2>
          </div>

          <button
            onClick={onClose}
            style={{
              border: 'none',
              background: '#EEF7F8',
              borderRadius: '50%',
              width: 32,
              height: 32,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
            }}
          >
            <X size={16} />
          </button>
        </div>

        {errorMsg && (
          <div
            style={{
              backgroundColor: '#FEF2F2',
              border: '1px solid #FCA5A5',
              color: '#B91C1C',
              padding: '12px 16px',
              borderRadius: '8px',
              marginBottom: '16px',
              fontSize: '13px',
              fontWeight: 600,
            }}
          >
            {errorMsg}
          </div>
        )}

        {/* STEP 1: Patient Profile & Complaint */}
        {step === 1 && (
          <div>
            {/* Consultation Summary Banner */}
            <div
              style={{
                backgroundColor: serviceType === 'general' ? '#D3F8FF' : '#EEF7F8',
                border: serviceType === 'general' ? '1px solid #BBF7D0' : '1px solid #DCEAEC',
                borderRadius: '12px',
                padding: '16px',
                marginBottom: '20px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <div>
                <span style={{ fontSize: '11px', fontWeight: 700, textTransform: 'uppercase', color: '#168CA5' }}>
                  {serviceType === 'general' ? 'Upasham Instant Triage' : 'Specialist Video Consultation'}
                </span>
                <h4 style={{ fontSize: '16px', fontWeight: 800, color: '#2A3F68', marginTop: '2px' }}>
                  {selectedDoctor ? selectedDoctor.full_name : 'Upasham Available On-Duty Doctor (Dhaka)'}
                </h4>
                {selectedDoctor && (
                  <p style={{ fontSize: '12px', color: '#64748B' }}>{selectedDoctor.workplace}</p>
                )}
              </div>
              <div style={{ textAlign: 'right' }}>
                <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 600 }}>Quoted Price:</span>
                <div style={{ fontSize: '22px', fontWeight: 800, color: '#2A3F68' }}>
                  ৳{quotedAmount} <span style={{ fontSize: '11px', color: '#64748B' }}>BDT</span>
                </div>
              </div>
            </div>

            {/* Select Patient Profile */}
            <div style={{ marginBottom: '18px' }}>
              <label style={{ display: 'block', fontSize: '13px', fontWeight: 700, color: '#2A3F68', marginBottom: '8px' }}>
                {t.booking_step_patient}:
              </label>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: '8px' }}>
                {patientProfiles.map((prof) => {
                  const isSelected = selectedProfileId === prof.id;
                  return (
                    <div
                      key={prof.id}
                      onClick={() => setSelectedProfileId(prof.id)}
                      style={{
                        padding: '12px',
                        borderRadius: '10px',
                        border: isSelected ? '2px solid #33C2DF' : '1px solid rgba(15, 10, 1, 0.12)',
                        backgroundColor: isSelected ? '#F0FDF4' : '#FFFFFF',
                        cursor: 'pointer',
                        transition: 'all 0.18s ease',
                      }}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '2px' }}>
                        <User size={13} color={isSelected ? '#33C2DF' : '#64748B'} />
                        <span style={{ fontSize: '13px', fontWeight: 700, color: '#2A3F68' }}>{prof.full_name}</span>
                      </div>
                      <span style={{ fontSize: '11px', color: '#64748B', textTransform: 'capitalize' }}>
                        {prof.relationship} • {prof.age} yrs • {prof.blood_group}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Chief Complaint Input */}
            <div style={{ marginBottom: '18px' }}>
              <label style={{ display: 'block', fontSize: '13px', fontWeight: 700, color: '#2A3F68', marginBottom: '6px' }}>
                {t.chief_complaint_label}
              </label>
              <textarea
                rows={3}
                placeholder={t.chief_complaint_placeholder}
                value={complaint}
                onChange={(e) => setComplaint(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '10px',
                  border: '1px solid rgba(15, 10, 1, 0.16)',
                  outline: 'none',
                  fontSize: '13.5px',
                  lineHeight: 1.5,
                  resize: 'vertical',
                }}
              />
            </div>

            {/* Preferred Time */}
            <div style={{ marginBottom: '18px' }}>
              <label style={{ display: 'block', fontSize: '13px', fontWeight: 700, color: '#2A3F68', marginBottom: '6px' }}>
                পছন্দের সময় / Preferred Consultation Window:
              </label>
              <select
                value={preferredTime}
                onChange={(e) => setPreferredTime(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '10px',
                  border: '1px solid rgba(15, 10, 1, 0.16)',
                  outline: 'none',
                  fontSize: '13.5px',
                  backgroundColor: '#FFFFFF',
                }}
              >
                <option value="Immediate (within 15 minutes)">তাৎক্ষণিক (১৫ মিনিটের মধ্যে / Immediate)</option>
                <option value="Today Evening (6:00 PM - 8:00 PM)">আজ সন্ধ্যা (৬:০০ - ৮:০০ টা)</option>
                <option value="Tonight (8:00 PM - 10:30 PM)">আজ রাত (৮:০০ - ১০:৩০ টা)</option>
                <option value="Tomorrow Morning (10:00 AM - 1:00 PM)">আগামীকাল সকাল (১০:০০ - ১:০০ টা)</option>
              </select>
            </div>

            {/* Telehealth Consent Checkbox */}
            <label
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: '8px',
                cursor: 'pointer',
                fontSize: '12.5px',
                color: '#334155',
                marginBottom: '24px',
                lineHeight: 1.4,
              }}
            >
              <input
                type="checkbox"
                checked={consentGiven}
                onChange={(e) => setConsentGiven(e.target.checked)}
                style={{ marginTop: '2px' }}
              />
              <span>{t.consent_checkbox} (I confirm all entered symptoms are accurate)</span>
            </label>

            {/* Next Button */}
            <button
              onClick={handleCreateBooking}
              disabled={loading}
              className="btn btn-lime"
              style={{ width: '100%', padding: '14px', fontSize: '15px', fontWeight: 800 }}
            >
              {loading ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  <span>Processing Server Quotation...</span>
                </>
              ) : (
                <>
                  <span>পেমেন্ট ধাপে এগিয়ে যান (৳{quotedAmount})</span>
                  <ArrowRight size={16} />
                </>
              )}
            </button>
          </div>
        )}

        {/* STEP 2: Gateway Selection & Simulation */}
        {step === 2 && (
          <div>
            <div
              style={{
                backgroundColor: '#EEF7F8',
                borderRadius: '12px',
                padding: '18px',
                marginBottom: '20px',
                border: '1px solid #DCEAEC',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span style={{ fontSize: '13px', color: '#64748B' }}>Patient:</span>
                <span style={{ fontSize: '13px', fontWeight: 700, color: '#2A3F68' }}>{selectedPatient?.full_name}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span style={{ fontSize: '13px', color: '#64748B' }}>Doctor / Service:</span>
                <span style={{ fontSize: '13px', fontWeight: 700, color: '#2A3F68' }}>
                  {selectedDoctor ? selectedDoctor.full_name : 'General Medical Triage'}
                </span>
              </div>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  borderTop: '1px solid #DCEAEC',
                  paddingTop: '8px',
                  marginTop: '8px',
                }}
              >
                <span style={{ fontSize: '14px', fontWeight: 700, color: '#2A3F68' }}>Total Payable Amount:</span>
                <span style={{ fontSize: '20px', fontWeight: 800, color: '#33C2DF' }}>৳{quotedAmount} BDT</span>
              </div>
            </div>

            {/* Gateway Radios */}
            <div style={{ marginBottom: '24px' }}>
              <label style={{ display: 'block', fontSize: '13px', fontWeight: 700, color: '#2A3F68', marginBottom: '8px' }}>
                পেমেন্ট মাধ্যম বেছে নিন / Choose Payment Gateway:
              </label>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                <div
                  onClick={() => setGateway('bkash')}
                  style={{
                    padding: '16px',
                    borderRadius: '12px',
                    border: gateway === 'bkash' ? '2px solid #E2136E' : '1px solid rgba(15, 10, 1, 0.12)',
                    backgroundColor: gateway === 'bkash' ? '#FFF1F6' : '#FFFFFF',
                    cursor: 'pointer',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '4px',
                    transition: 'all 0.2s',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: '15px', fontWeight: 800, color: '#E2136E' }}>bKash Direct</span>
                    <CreditCard size={18} color="#E2136E" />
                  </div>
                  <span style={{ fontSize: '11px', color: '#64748B' }}>Instant tokenized checkout</span>
                </div>

                <div
                  onClick={() => setGateway('sslcommerz')}
                  style={{
                    padding: '16px',
                    borderRadius: '12px',
                    border: gateway === 'sslcommerz' ? '2px solid #9B468A' : '1px solid rgba(15, 10, 1, 0.12)',
                    backgroundColor: gateway === 'sslcommerz' ? '#F0F9FF' : '#FFFFFF',
                    cursor: 'pointer',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '4px',
                    transition: 'all 0.2s',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: '15px', fontWeight: 800, color: '#9B468A' }}>SSLCOMMERZ</span>
                    <CreditCard size={18} color="#9B468A" />
                  </div>
                  <span style={{ fontSize: '11px', color: '#64748B' }}>Cards, Nagad, Rocket, Netbanking</span>
                </div>
              </div>
            </div>

            {/* Test Simulation Button */}
            <button
              onClick={handleSimulatePayment}
              disabled={loading}
              className="btn btn-lime"
              style={{ width: '100%', padding: '14px', fontSize: '15px', fontWeight: 800 }}
            >
              {loading ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  <span>Verifying Gateway Callback...</span>
                </>
              ) : (
                <>
                  <CreditCard size={16} />
                  <span>
                    {gateway === 'bkash' ? t.pay_btn_bkash : t.pay_btn_ssl} (৳{quotedAmount})
                  </span>
                </>
              )}
            </button>
          </div>
        )}

        {/* STEP 3: Success Confirmation */}
        {step === 3 && (
          <div style={{ textAlign: 'center', padding: '16px 0' }}>
            <div
              style={{
                width: 64,
                height: 64,
                borderRadius: '50%',
                backgroundColor: '#D3F8FF',
                color: '#168CA5',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 16px',
              }}
            >
              <CheckCircle size={36} />
            </div>

            <h3 style={{ fontSize: '20px', fontWeight: 800, color: '#2A3F68', marginBottom: '8px' }}>
              পেমেন্ট সফলভাবে সম্পন্ন হয়েছে!
            </h3>
            <p style={{ fontSize: '14px', color: '#475569', marginBottom: '20px', lineHeight: 1.5 }}>
              আপনার অ্যাপয়েন্টমেন্ট <strong>"Paid/Awaiting Assignment"</strong> অবস্থায় ঢাকা অপারেশন্স সেন্টারে তালিকাভুক্ত হয়েছে। খুব শীঘ্রই দায়িত্বপ্রাপ্ত চিকিৎসক আপনাকে কনফার্মেশন ও ভিডিও কলের লিংক পাঠাবেন।
            </p>

            <div
              style={{
                backgroundColor: '#EEF7F8',
                borderRadius: '10px',
                padding: '12px',
                fontSize: '12px',
                color: '#64748B',
                marginBottom: '24px',
              }}
            >
              Booking Reference: <strong>{createdBookingId}</strong> • Payment Status: <strong>PAID</strong>
            </div>

            <button
              onClick={onClose}
              className="btn btn-primary"
              style={{ width: '100%', padding: '12px', fontWeight: 700 }}
            >
              আমার অ্যাপয়েন্টমেন্ট তালিকায় যান
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

