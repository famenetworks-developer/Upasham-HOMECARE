import React from 'react';
import { Doctor } from '../types';
import { X, CheckCircle2, Award, MapPin, Globe2, Star, Calendar, ShieldCheck } from 'lucide-react';
import { Language, translations } from '../utils/translations';

interface DoctorDetailModalProps {
  doctor: Doctor | null;
  onClose: () => void;
  onBook: (doctor: Doctor) => void;
  lang: Language;
}

export const DoctorDetailModal: React.FC<DoctorDetailModalProps> = ({
  doctor,
  onClose,
  onBook,
  lang,
}) => {
  if (!doctor) return null;
  const t = translations[lang];

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(15, 10, 1, 0.65)',
        backdropFilter: 'blur(4px)',
        zIndex: 200,
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
          maxWidth: '680px',
          maxHeight: '90vh',
          overflowY: 'auto',
          padding: '32px',
          backgroundColor: '#FFFFFF',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Modal Header */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '20px' }}>
          <div style={{ display: 'flex', gap: '20px', alignItems: 'center' }}>
            <img
              src={doctor.avatar_url}
              alt={doctor.full_name}
              style={{
                width: 84,
                height: 84,
                borderRadius: '16px',
                objectFit: 'cover',
                border: '2px solid rgba(42, 63, 104, 0.08)',
              }}
            />
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                <span className="badge badge-green">
                  <CheckCircle2 size={12} /> {t.doctor_verified}
                </span>
                {doctor.bmdc_number && (
                  <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 600 }}>
                    Reg: {doctor.bmdc_number}
                  </span>
                )}
              </div>
              <h2 style={{ fontSize: '22px', fontWeight: 800, color: '#2A3F68', lineHeight: 1.2 }}>
                {doctor.full_name}
              </h2>
              <p style={{ fontSize: '13.5px', color: '#475569', marginTop: '2px' }}>
                {doctor.qualifications}
              </p>
            </div>
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

        {/* Doctor Badges & Highlights */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))',
            gap: '12px',
            backgroundColor: '#EEF7F8',
            borderRadius: '12px',
            padding: '16px',
            marginBottom: '24px',
            border: '1px solid rgba(15, 10, 1, 0.06)',
          }}
        >
          <div>
            <span style={{ fontSize: '11px', color: '#64748B', display: 'block', textTransform: 'uppercase', fontWeight: 600 }}>
              Hospital Affiliation
            </span>
            <div style={{ display: 'flex', alignItems: 'center', gap: '5px', marginTop: '3px' }}>
              <MapPin size={14} color="#9B468A" />
              <span style={{ fontSize: '13px', fontWeight: 600, color: '#2A3F68' }}>{doctor.workplace}</span>
            </div>
          </div>

          {doctor.experience_years > 0 && (
            <div>
              <span style={{ fontSize: '11px', color: '#64748B', display: 'block', textTransform: 'uppercase', fontWeight: 600 }}>
                Clinical Experience
              </span>
              <div style={{ display: 'flex', alignItems: 'center', gap: '5px', marginTop: '3px' }}>
                <Award size={14} color="#33C2DF" />
                <span style={{ fontSize: '13px', fontWeight: 600, color: '#2A3F68' }}>{doctor.experience_years}+ Years</span>
              </div>
            </div>
          )}

          {doctor.total_reviews > 0 && (
            <div>
              <span style={{ fontSize: '11px', color: '#64748B', display: 'block', textTransform: 'uppercase', fontWeight: 600 }}>
                Rating & Reviews
              </span>
              <div style={{ display: 'flex', alignItems: 'center', gap: '5px', marginTop: '3px' }}>
                <Star size={14} color="#F59E0B" fill="#F59E0B" />
                <span style={{ fontSize: '13px', fontWeight: 600, color: '#2A3F68' }}>
                  {doctor.rating} ({doctor.total_reviews} reviews)
                </span>
              </div>
            </div>
          )}
        </div>

        {/* Full Biography */}
        <div style={{ marginBottom: '24px' }}>
          <h4 style={{ fontSize: '14px', fontWeight: 700, color: '#2A3F68', marginBottom: '8px', textTransform: 'uppercase', letterSpacing: '0.03em' }}>
            Doctor's Profile & Clinical Focus
          </h4>
          <p style={{ fontSize: '14.5px', color: '#334155', lineHeight: 1.6 }}>
            {lang === 'bn' && doctor.bio_bn ? doctor.bio_bn : doctor.bio}
          </p>
        </div>

        {/* Verification & Privacy notice */}
        <div
          style={{
            backgroundColor: '#F0FDF4',
            border: '1px solid #BBF7D0',
            borderRadius: '12px',
            padding: '14px',
            marginBottom: '28px',
            display: 'flex',
            alignItems: 'flex-start',
            gap: '10px',
          }}
        >
          <ShieldCheck size={20} color="#33C2DF" style={{ flexShrink: 0, marginTop: '2px' }} />
          <div style={{ fontSize: '12.5px', color: '#166534', lineHeight: 1.45 }}>
            <strong>Verified Bangladesh Medical & Dental Council (BMDC) Practitioner:</strong> All tele-consultations are encrypted and occur inside the Upasham Plus secure video environment. Direct private doctor contact numbers are protected per platform safety rules.
          </div>
        </div>

        {/* Bottom Booking Action */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderTop: '1px solid rgba(42, 63, 104, 0.08)',
            paddingTop: '20px',
          }}
        >
          <div>
            <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 600 }}>Consultation Fee:</span>
            <div style={{ fontSize: '24px', fontWeight: 800, color: '#2A3F68' }}>
              ৳{doctor.consultation_fee} <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 500 }}>BDT (20 mins)</span>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '10px' }}>
            <button onClick={onClose} className="btn btn-outline" style={{ padding: '10px 18px' }}>
              Close
            </button>
            <button
              onClick={() => {
                onClose();
                onBook(doctor);
              }}
              className="btn btn-lime"
              style={{ padding: '10px 24px', fontWeight: 800 }}
            >
              <Calendar size={15} />
              Book Appointment
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

