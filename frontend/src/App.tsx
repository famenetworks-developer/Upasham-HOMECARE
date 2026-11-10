import React, { useState, useEffect } from 'react';
import { RoleSwitcher } from './components/RoleSwitcher';
import { Navbar } from './components/Navbar';
import { PublicHero } from './components/PublicHero';
import { PublicSections } from './components/PublicSections';
import { DoctorDirectory } from './components/DoctorDirectory';
import { DoctorDetailModal } from './components/DoctorDetailModal';
import { BookingModal } from './components/BookingModal';
import { PatientPortal } from './components/PatientPortal';
import { DoctorPortal } from './components/DoctorPortal';
import { OperationsPortal } from './components/OperationsPortal';
import { FinancePortal } from './components/FinancePortal';
import { AdminPortal } from './components/AdminPortal';
import { CMSDashboard } from './components/CMSDashboard';
import { JitsiVideoModal } from './components/JitsiVideoModal';
import { Language, translations } from './utils/translations';
import { Doctor, Specialty, PatientProfile, Booking, Prescription, PaymentTransaction, AuditEvent, User } from './types';
import { ShieldCheck, PhoneCall, Heart, MapPin } from 'lucide-react';
import { defaultSiteContent, local, mergeSiteContent, SiteContent } from './content';

