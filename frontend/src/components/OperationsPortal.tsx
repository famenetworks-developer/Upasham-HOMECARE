import React, { useState } from 'react';
import { Doctor } from '../types';
import { Briefcase, UserCheck, Calendar, Clock, CheckCircle2, AlertCircle, XCircle } from 'lucide-react';
import { Language } from '../utils/translations';

interface OperationsPortalProps {
  queue: any[];
  doctors: Doctor[];
  onRefresh: () => void;
  lang: Language;
}

export const OperationsPortal: React.FC<OperationsPortalProps> = ({
  queue,
  doctors,
  onRefresh,
  lang,
}) => {
  const [selectedBooking, setSelectedBooking] = useState<any>(null);
  const [modalType, setModalType] = useState<'assign' | 'confirm' | 'cancel' | null>(null);

  const [assignedDoctorId, setAssignedDoctorId] = useState<string>(doctors[0]?.id || '');
  const [assignNotes, setAssignNotes] = useState<string>('Assigned based on on-duty Dhaka availability');
  const [scheduledTime, setScheduledTime] = useState<string>(
    new Date(Date.now() + 30 * 60000).toISOString().slice(0, 16)
  );
  const [cancelReason, setCancelReason] = useState<string>('Patient requested refund or doctor unavailable');
  const [loading, setLoading] = useState<boolean>(false);

  const handleAssignDoctor = async () => {
    if (!selectedBooking) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/operations/assign', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          booking_id: selectedBooking.id,
          doctor_id: assignedDoctorId,
          notes: assignNotes,
        }),
      });

      if (res.ok) {
        setModalType(null);
        setSelectedBooking(null);
        onRefresh();
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleConfirmAppointment = async () => {
    if (!selectedBooking) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/operations/confirm', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          booking_id: selectedBooking.id,
          scheduled_at: scheduledTime,
          notes: 'Slot confirmed and Jitsi consultation room provisioned',
        }),
      });

      if (res.ok) {
        setModalType(null);
        setSelectedBooking(null);
        onRefresh();
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async () => {
    if (!selectedBooking) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/operations/cancel', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          booking_id: selectedBooking.id,
          reason: cancelReason,
        }),
      });

      if (res.ok) {
        setModalType(null);
        setSelectedBooking(null);
        onRefresh();
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const awaitingCount = queue.filter((q) => q.status === 'Paid/Awaiting Assignment').length;

  return (
    <div style={{ padding: '36px 0', backgroundColor: '#F7FBFC', minHeight: '80vh' }}>
      <div className="container">
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <span style={{ fontSize: '12px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
                DHAKA OPERATIONS CONTROL
              </span>
              {awaitingCount > 0 && (
                <span className="badge badge-yellow animate-pulse-live">
                  {awaitingCount} Awaiting Doctor Assignment
                </span>
              )}
            </div>
            <h1 style={{ fontSize: '28px', fontWeight: 800, color: '#2A3F68' }}>
              Operations Manual Assignment & Schedule Queue
            </h1>
          </div>

          <div style={{ display: 'flex', gap: '10px' }}>
            <button onClick={onRefresh} className="btn btn-outline" style={{ fontSize: '13px' }}>
              Refresh Queue
            </button>
          </div>
        </div>

        {/* Queue Table */}
        <div className="card-safeer" style={{ overflow: 'hidden', padding: 0 }}>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
              <thead style={{ backgroundColor: '#EEF7F8', borderBottom: '1px solid #DCEAEC', color: '#475569' }}>
                <tr>
                  <th style={{ padding: '14px 18px' }}>Booking Ref</th>
                  <th style={{ padding: '14px 18px' }}>Patient Info</th>
                  <th style={{ padding: '14px 18px' }}>Chief Complaint</th>
                  <th style={{ padding: '14px 18px' }}>Payment & Gateway</th>
                  <th style={{ padding: '14px 18px' }}>Assigned Doctor</th>
                  <th style={{ padding: '14px 18px' }}>Status</th>
                  <th style={{ padding: '14px 18px', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {queue.length === 0 ? (
                  <tr>
                    <td colSpan={7} style={{ padding: '40px', textAlign: 'center', color: '#64748B' }}>
                      No active requests in operations queue.
                    </td>
                  </tr>
                ) : (
                  queue.map((item) => (
                    <tr key={item.id} style={{ borderBottom: '1px solid #EEF7F8' }}>
                      <td style={{ padding: '14px 18px', fontWeight: 700, color: '#2A3F68' }}>
                        {item.booking_ref}
                        <div style={{ fontSize: '11px', color: '#64748B', fontWeight: 400 }}>
                          {item.service_type === 'general' ? '20৳ General Triage' : 'Specialist Consult'}
                        </div>
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <strong>{item.patient_name}</strong>
                        <div style={{ fontSize: '12px', color: '#64748B' }}>
                          {item.patient_age} yrs • {item.relationship}
                        </div>
                      </td>

                      <td style={{ padding: '14px 18px', maxWidth: '240px' }}>
                        <div style={{ textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }}>
                          {item.chief_complaint}
                        </div>
                        <div style={{ fontSize: '11px', color: '#33C2DF', fontWeight: 600 }}>
                          Pref: {item.preferred_time}
                        </div>
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <span style={{ fontWeight: 700, color: '#33C2DF' }}>৳{item.paid_amount} BDT</span>
                        <div style={{ fontSize: '11px', color: '#64748B' }}>
                          {item.gateway?.toUpperCase()} ({item.gateway_tx_id})
                        </div>
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        {item.doctor_name ? (
                          <span style={{ fontWeight: 600, color: '#2A3F68' }}>{item.doctor_name}</span>
                        ) : (
                          <span style={{ color: '#DC2626', fontWeight: 600, fontSize: '12px' }}>
                            Unassigned
                          </span>
                        )}
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <span
                          className={`badge ${
                            item.status === 'Paid/Awaiting Assignment'
                              ? 'badge-yellow'
                              : item.status === 'Assigned'
                              ? 'badge-blue'
                              : item.status === 'Confirmed'
                              ? 'badge-green'
                              : 'badge-gray'
                          }`}
                        >
                          {item.status}
                        </span>
                      </td>

                      <td style={{ padding: '14px 18px', textAlign: 'right' }}>
                        <div style={{ display: 'inline-flex', gap: '6px' }}>
                          {item.status === 'Paid/Awaiting Assignment' && (
                            <button
                              onClick={() => {
                                setSelectedBooking(item);
                                setModalType('assign');
                              }}
                              className="btn btn-lime"
                              style={{ padding: '6px 12px', fontSize: '12px', fontWeight: 800 }}
                            >
                              <UserCheck size={13} /> Assign Doctor
                            </button>
                          )}

                          {item.status === 'Assigned' && (
                            <button
                              onClick={() => {
                                setSelectedBooking(item);
                                setModalType('confirm');
                              }}
                              className="btn btn-green"
                              style={{ padding: '6px 12px', fontSize: '12px', fontWeight: 700 }}
                            >
                              <Calendar size={13} /> Confirm Slot
                            </button>
                          )}

                          <button
                            onClick={() => {
                              setSelectedBooking(item);
                              setModalType('cancel');
                            }}
                            className="btn btn-outline"
                            style={{ padding: '6px 10px', fontSize: '12px', color: '#DC2626' }}
                            title="Cancel / Refund"
                          >
                            Cancel
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Modal: Manual Assign Doctor */}
        {modalType === 'assign' && selectedBooking && (
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
            <div className="card-safeer animate-fade-in" style={{ width: '100%', maxWidth: '480px', padding: '28px', backgroundColor: '#FFF' }}>
              <h3 style={{ fontSize: '18px', fontWeight: 800, marginBottom: '6px' }}>
                Assign Physician: {selectedBooking.booking_ref}
              </h3>
              <p style={{ fontSize: '13px', color: '#64748B', marginBottom: '18px' }}>
                Patient: {selectedBooking.patient_name} • Complaint: {selectedBooking.chief_complaint}
              </p>

              <div style={{ marginBottom: '16px' }}>
                <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '6px' }}>
                  Select Verified Doctor:
                </label>
                <select
                  value={assignedDoctorId}
                  onChange={(e) => setAssignedDoctorId(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', border: '1px solid #C7D8DE', backgroundColor: '#FFF' }}
                >
                  {doctors.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.full_name} — {d.qualifications} ({d.workplace})
                    </option>
                  ))}
                </select>
              </div>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '6px' }}>
                  Operations Assignment Note:
                </label>
                <input
                  type="text"
                  value={assignNotes}
                  onChange={(e) => setAssignNotes(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
                <button onClick={() => setModalType(null)} className="btn btn-outline">
                  Cancel
                </button>
                <button onClick={handleAssignDoctor} disabled={loading} className="btn btn-lime" style={{ fontWeight: 800 }}>
                  {loading ? 'Assigning...' : 'Confirm Assignment'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Modal: Confirm Appointment */}
        {modalType === 'confirm' && selectedBooking && (
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
            <div className="card-safeer animate-fade-in" style={{ width: '100%', maxWidth: '480px', padding: '28px', backgroundColor: '#FFF' }}>
              <h3 style={{ fontSize: '18px', fontWeight: 800, marginBottom: '6px' }}>
                Confirm Slot & Provision Jitsi Meeting
              </h3>
              <p style={{ fontSize: '13px', color: '#64748B', marginBottom: '18px' }}>
                Booking: {selectedBooking.booking_ref} • Doctor: {selectedBooking.doctor_name}
              </p>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '6px' }}>
                  Scheduled Consultation Date & Time (Dhaka):
                </label>
                <input
                  type="datetime-local"
                  value={scheduledTime}
                  onChange={(e) => setScheduledTime(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
                <button onClick={() => setModalType(null)} className="btn btn-outline">
                  Cancel
                </button>
                <button onClick={handleConfirmAppointment} disabled={loading} className="btn btn-green" style={{ fontWeight: 800 }}>
                  {loading ? 'Confirming...' : 'Provision Video Room & Confirm'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Modal: Cancel Booking */}
        {modalType === 'cancel' && selectedBooking && (
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
            <div className="card-safeer animate-fade-in" style={{ width: '100%', maxWidth: '440px', padding: '28px', backgroundColor: '#FFF' }}>
              <h3 style={{ fontSize: '18px', fontWeight: 800, color: '#DC2626', marginBottom: '6px' }}>
                Cancel Consultation
              </h3>
              <p style={{ fontSize: '13px', color: '#64748B', marginBottom: '18px' }}>
                Booking Ref: {selectedBooking.booking_ref}
              </p>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '6px' }}>
                  Reason for Cancellation:
                </label>
                <input
                  type="text"
                  value={cancelReason}
                  onChange={(e) => setCancelReason(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
                <button onClick={() => setModalType(null)} className="btn btn-outline">
                  Back
                </button>
                <button onClick={handleCancel} disabled={loading} className="btn btn-primary" style={{ backgroundColor: '#DC2626', borderColor: '#DC2626' }}>
                  {loading ? 'Processing...' : 'Confirm Cancel'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

