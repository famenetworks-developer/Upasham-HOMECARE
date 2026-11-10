import React, { useState, useEffect } from 'react';
import { X, Mic, MicOff, Video, VideoOff, PhoneOff, ExternalLink, ShieldAlert, Clock, PhoneCall } from 'lucide-react';

interface JitsiVideoModalProps {
  roomId: string;
  onClose: () => void;
  userRole: string;
}

export const JitsiVideoModal: React.FC<JitsiVideoModalProps> = ({
  roomId,
  onClose,
  userRole,
}) => {
  const [roomData, setRoomData] = useState<any>(null);
  const [micOn, setMicOn] = useState<boolean>(true);
  const [camOn, setCamOn] = useState<boolean>(true);
  const [callDuration, setCallDuration] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMsg, setErrorMsg] = useState<string>('');

  useEffect(() => {
    const fetchAccess = async () => {
      try {
        const token = localStorage.getItem('upasham_token') || '';
        const res = await fetch(`/api/video/rooms/${roomId}/access`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to access consultation room');

        setRoomData(data);

        // Log join attendance
        await fetch('/api/video/attendance', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            video_room_id: roomId,
            event_type: 'join',
          }),
        });
      } catch (err: any) {
        setErrorMsg(err.message || 'Room access error');
      } finally {
        setLoading(false);
      }
    };

    fetchAccess();
  }, [roomId]);

  // Duration timer
  useEffect(() => {
    const timer = setInterval(() => {
      setCallDuration((prev) => prev + 1);
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  const handleEndCall = async () => {
    try {
      const token = localStorage.getItem('upasham_token') || '';
      await fetch('/api/video/attendance', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          video_room_id: roomId,
          event_type: 'leave',
        }),
      });
    } catch (e) {
      console.error(e);
    }
    onClose();
  };

  const formatTime = (secs: number) => {
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m < 10 ? '0' : ''}${m}:${s < 10 ? '0' : ''}${s}`;
  };

  if (loading) {
    return (
      <div style={{ position: 'fixed', inset: 0, backgroundColor: '#2A3F68', zIndex: 300, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#FFF' }}>
        <p>Connecting to Upasham Secure Telehealth Server...</p>
      </div>
    );
  }

  if (errorMsg) {
    return (
      <div style={{ position: 'fixed', inset: 0, backgroundColor: '#2A3F68', zIndex: 300, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '20px' }}>
        <div style={{ backgroundColor: '#1E293B', padding: '28px', borderRadius: '16px', maxWidth: '440px', color: '#FFF', textAlign: 'center' }}>
          <ShieldAlert size={40} color="#EF4444" style={{ margin: '0 auto 12px' }} />
          <h3 style={{ fontSize: '18px', fontWeight: 800, marginBottom: '8px' }}>Room Access Denied</h3>
          <p style={{ fontSize: '14px', color: '#94A3B8', marginBottom: '20px' }}>{errorMsg}</p>
          <button onClick={onClose} className="btn btn-outline" style={{ color: '#FFF', borderColor: '#475569' }}>
            Close Window
          </button>
        </div>
      </div>
    );
  }

  const jitsiWebUrl = `https://${roomData?.jitsi_domain || 'meet.jit.si'}/${roomData?.room_name}#config.prejoinPageEnabled=false&userInfo.displayName="${encodeURIComponent(userRole === 'doctor' ? roomData?.doctor_name : roomData?.patient_name)}"`;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        backgroundColor: '#2A3F68',
        zIndex: 300,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Top Telehealth Bar */}
      <div
        style={{
          padding: '12px 24px',
          backgroundColor: '#18181B',
          borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          color: '#FFF',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: '#22C55E', display: 'inline-block' }} />
            <span style={{ fontSize: '13px', fontWeight: 800, color: '#33C2DF' }}>
              UPASHAM PROTECTED TELECONSULTATION
            </span>
          </div>
          <span style={{ color: '#71717A', fontSize: '12px' }}>|</span>
          <span style={{ fontSize: '13px', color: '#E4E4E7' }}>
            Ref: <strong>{roomData?.booking_ref}</strong>
          </span>
          <span style={{ color: '#71717A', fontSize: '12px' }}>|</span>
          <span style={{ fontSize: '13px', color: '#A1A1AA' }}>
            Patient: <strong style={{ color: '#FFF' }}>{roomData?.patient_name}</strong>
          </span>
          {roomData?.doctor_name && (
            <>
              <span style={{ color: '#71717A', fontSize: '12px' }}>|</span>
              <span style={{ fontSize: '13px', color: '#A1A1AA' }}>
                Doctor: <strong style={{ color: '#FFF' }}>{roomData?.doctor_name}</strong>
              </span>
            </>
          )}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', color: '#33C2DF', fontWeight: 700 }}>
            <Clock size={14} />
            <span>{formatTime(callDuration)}</span>
          </div>

          <a
            href={jitsiWebUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="btn btn-outline"
            style={{
              color: '#33C2DF',
              borderColor: 'rgba(213, 255, 85, 0.4)',
              padding: '6px 12px',
              fontSize: '12px',
            }}
          >
            <ExternalLink size={13} />
            <span>Open in Full Jitsi Tab</span>
          </a>

          <button
            onClick={handleEndCall}
            style={{
              backgroundColor: '#DC2626',
              color: '#FFF',
              border: 'none',
              borderRadius: '8px',
              padding: '6px 14px',
              fontSize: '13px',
              fontWeight: 700,
              cursor: 'pointer',
            }}
          >
            End Call
          </button>
        </div>
      </div>

      {/* Main Video Stage */}
      <div style={{ flex: 1, position: 'relative', display: 'flex', backgroundColor: '#09090B' }}>
        {/* Real Embedded Jitsi Meet WebRTC IFrame */}
        <iframe
          src={jitsiWebUrl}
          allow="camera; microphone; fullscreen; display-capture"
          style={{
            width: '100%',
            height: '100%',
            border: 'none',
          }}
          title="Upasham Consultation Video Room"
        />

        {/* Low-Bandwidth / Emergency Fallback Bar */}
        <div
          style={{
            position: 'absolute',
            bottom: '76px',
            left: '50%',
            transform: 'translateX(-50%)',
            backgroundColor: 'rgba(24, 24, 27, 0.92)',
            backdropFilter: 'blur(8px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '999px',
            padding: '6px 18px',
            fontSize: '12px',
            color: '#E4E4E7',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            zIndex: 10,
          }}
        >
          <PhoneCall size={13} color="#22C55E" />
          <span>Low bandwidth or video drop? Emergency operations support:</span>
          <strong style={{ color: '#33C2DF' }}>+880 9610-UPASHAM (Dhaka HQ)</strong>
        </div>
      </div>

      {/* Bottom Floating Control Bar */}
      <div
        style={{
          padding: '14px 24px',
          backgroundColor: '#18181B',
          borderTop: '1px solid rgba(255, 255, 255, 0.1)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '16px',
        }}
      >
        <button
          onClick={() => setMicOn(!micOn)}
          style={{
            width: 44,
            height: 44,
            borderRadius: '50%',
            backgroundColor: micOn ? '#27272A' : '#DC2626',
            color: '#FFF',
            border: 'none',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
          }}
          title={micOn ? 'Mute Microphone' : 'Unmute Microphone'}
        >
          {micOn ? <Mic size={20} /> : <MicOff size={20} />}
        </button>

        <button
          onClick={() => setCamOn(!camOn)}
          style={{
            width: 44,
            height: 44,
            borderRadius: '50%',
            backgroundColor: camOn ? '#27272A' : '#DC2626',
            color: '#FFF',
            border: 'none',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
          }}
          title={camOn ? 'Turn Off Camera' : 'Turn On Camera'}
        >
          {camOn ? <Video size={20} /> : <VideoOff size={20} />}
        </button>

        <button
          onClick={handleEndCall}
          style={{
            height: 44,
            padding: '0 24px',
            borderRadius: '999px',
            backgroundColor: '#DC2626',
            color: '#FFF',
            border: 'none',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontWeight: 800,
            cursor: 'pointer',
          }}
        >
          <PhoneOff size={18} />
          <span>Leave Consultation</span>
        </button>
      </div>
    </div>
  );
};