export function App() {
  const [lang, setLang] = useState<Language>('bn');
  const [currentRole, setCurrentRole] = useState<string>('public');
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [activeTab, setActiveTab] = useState<string>('home');

  // Shared Data States
  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [specialties, setSpecialties] = useState<Specialty[]>([]);
  const [patientProfiles, setPatientProfiles] = useState<PatientProfile[]>([]);
  const [patientBookings, setPatientBookings] = useState<Booking[]>([]);
  const [patientPrescriptions, setPatientPrescriptions] = useState<Prescription[]>([]);
  const [doctorAssignments, setDoctorAssignments] = useState<any[]>([]);
  const [operationsQueue, setOperationsQueue] = useState<any[]>([]);
  const [transactions, setTransactions] = useState<PaymentTransaction[]>([]);
  const [financeDashboard, setFinanceDashboard] = useState<any>(null);
  const [adminStats, setAdminStats] = useState<any>(null);
  const [adminDoctors, setAdminDoctors] = useState<any[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditEvent[]>([]);
  const [catalogue, setCatalogue] = useState<any>(null);
  const [siteContent, setSiteContent] = useState<SiteContent>(defaultSiteContent);

  // Modals
  const [selectedDoctorForModal, setSelectedDoctorForModal] = useState<Doctor | null>(null);
  const [isBookingModalOpen, setIsBookingModalOpen] = useState<boolean>(false);
  const [bookingServiceType, setBookingServiceType] = useState<'general' | 'specialist'>('general');
  const [bookingDoctor, setBookingDoctor] = useState<Doctor | null>(null);
  const [activeVideoRoomId, setActiveVideoRoomId] = useState<string | null>(null);

  // 1. Initial Data Fetch (Public catalogue & doctors)
  const fetchPublicData = async () => {
    try {
      const docRes = await fetch('/api/doctors');
      const docData = await docRes.json();
      if (docData.doctors) {
        setDoctors(docData.doctors);
      }

      const catRes = await fetch('/api/catalogue');
      const catData = await catRes.json();
      if (catData.specialties) {
        setSpecialties(catData.specialties);
        setCatalogue(catData);
      }

      const contentRes = await fetch('/api/content');
      if (contentRes.ok) {
        const contentData = await contentRes.json();
        setSiteContent(mergeSiteContent(contentData.content));
      }
    } catch (err) {
      console.error('Failed to load initial directory data:', err);
    }
  };

  useEffect(() => {
    fetchPublicData();
  }, []);

  useEffect(() => {
    document.title = local(siteContent.seo.title, lang);
    document.documentElement.style.setProperty('--upasham-green', siteContent.theme.primary);
    document.documentElement.style.setProperty('--upasham-green-dark', siteContent.theme.dark);
    document.documentElement.style.setProperty('--brand-lime', siteContent.theme.accent);
  }, [siteContent, lang]);

  // 2. Fetch authenticated data based on active role
  const fetchRoleData = async (role: string, token: string) => {
    if (!token) return;
    const authHeaders = { Authorization: `Bearer ${token}` };

    try {
      if (role === 'patient') {
        const [profRes, bkRes, rxRes] = await Promise.all([
          fetch('/api/patient/profiles', { headers: authHeaders }),
          fetch('/api/patient/bookings', { headers: authHeaders }),
          fetch('/api/patient/prescriptions', { headers: authHeaders }),
        ]);

        if (profRes.ok) setPatientProfiles(await profRes.json());
        if (bkRes.ok) setPatientBookings(await bkRes.json());
        if (rxRes.ok) setPatientPrescriptions(await rxRes.json());
      } else if (role.startsWith('doctor')) {
        const asgRes = await fetch('/api/doctor/assignments', { headers: authHeaders });
        if (asgRes.ok) setDoctorAssignments(await asgRes.json());
      } else if (role === 'operations') {
        const qRes = await fetch('/api/operations/queue', { headers: authHeaders });
        if (qRes.ok) setOperationsQueue(await qRes.json());
      } else if (role === 'finance') {
        const [dashRes, txRes] = await Promise.all([
          fetch('/api/finance/dashboard', { headers: authHeaders }),
          fetch('/api/finance/transactions', { headers: authHeaders }),
        ]);
        if (dashRes.ok) setFinanceDashboard(await dashRes.json());
        if (txRes.ok) setTransactions(await txRes.json());
      } else if (role === 'admin') {
        const [stRes, adDocRes, audRes] = await Promise.all([
          fetch('/api/admin/stats', { headers: authHeaders }),
          fetch('/api/admin/doctors', { headers: authHeaders }),
          fetch('/api/admin/audit-logs', { headers: authHeaders }),
        ]);
        if (stRes.ok) setAdminStats(await stRes.json());
        if (adDocRes.ok) setAdminDoctors(await adDocRes.json());
        if (audRes.ok) setAuditLogs(await audRes.json());
      }
    } catch (err) {
      console.error('Role data error:', err);
    }
  };

  // 3. Handle Switch Role
  const handleSwitchRole = async (targetRole: string) => {
    setCurrentRole(targetRole);

    if (targetRole === 'public') {
      setCurrentUser(null);
      localStorage.removeItem('upasham_token');
      setActiveTab('home');
      fetchPublicData();
      return;
    }

    try {
      const res = await fetch('/api/auth/demo-login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: targetRole }),
      });
      const data = await res.json();
      if (res.ok && data.token) {
        localStorage.setItem('upasham_token', data.token);
        setCurrentUser(data.user);

        if (targetRole === 'patient') {
          setActiveTab('patient_appointments');
        } else if (targetRole.startsWith('doctor')) {
          setActiveTab('doctor');
        } else if (targetRole === 'operations') {
          setActiveTab('operations');
        } else if (targetRole === 'finance') {
          setActiveTab('finance');
        } else if (targetRole === 'admin') {
          setActiveTab('cms');
        }

        fetchRoleData(targetRole, data.token);
      }
    } catch (err) {
      console.error('Demo login switch error:', err);
    }
  };

  // Booking action handlers
  const handleOpenBookingGeneral = () => {
    // If not logged in as patient, auto-switch to patient for seamless demo!
    if (!currentUser || currentUser.role !== 'patient') {
      handleSwitchRole('patient').then(() => {
        setBookingServiceType('general');
        setBookingDoctor(null);
        setIsBookingModalOpen(true);
      });
    } else {
      setBookingServiceType('general');
      setBookingDoctor(null);
      setIsBookingModalOpen(true);
    }
  };

  const handleOpenBookingSpecialist = (doc: Doctor) => {
    if (!currentUser || currentUser.role !== 'patient') {
      handleSwitchRole('patient').then(() => {
        setBookingServiceType('specialist');
        setBookingDoctor(doc);
        setIsBookingModalOpen(true);
      });
    } else {
      setBookingServiceType('specialist');
      setBookingDoctor(doc);
      setIsBookingModalOpen(true);
    }
  };

  const handleBookingSuccess = (_bookingId: string) => {
    const token = localStorage.getItem('upasham_token') || '';
    fetchRoleData('patient', token);
    setActiveTab('patient_appointments');
  };

  useEffect(() => {
    if (window.location.hash === '#admin-cms') handleSwitchRole('admin');
  }, []);

  const handleRefreshCurrentView = () => {
    fetchPublicData();
    const token = localStorage.getItem('upasham_token') || '';
    if (token) fetchRoleData(currentRole, token);
  };

  if (currentRole === 'admin' && activeTab === 'cms') {
    return <>
      <CMSDashboard stats={adminStats} doctors={adminDoctors} auditLogs={auditLogs} onRefresh={handleRefreshCurrentView} onOpenLegacy={()=>setActiveTab('admin')} onExit={()=>handleSwitchRole('public')} />
      <RoleSwitcher currentRole={currentRole} onSwitchRole={handleSwitchRole} lang={lang}/>
    </>;
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* 1. Fixed Surface / Demo Role Switcher Bar */}
      <RoleSwitcher
        currentRole={currentRole}
        onSwitchRole={handleSwitchRole}
        lang={lang}
      />

      {/* 2. Top Navigation */}
      <Navbar
        lang={lang}
        onToggleLang={() => setLang(lang === 'bn' ? 'en' : 'bn')}
        currentUser={currentUser}
        activeTab={activeTab}
        onSelectTab={(tab) => {
          setActiveTab(tab);
          if (tab === 'home') setCurrentRole('public');
        }}
        onOpenBookingGeneral={handleOpenBookingGeneral}
        content={siteContent}
      />

      {/* 3. Main Views */}
      <main style={{ flex: 1 }}>
        {/* PUBLIC SITE VIEW */}
        {(activeTab === 'home' || currentRole === 'public') && (
          <>
            <PublicHero
              lang={lang}
              content={siteContent}
              onOpenBookingGeneral={handleOpenBookingGeneral}
              onScrollToSpecialists={() => {
                const el = document.getElementById('specialists');
                if (el) el.scrollIntoView({ behavior: 'smooth' });
              }}
            />
            <DoctorDirectory
              doctors={doctors}
              specialties={specialties}
              lang={lang}
              onSelectDoctor={(doc) => setSelectedDoctorForModal(doc)}
              onBookDoctor={(doc) => handleOpenBookingSpecialist(doc)}
            />
            <PublicSections lang={lang} content={siteContent} onBook={handleOpenBookingGeneral} />
          </>
        )}

        {/* PATIENT PORTAL VIEWS */}
        {currentRole === 'patient' && (activeTab === 'patient_appointments' || activeTab === 'patient_family' || activeTab === 'patient') && (
          <PatientPortal
            profiles={patientProfiles}
            bookings={patientBookings}
            prescriptions={patientPrescriptions}
            onRefresh={handleRefreshCurrentView}
            onJoinVideo={(roomId) => setActiveVideoRoomId(roomId)}
            onOpenBooking={handleOpenBookingGeneral}
            lang={lang}
          />
        )}

        {/* DOCTOR PORTAL VIEW */}
        {currentRole.startsWith('doctor') && activeTab === 'doctor' && (
          <DoctorPortal
            assignments={doctorAssignments}
            onRefresh={handleRefreshCurrentView}
            onJoinVideo={(roomId) => setActiveVideoRoomId(roomId)}
            lang={lang}
          />
        )}

        {/* OPERATIONS PORTAL VIEW */}
        {currentRole === 'operations' && activeTab === 'operations' && (
          <OperationsPortal
            queue={operationsQueue}
            doctors={doctors}
            onRefresh={handleRefreshCurrentView}
            lang={lang}
          />
        )}

        {/* FINANCE PORTAL VIEW */}
        {currentRole === 'finance' && activeTab === 'finance' && (
          <FinancePortal
            transactions={transactions}
            dashboard={financeDashboard}
            onRefresh={handleRefreshCurrentView}
            lang={lang}
          />
        )}

        {/* ADMIN PORTAL VIEW */}
        {currentRole === 'admin' && activeTab === 'admin' && (
          <AdminPortal
            stats={adminStats}
            doctors={adminDoctors}
            auditLogs={auditLogs}
            catalogue={catalogue}
            onRefresh={handleRefreshCurrentView}
            lang={lang}
          />
        )}
      </main>

      {/* 4. Doctor Detail Modal */}
      <DoctorDetailModal
        doctor={selectedDoctorForModal}
        onClose={() => setSelectedDoctorForModal(null)}
        onBook={(doc) => handleOpenBookingSpecialist(doc)}
        lang={lang}
      />

      {/* 5. Booking Modal (5-Step Wizard) */}
      <BookingModal
        isOpen={isBookingModalOpen}
        onClose={() => setIsBookingModalOpen(false)}
        selectedDoctor={bookingDoctor}
        serviceType={bookingServiceType}
        patientProfiles={patientProfiles}
        onBookingSuccess={handleBookingSuccess}
        lang={lang}
      />

      {/* 6. Jitsi Teleconsultation Video Room Modal */}
      {activeVideoRoomId && (
        <JitsiVideoModal
          roomId={activeVideoRoomId}
          onClose={() => {
            setActiveVideoRoomId(null);
            handleRefreshCurrentView();
          }}
          userRole={currentRole}
        />
      )}

      {/* 7. Safeer Editorial Footer */}
      <footer
        className="health-footer"
        style={{
          backgroundColor: '#073B36',
          color: '#FFFFFF',
          borderTop: '1px solid rgba(255, 255, 255, 0.1)',
          padding: '48px 0 24px',
        }}
      >
        <div className="container">
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: '32px',
              marginBottom: '40px',
            }}
          >
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <span
                  style={{
                    fontFamily: "'Hind Siliguri', sans-serif",
                    fontSize: '22px',
                    fontWeight: 800,
                    color: '#0B6B5C',
                  }}
                >
                  উপশম প্লাস
                </span>
                <span style={{ fontSize: '11px', backgroundColor: '#15803D', color: '#FFF', padding: '2px 8px', borderRadius: '999px', fontWeight: 700 }}>
                  CONNECTED CARE
                </span>
              </div>
              <p style={{ fontSize: '13.5px', color: '#AAADB0', lineHeight: 1.5, maxWidth: '280px' }}>
                {local(siteContent.footer.description, lang)}
              </p>
            </div>

            <div>
              <h4 style={{ fontSize: '14px', fontWeight: 700, color: '#0B6B5C', marginBottom: '12px' }}>
                {local(siteContent.footer.servicesTitle, lang)}
              </h4>
              <ul style={{ listStyle: 'none', display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px', color: '#D9D9D9' }}>
                {siteContent.footer.services.map((x,i)=><li key={i}>{local(x,lang)}</li>)}
              </ul>
            </div>

            <div>
              <h4 style={{ fontSize: '14px', fontWeight: 700, color: '#0B6B5C', marginBottom: '12px' }}>
                {local(siteContent.footer.coverageTitle,lang)}
              </h4>
              <p style={{ fontSize: '13px', color: '#AAADB0', lineHeight: 1.5 }}>
                {local(siteContent.footer.coverage,lang)}
              </p>
            </div>

            <div>
              <h4 style={{ fontSize: '14px', fontWeight: 700, color: '#0B6B5C', marginBottom: '12px' }}>
                {local(siteContent.footer.supportTitle,lang)}
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px', color: '#D9D9D9' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <PhoneCall size={14} color="#0B6B5C" />
                  <span>জরুরি হটলাইন: <strong>৯৯৯</strong></span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <ShieldCheck size={14} color="#0B6B5C" />
                  <span>{lang==='bn'?'উপশম হেল্পলাইন':'Upasham hotline'}: {siteContent.footer.hotline}</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <MapPin size={14} color="#0284C7" />
                  <span>{local(siteContent.footer.address,lang)}</span>
                </div>
              </div>
            </div>
          </div>

          <div
            style={{
              borderTop: '1px solid rgba(255, 255, 255, 0.1)',
              paddingTop: '20px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              flexWrap: 'wrap',
              gap: '12px',
              fontSize: '12px',
              color: '#AAADB0',
            }}
          >
            <div>
              {siteContent.footer.copyright}
            </div>
            <div style={{ display: 'flex', gap: '16px' }}>
              <span>bKash & SSLCOMMERZ Protected</span>
              <span>•</span>
              <span>Encrypted Jitsi Telehealth</span>
              <span>•</span>
              <span>Security controls enabled - legal review pending</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;


