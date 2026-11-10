import React, { useState } from 'react';
import { AuditEvent } from '../types';
import { Shield, CheckCircle2, XCircle, Users, Activity, FileCheck, Layers, Eye, RefreshCw, LayoutTemplate } from 'lucide-react';
import { Language } from '../utils/translations';
import { ContentStudio } from './ContentStudio';

interface AdminPortalProps {
  stats: any;
  doctors: any[];
  auditLogs: AuditEvent[];
  catalogue: any;
  onRefresh: () => void;
  lang: Language;
}

export const AdminPortal: React.FC<AdminPortalProps> = ({
  stats,
  doctors,
  auditLogs,
  catalogue,
  onRefresh,
  lang,
}) => {
  const [activeTab, setActiveTab] = useState<'content' | 'doctors' | 'audit' | 'catalogue'>('content');
  const [updatingId, setUpdatingId] = useState<string>('');

  const handleToggleVerify = async (docId: string) => {
    setUpdatingId(docId);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      await fetch(`/api/admin/doctors/${docId}/verify`, {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${token}` },
      });
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setUpdatingId('');
    }
  };

  const handleTogglePublish = async (docId: string) => {
    setUpdatingId(docId);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      await fetch(`/api/admin/doctors/${docId}/publish`, {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${token}` },
      });
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setUpdatingId('');
    }
  };

  return (
    <div style={{ padding: '36px 0', backgroundColor: '#F7FBFC', minHeight: '80vh' }}>
      <div className="container">
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <span style={{ fontSize: '12px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
              SUPERVISORY CONTROL
            </span>
            <h1 style={{ fontSize: '28px', fontWeight: 800, color: '#2A3F68' }}>
              System Administration & Clinical Governance
            </h1>
          </div>

          <button onClick={onRefresh} className="btn btn-outline" style={{ fontSize: '13px' }}>
            <RefreshCw size={14} /> Refresh Records
          </button>
        </div>

        {/* High-level Platform Metrics */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '32px' }}>
          <div className="card-safeer" style={{ padding: '20px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Total Users
            </span>
            <div style={{ fontSize: '26px', fontWeight: 800, color: '#2A3F68', marginTop: '2px' }}>
              {stats?.total_users || 0}
            </div>
            <span style={{ fontSize: '12px', color: '#33C2DF' }}>Patients & Staff</span>
          </div>

          <div className="card-safeer" style={{ padding: '20px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Practicing Doctors
            </span>
            <div style={{ fontSize: '26px', fontWeight: 800, color: '#2A3F68', marginTop: '2px' }}>
              {stats?.total_doctors || 0}
            </div>
            <span style={{ fontSize: '12px', color: '#9B468A' }}>BMDC Licensed</span>
          </div>

          <div className="card-safeer" style={{ padding: '20px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Total Consultations
            </span>
            <div style={{ fontSize: '26px', fontWeight: 800, color: '#2A3F68', marginTop: '2px' }}>
              {stats?.total_bookings || 0}
            </div>
            <span style={{ fontSize: '12px', color: '#D97706' }}>Dhaka City Triage</span>
          </div>

          <div className="card-safeer" style={{ padding: '20px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Platform Gross
            </span>
            <div style={{ fontSize: '26px', fontWeight: 800, color: '#33C2DF', marginTop: '2px' }}>
              ৳{(stats?.total_revenue || 0).toLocaleString()}
            </div>
            <span style={{ fontSize: '12px', color: '#64748B' }}>BDT Processed</span>
          </div>
        </div>

        {/* Tab Navigation */}
        <div style={{ display: 'flex', gap: '8px', borderBottom: '1px solid rgba(15,10,1,0.1)', marginBottom: '28px' }}>
          <button onClick={()=>setActiveTab('content')} style={{padding:'10px 18px',border:'none',background:'none',borderBottom:activeTab==='content'?'2px solid #0B6B5C':'2px solid transparent',color:activeTab==='content'?'#0B6B5C':'#64748B',fontWeight:activeTab==='content'?700:500,fontSize:'14.5px',cursor:'pointer',display:'inline-flex',alignItems:'center',gap:'6px'}}><LayoutTemplate size={16}/><span>Content Studio</span></button>
          <button
            onClick={() => setActiveTab('doctors')}
            style={{
              padding: '10px 18px',
              border: 'none',
              background: 'none',
              borderBottom: activeTab === 'doctors' ? '2px solid #33C2DF' : '2px solid transparent',
              color: activeTab === 'doctors' ? '#33C2DF' : '#64748B',
              fontWeight: activeTab === 'doctors' ? 700 : 500,
              fontSize: '14.5px',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <Users size={16} />
            <span>Doctor Credentials & Publishing ({doctors.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('audit')}
            style={{
              padding: '10px 18px',
              border: 'none',
              background: 'none',
              borderBottom: activeTab === 'audit' ? '2px solid #33C2DF' : '2px solid transparent',
              color: activeTab === 'audit' ? '#33C2DF' : '#64748B',
              fontWeight: activeTab === 'audit' ? 700 : 500,
              fontSize: '14.5px',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <Activity size={16} />
            <span>Security & Clinical Audit Trail ({auditLogs.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('catalogue')}
            style={{
              padding: '10px 18px',
              border: 'none',
              background: 'none',
              borderBottom: activeTab === 'catalogue' ? '2px solid #33C2DF' : '2px solid transparent',
              color: activeTab === 'catalogue' ? '#33C2DF' : '#64748B',
              fontWeight: activeTab === 'catalogue' ? 700 : 500,
              fontSize: '14.5px',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <Layers size={16} />
            <span>Catalogue & Service Zones</span>
          </button>
        </div>

        {activeTab === 'content' && <ContentStudio />}

        {/* TAB 1: DOCTORS MANAGEMENT */}
        {activeTab === 'doctors' && (
          <div className="card-safeer" style={{ overflow: 'hidden', padding: 0 }}>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
                <thead style={{ backgroundColor: '#EEF7F8', borderBottom: '1px solid #DCEAEC', color: '#475569' }}>
                  <tr>
                    <th style={{ padding: '14px 18px' }}>Doctor Name</th>
                    <th style={{ padding: '14px 18px' }}>BMDC Reg</th>
                    <th style={{ padding: '14px 18px' }}>Qualifications & Workplace</th>
                    <th style={{ padding: '14px 18px' }}>Fee (BDT)</th>
                    <th style={{ padding: '14px 18px' }}>BMDC Verified</th>
                    <th style={{ padding: '14px 18px' }}>Public Directory</th>
                    <th style={{ padding: '14px 18px', textAlign: 'right' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {doctors.map((doc) => (
                    <tr key={doc.id} style={{ borderBottom: '1px solid #EEF7F8' }}>
                      <td style={{ padding: '14px 18px', fontWeight: 700, color: '#2A3F68' }}>
                        {doc.full_name}
                        <div style={{ fontSize: '11px', color: '#64748B' }}>{doc.email}</div>
                      </td>

                      <td style={{ padding: '14px 18px', fontWeight: 600, color: '#9B468A' }}>
                        {doc.bmdc_number}
                      </td>

                      <td style={{ padding: '14px 18px', maxWidth: '280px' }}>
                        <div>{doc.qualifications}</div>
                        <div style={{ fontSize: '12px', color: '#64748B' }}>{doc.workplace}</div>
                      </td>

                      <td style={{ padding: '14px 18px', fontWeight: 700 }}>
                        ৳{doc.consultation_fee}
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <span className={`badge ${doc.is_verified ? 'badge-green' : 'badge-red'}`}>
                          {doc.is_verified ? 'Verified' : 'Pending'}
                        </span>
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <span className={`badge ${doc.is_published ? 'badge-lime' : 'badge-gray'}`}>
                          {doc.is_published ? 'Published' : 'Hidden'}
                        </span>
                      </td>

                      <td style={{ padding: '14px 18px', textAlign: 'right' }}>
                        <div style={{ display: 'inline-flex', gap: '6px' }}>
                          <button
                            onClick={() => handleToggleVerify(doc.id)}
                            disabled={updatingId === doc.id}
                            className="btn btn-outline"
                            style={{ padding: '5px 10px', fontSize: '12px' }}
                          >
                            {doc.is_verified ? 'Unverify' : 'Verify BMDC'}
                          </button>
                          <button
                            onClick={() => handleTogglePublish(doc.id)}
                            disabled={updatingId === doc.id}
                            className="btn btn-outline"
                            style={{ padding: '5px 10px', fontSize: '12px' }}
                          >
                            {doc.is_published ? 'Unpublish' : 'Publish'}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* TAB 2: AUDIT LOGS */}
        {activeTab === 'audit' && (
          <div className="card-safeer" style={{ overflow: 'hidden', padding: 0 }}>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13px' }}>
                <thead style={{ backgroundColor: '#EEF7F8', borderBottom: '1px solid #DCEAEC', color: '#475569' }}>
                  <tr>
                    <th style={{ padding: '12px 16px' }}>Timestamp (UTC)</th>
                    <th style={{ padding: '12px 16px' }}>Action</th>
                    <th style={{ padding: '12px 16px' }}>Entity</th>
                    <th style={{ padding: '12px 16px' }}>Entity ID</th>
                    <th style={{ padding: '12px 16px' }}>IP Address</th>
                    <th style={{ padding: '12px 16px' }}>Audit Payload</th>
                  </tr>
                </thead>
                <tbody>
                  {auditLogs.map((log) => (
                    <tr key={log.id} style={{ borderBottom: '1px solid #EEF7F8' }}>
                      <td style={{ padding: '12px 16px', color: '#64748B', whiteSpace: 'nowrap' }}>
                        {new Date(log.created_at).toISOString().replace('T', ' ').slice(0, 19)}
                      </td>
                      <td style={{ padding: '12px 16px' }}>
                        <span className="badge badge-gray" style={{ fontWeight: 700 }}>
                          {log.action}
                        </span>
                      </td>
                      <td style={{ padding: '12px 16px', fontWeight: 600 }}>{log.entity_type}</td>
                      <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontSize: '11.5px', color: '#9B468A' }}>
                        {log.entity_id || '-'}
                      </td>
                      <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontSize: '11.5px', color: '#64748B' }}>
                        {log.ip_address || '127.0.0.1'}
                      </td>
                      <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontSize: '11.5px', color: '#334155', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {log.details_json}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* TAB 3: CATALOGUE & SERVICE ZONES */}
        {activeTab === 'catalogue' && (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
            <div className="card-safeer" style={{ padding: '24px' }}>
              <h3 style={{ fontSize: '16px', fontWeight: 800, marginBottom: '16px' }}>
                Service Pricing Catalogue
              </h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {catalogue?.services?.map((s: any) => (
                  <div key={s.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '12px', backgroundColor: '#EEF7F8', borderRadius: '8px' }}>
                    <div>
                      <strong style={{ fontSize: '14px', color: '#2A3F68' }}>{s.name}</strong>
                      <div style={{ fontSize: '12px', color: '#64748B' }}>{s.description}</div>
                    </div>
                    <span style={{ fontWeight: 800, color: '#33C2DF', fontSize: '16px' }}>
                      ৳{s.base_price} BDT
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className="card-safeer" style={{ padding: '24px' }}>
              <h3 style={{ fontSize: '16px', fontWeight: 800, marginBottom: '16px' }}>
                Dhaka Service Coverage Zones
              </h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {catalogue?.zones?.map((z: any) => (
                  <div key={z.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '12px', backgroundColor: '#EEF7F8', borderRadius: '8px' }}>
                    <div>
                      <strong style={{ fontSize: '14px', color: '#2A3F68' }}>{z.name}</strong>
                      <div style={{ fontSize: '12px', color: '#64748B' }}>Zone Code: {z.code}</div>
                    </div>
                    <span className="badge badge-green">ACTIVE (Dhaka)</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

