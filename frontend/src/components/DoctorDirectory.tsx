import React, { useState } from 'react';
import { Doctor, Specialty } from '../types';
import { Search, Star, Award, MapPin, Globe2, Clock, CheckCircle2 } from 'lucide-react';
import { Language, translations } from '../utils/translations';

interface DoctorDirectoryProps {
  doctors: Doctor[];
  specialties: Specialty[];
  lang: Language;
  onSelectDoctor: (doctor: Doctor) => void;
  onBookDoctor: (doctor: Doctor) => void;
}

export const DoctorDirectory: React.FC<DoctorDirectoryProps> = ({
  doctors,
  specialties,
  lang,
  onSelectDoctor,
  onBookDoctor,
}) => {
  const t = translations[lang];
  const [selectedSpecialty, setSelectedSpecialty] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [selectedLanguage, setSelectedLanguage] = useState<string>('all');

  const filteredDoctors = doctors.filter((doc) => {
    // Specialty match
    if (selectedSpecialty !== 'all') {
      const matchSpec = doc.specialties.includes(selectedSpecialty);
      if (!matchSpec) return false;
    }
    // Search match
    if (searchQuery.trim() !== '') {
      const q = searchQuery.toLowerCase();
      const matchName = doc.full_name.toLowerCase().includes(q);
      const matchQual = doc.qualifications.toLowerCase().includes(q);
      const matchWork = doc.workplace.toLowerCase().includes(q);
      if (!matchName && !matchQual && !matchWork) return false;
    }
    // Language match
    if (selectedLanguage !== 'all') {
      if (!doc.languages.includes(selectedLanguage)) return false;
    }
    return true;
  });

  return (
    <section id="specialists" style={{ padding: '64px 0', backgroundColor: '#F7FBFC' }}>
      <div className="container">
        {/* Section Header */}
        <div style={{ marginBottom: '28px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '16px', marginBottom: '8px' }}>
            <div>
              <span
                style={{
                  fontSize: '12px',
                  fontWeight: 800,
                  textTransform: 'uppercase',
                  color: '#9B468A',
                  letterSpacing: '0.04em',
                }}
              >
                {lang === 'bn' ? 'ঢাকা স্পেশালিস্ট ডিরেক্টরি' : 'DHAKA SPECIALIST DIRECTORY'}
              </span>
              <h2
                style={{
                  fontSize: 'clamp(24px, 3vw, 36px)',
                  fontWeight: 800,
                  color: '#2A3F68',
                  letterSpacing: '-0.02em',
                }}
              >
                {t.specialties_title}
              </h2>
            </div>

            {/* Language filter pills */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 600 }}>Language:</span>
              <button
                onClick={() => setSelectedLanguage('all')}
                style={{
                  padding: '4px 10px',
                  borderRadius: '999px',
                  fontSize: '12px',
                  fontWeight: selectedLanguage === 'all' ? 700 : 500,
                  backgroundColor: selectedLanguage === 'all' ? '#2A3F68' : '#FFFFFF',
                  color: selectedLanguage === 'all' ? '#FFFFFF' : '#475569',
                  border: '1px solid rgba(15, 10, 1, 0.1)',
                  cursor: 'pointer',
                }}
              >
                All
              </button>
              <button
                onClick={() => setSelectedLanguage('Bangla')}
                style={{
                  padding: '4px 10px',
                  borderRadius: '999px',
                  fontSize: '12px',
                  fontWeight: selectedLanguage === 'Bangla' ? 700 : 500,
                  backgroundColor: selectedLanguage === 'Bangla' ? '#2A3F68' : '#FFFFFF',
                  color: selectedLanguage === 'Bangla' ? '#FFFFFF' : '#475569',
                  border: '1px solid rgba(15, 10, 1, 0.1)',
                  cursor: 'pointer',
                }}
              >
                বাংলা
              </button>
              <button
                onClick={() => setSelectedLanguage('English')}
                style={{
                  padding: '4px 10px',
                  borderRadius: '999px',
                  fontSize: '12px',
                  fontWeight: selectedLanguage === 'English' ? 700 : 500,
                  backgroundColor: selectedLanguage === 'English' ? '#2A3F68' : '#FFFFFF',
                  color: selectedLanguage === 'English' ? '#FFFFFF' : '#475569',
                  border: '1px solid rgba(15, 10, 1, 0.1)',
                  cursor: 'pointer',
                }}
              >
                English
              </button>
            </div>
          </div>

          <p style={{ color: '#64748B', fontSize: '15px' }}>{t.specialties_subtitle}</p>
        </div>

        {/* Search Bar */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            backgroundColor: '#FFFFFF',
            border: '1px solid rgba(42, 63, 104, 0.12)',
            borderRadius: '12px',
            padding: '12px 18px',
            marginBottom: '20px',
            boxShadow: '0 8px 24px rgba(42,63,104,0.06)',
          }}
        >
          <Search size={18} color="#64748B" />
          <input
            type="text"
            placeholder={t.doctor_search_placeholder}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              width: '100%',
              border: 'none',
              outline: 'none',
              fontSize: '14.5px',
              color: '#2A3F68',
              backgroundColor: 'transparent',
            }}
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              style={{
                border: 'none',
                background: 'none',
                color: '#64748B',
                cursor: 'pointer',
                fontSize: '12px',
                fontWeight: 600,
              }}
            >
              Clear
            </button>
          )}
        </div>

        {/* Specialty Pills Horizontal Bar */}
        <div
          style={{
            display: 'flex',
            gap: '8px',
            overflowX: 'auto',
            paddingBottom: '12px',
            marginBottom: '32px',
            scrollbarWidth: 'thin',
          }}
        >
          <button
            onClick={() => setSelectedSpecialty('all')}
            style={{
              padding: '8px 18px',
              borderRadius: '999px',
              border: selectedSpecialty === 'all' ? '1px solid #33C2DF' : '1px solid rgba(42, 63, 104, 0.1)',
              backgroundColor: selectedSpecialty === 'all' ? '#33C2DF' : '#FFFFFF',
              color: selectedSpecialty === 'all' ? '#FFFFFF' : '#2A3F68',
              fontWeight: 700,
              fontSize: '13.5px',
              whiteSpace: 'nowrap',
              cursor: 'pointer',
              transition: 'all 0.2s',
            }}
          >
            {t.specialties_all} ({doctors.length})
          </button>

          {specialties.map((spec) => {
            const count = doctors.filter((d) => d.specialties.includes(spec.id)).length;
            const isSelected = selectedSpecialty === spec.id;

            return (
              <button
                key={spec.id}
                onClick={() => setSelectedSpecialty(spec.id)}
                style={{
                  padding: '8px 16px',
                  borderRadius: '999px',
                  border: isSelected ? '1px solid #33C2DF' : '1px solid rgba(42, 63, 104, 0.1)',
                  backgroundColor: isSelected ? '#33C2DF' : '#FFFFFF',
                  color: isSelected ? '#FFFFFF' : '#2A3F68',
                  fontWeight: isSelected ? 700 : 500,
                  fontSize: '13.5px',
                  whiteSpace: 'nowrap',
                  cursor: 'pointer',
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '6px',
                  transition: 'all 0.2s',
                }}
              >
                <span>{lang === 'bn' ? spec.name_bn : spec.name}</span>
                <span
                  style={{
                    fontSize: '11px',
                    opacity: 0.85,
                    backgroundColor: isSelected ? 'rgba(255,255,255,0.2)' : '#F1F5F9',
                    padding: '1px 6px',
                    borderRadius: '999px',
                  }}
                >
                  {count}
                </span>
              </button>
            );
          })}
        </div>

        {/* Doctor Grid */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))',
            gap: '24px',
          }}
        >
          {filteredDoctors.map((doc) => (
            <div
              key={doc.id}
              className="card-safeer"
              style={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
                padding: '24px',
                position: 'relative',
              }}
            >
              <div>
                {/* Doctor Header: Avatar, Name, BMDC verified badge */}
                <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start', marginBottom: '14px' }}>
                  <img
                    src={doc.avatar_url}
                    alt={doc.full_name}
                    style={{
                      width: 68,
                      height: 68,
                      borderRadius: '12px',
                      objectFit: 'cover',
                      border: '1px solid rgba(15, 10, 1, 0.08)',
                    }}
                  />
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '4px' }}>
                      <span className="badge badge-green" style={{ fontSize: '11px', padding: '3px 8px' }}>
                        <CheckCircle2 size={11} /> {t.doctor_verified}
                      </span>
                      {doc.bmdc_number && (
                        <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 600 }}>
                          {doc.bmdc_number}
                        </span>
                      )}
                    </div>

                    <h3
                      onClick={() => onSelectDoctor(doc)}
                      style={{
                        fontSize: '17px',
                        fontWeight: 800,
                        color: '#2A3F68',
                        cursor: 'pointer',
                        lineHeight: 1.25,
                      }}
                    >
                      {doc.full_name}
                    </h3>

                    <p style={{ fontSize: '12.5px', color: '#475569', marginTop: '3px' }}>
                      {doc.qualifications}
                    </p>
                  </div>
                </div>

                {/* Experience & Workplace */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', marginBottom: '16px', fontSize: '13px', color: '#475569' }}>
                  {doc.experience_years > 0 && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <Award size={14} color="#33C2DF" />
                      <span>
                        <strong>{doc.experience_years}+</strong> {t.doctor_exp}
                      </span>
                    </div>
                  )}

                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <MapPin size={14} color="#0284C7" />
                    <span style={{ color: '#334155', fontWeight: 500 }}>{doc.workplace}</span>
                  </div>

                  {doc.availability && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <Clock size={14} color="#D97706" />
                      <span style={{ color: '#B45309', fontWeight: 600, fontSize: '12px' }}>
                        {doc.availability}
                      </span>
                    </div>
                  )}

                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <Globe2 size={14} color="#64748B" />
                    <span style={{ fontSize: '12px', color: '#64748B' }}>
                      Languages: {doc.languages.join(', ')}
                    </span>
                  </div>
                </div>

                {/* Short Bio */}
                <p style={{ fontSize: '13px', color: '#64748B', lineHeight: 1.45, marginBottom: '20px' }}>
                  {lang === 'bn' && doc.bio_bn ? doc.bio_bn : doc.bio}
                </p>
              </div>

              {/* Bottom Card Footer: Fee, Rating, Actions */}
              <div
                style={{
                  borderTop: '1px solid rgba(15, 10, 1, 0.08)',
                  paddingTop: '16px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: '8px',
                }}
              >
                <div>
                  <span style={{ fontSize: '11px', color: '#64748B', display: 'block', textTransform: 'uppercase', fontWeight: 600 }}>
                    {t.doctor_fee}
                  </span>
                  <div style={{ display: 'flex', alignItems: 'baseline', gap: '4px' }}>
                    <span style={{ fontSize: '20px', fontWeight: 800, color: '#2A3F68' }}>
                      ৳{doc.consultation_fee}
                    </span>
                    <span style={{ fontSize: '11px', color: '#64748B' }}>/ 20 min</span>
                  </div>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <button
                    onClick={() => onSelectDoctor(doc)}
                    className="btn btn-outline"
                    style={{ padding: '8px 12px', fontSize: '12.5px' }}
                  >
                    {t.doctor_view_btn}
                  </button>

                  <button
                    onClick={() => onBookDoctor(doc)}
                    className="btn btn-lime"
                    style={{ padding: '8px 16px', fontSize: '13px', fontWeight: 800 }}
                  >
                    {t.doctor_book_btn}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {filteredDoctors.length === 0 && (
          <div
            style={{
              textAlign: 'center',
              padding: '48px 20px',
              backgroundColor: '#FFFFFF',
              borderRadius: '12px',
              border: '1px dashed rgba(15, 10, 1, 0.15)',
            }}
          >
            <p style={{ fontSize: '16px', fontWeight: 600, color: '#475569' }}>
              No doctors found matching your filter criteria.
            </p>
            <button
              onClick={() => {
                setSelectedSpecialty('all');
                setSearchQuery('');
                setSelectedLanguage('all');
              }}
              className="btn btn-outline"
              style={{ marginTop: '12px' }}
            >
              Reset Filters
            </button>
          </div>
        )}
      </div>
    </section>
  );
};

