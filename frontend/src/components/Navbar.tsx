import React from 'react';
import { Logo } from './Logo';
import { User as UserIcon, PhoneCall, AlertTriangle, Languages, Search, Heart, ShoppingCart } from 'lucide-react';
import { translations, Language } from '../utils/translations';
import { User } from '../types';
import { defaultSiteContent, local, SiteContent } from '../content';

interface NavbarProps {
  lang: Language;
  onToggleLang: () => void;
  currentUser: User | null;
  activeTab: string;
  onSelectTab: (tab: string) => void;
  onOpenBookingGeneral: () => void;
  content?: SiteContent;
}

export const Navbar: React.FC<NavbarProps> = ({
  lang,
  onToggleLang,
  currentUser,
  activeTab,
  onSelectTab,
  onOpenBookingGeneral,
  content = defaultSiteContent,
}) => {
  const t = translations[lang];

  return (
    <header style={{ backgroundColor: '#FFFFFF', borderBottom: '1px solid rgba(42, 63, 104, 0.08)' }}>
      {/* Emergency Alert Ribbon (Safeer editorial high-contrast notice) */}
      {content.announcement.enabled && <div
        style={{
          backgroundColor: '#FFE9EB',
          borderBottom: '1px solid #F5C9CF',
          padding: '6px 16px',
          fontSize: '12px',
          color: '#8C3D4B',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
          fontWeight: 600,
        }}
      >
        <AlertTriangle size={14} color="#DC2626" />
        <span>{local(content.announcement.text,lang)}</span>
        <span style={{ opacity: 0.5 }}>|</span>
        <a href={`tel:${content.announcement.phone}`} style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', color: '#991B1B', textDecoration: 'underline' }}>
          <PhoneCall size={12} /> {content.announcement.phone}
        </a>
      </div>}

      {/* Main Navigation Bar */}
      <div className="container" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '16px 20px' }}>
        {/* Brand Logo */}
        <div style={{ cursor: 'pointer' }} onClick={() => onSelectTab('home')}>
          <Logo size="md" showTagline={true} />
        </div>

        {/* Center Nav Links */}
        <nav style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
          <button
            onClick={() => onSelectTab('home')}
            style={{
              background: 'none',
              border: 'none',
              fontSize: '14.5px',
              fontWeight: activeTab === 'home' ? 700 : 500,
              color: activeTab === 'home' ? '#0B6B5C' : '#073B36',
              cursor: 'pointer',
              padding: '6px 2px',
            }}
          >
            {t.nav_doctors}
          </button>

          <button
            onClick={onOpenBookingGeneral}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
              backgroundColor: '#0B6B5C',
              border: '1px solid #0B6B5C',
              padding: '6px 14px',
              borderRadius: '8px',
              fontSize: '13.5px',
              fontWeight: 700,
              color: '#FFFFFF',
              cursor: 'pointer',
              transition: 'all 0.2s',
            }}
          >
            <span style={{ width: 7, height: 7, borderRadius: '50%', backgroundColor: '#FFFFFF', display: 'inline-block' }} />
            {t.nav_triage}
          </button>

          {currentUser?.role === 'patient' && (
            <>
              <button
                onClick={() => onSelectTab('patient_family')}
                style={{
                  background: 'none',
                  border: 'none',
                  fontSize: '14.5px',
                  fontWeight: activeTab === 'patient_family' ? 700 : 500,
                  color: activeTab === 'patient_family' ? '#0B6B5C' : '#073B36',
                  cursor: 'pointer',
                }}
              >
                {t.nav_family}
              </button>
              <button
                onClick={() => onSelectTab('patient_appointments')}
                style={{
                  background: 'none',
                  border: 'none',
                  fontSize: '14.5px',
                  fontWeight: activeTab === 'patient_appointments' ? 700 : 500,
                  color: activeTab === 'patient_appointments' ? '#0B6B5C' : '#073B36',
                  cursor: 'pointer',
                }}
              >
                {t.nav_appointments}
              </button>
            </>
          )}

          {currentUser && currentUser.role !== 'patient' && currentUser.role !== 'public' && (
            <button
              onClick={() => onSelectTab(currentUser.role === 'admin' ? 'cms' : currentUser.role)}
              style={{
                background: 'rgba(2, 132, 199, 0.08)',
                border: '1px solid rgba(2, 132, 199, 0.2)',
                color: '#0284C7',
                padding: '6px 12px',
                borderRadius: '8px',
                fontSize: '13.5px',
                fontWeight: 700,
                cursor: 'pointer',
              }}
            >
              {currentUser.role.toUpperCase()} {t.nav_portal}
            </button>
          )}
        </nav>

        {/* Right Actions: Bilingual Toggle & Current User */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          {/* Language Switcher */}
          <button
            onClick={onToggleLang}
            className="btn btn-outline"
            style={{
              padding: '6px 14px',
              fontSize: '13px',
              borderRadius: '999px',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
            }}
            title="Switch Language / ভাষা পরিবর্তন করুন"
          >
            <Languages size={14} color="#073B36" />
            <span>{lang === 'bn' ? 'English' : 'বাংলা'}</span>
          </button>

          {/* User Badge / Profile button */}
          {currentUser && (
            <div
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '8px',
                padding: '6px 14px',
                backgroundColor: '#EEF7F8',
                borderRadius: '999px',
                fontSize: '13px',
                fontWeight: 600,
              }}
            >
              <UserIcon size={14} color="#0B6B5C" />
              <span>{currentUser.full_name}</span>
            </div>
          )}
        </div>
      </div>
      <div className="container header-search-row">
        <div className="header-search"><Search size={18}/><input aria-label="Search doctors and services" placeholder={lang === 'bn' ? 'ডাক্তার বা সেবা খুঁজুন' : 'Search doctors and services'}/></div>
        <button aria-label="Saved doctors"><Heart size={18}/></button>
        <button aria-label="Bookings"><ShoppingCart size={18}/></button>
      </div>
      <nav className="category-nav container">
        <button onClick={() => onSelectTab('home')}>Home</button>
        <button onClick={onOpenBookingGeneral}>Instant Doctor</button>
        <button onClick={() => onSelectTab('home')}>Specialist Doctor</button>
        <button>Home Lab Test</button><button>Mental Wellness</button><button>Home Care</button><button>More</button>
      </nav>
    </header>
  );
};


