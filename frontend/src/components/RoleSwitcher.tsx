import React from 'react';
import { User, Shield, Stethoscope, Briefcase, DollarSign, Globe, ChevronUp } from 'lucide-react';

interface RoleSwitcherProps { currentRole: string; onSwitchRole: (role: string) => void; lang: 'en' | 'bn'; }

export const RoleSwitcher: React.FC<RoleSwitcherProps> = ({ currentRole, onSwitchRole, lang }) => {
  const roles = [
    { id:'public', label: lang === 'bn' ? 'পাবলিক ওয়েবসাইট' : 'Public site', icon: Globe },
    { id:'patient', label: lang === 'bn' ? 'পেশেন্ট' : 'Patient', icon: User },
    { id:'doctor_jony', label: lang === 'bn' ? 'ডাঃ জনি' : 'Doctor', icon: Stethoscope },
    { id:'operations', label: lang === 'bn' ? 'অপারেশনস' : 'Operations', icon: Briefcase },
    { id:'finance', label: lang === 'bn' ? 'ফিন্যান্স' : 'Finance', icon: DollarSign },
    { id:'admin', label: lang === 'bn' ? 'এডমিন' : 'Admin', icon: Shield },
  ];
  const active = roles.find(r => currentRole.startsWith('doctor') ? r.id.startsWith('doctor') : r.id === currentRole);
  return <details className="workspace-switcher">
    <summary><span>Demo workspace</span><strong>{active?.label || 'Public site'}</strong><ChevronUp size={15}/></summary>
    <div className="workspace-menu">{roles.map(({id,label,icon:Icon}) => <button key={id} className={currentRole===id || (currentRole.startsWith('doctor')&&id.startsWith('doctor')) ? 'active' : ''} onClick={()=>onSwitchRole(id)}><Icon size={15}/><span>{label}</span></button>)}</div>
  </details>;
};
