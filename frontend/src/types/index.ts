export type Role = 'patient' | 'doctor' | 'operations' | 'finance' | 'admin' | 'public';

export interface User {
  id: string;
  email: string;
  phone: string;
  full_name: string;
  role: Role;
  patient_account_id?: string;
  doctor_id?: string;
}

export interface PatientProfile {
  id: string;
  account_id: string;
  full_name: string;
  relationship: 'self' | 'child' | 'spouse' | 'parent' | 'other';
  gender: string;
  age: number;
  blood_group: string;
  known_allergies: string;
  chronic_conditions: string;
  notes: string;
  created_at: string;
}

export interface Doctor {
  id: string;
  user_id?: string;
  full_name: string;
  bmdc_number: string;
  specialties: string[];
  qualifications: string;
  experience_years: number;
  workplace: string;
  languages: string[];
  bio: string;
  bio_bn?: string;
  consultation_fee: number;
  is_verified: boolean;
  is_published: boolean;
  avatar_url: string;
  rating: number;
  total_reviews: number;
  availability?: string;
}

export interface Specialty {
  id: string;
  name: string;
  name_bn: string;
  slug: string;
  icon: string;
  description: string;
  description_bn?: string;
}

export interface Booking {
  id: string;
  booking_ref: string;
  patient_account_id: string;
  patient_profile_id: string;
  doctor_id?: string;
  specialty_id?: string;
  service_type: 'general' | 'specialist';
  status:
    | 'Draft'
    | 'Payment Pending'
    | 'Paid/Awaiting Assignment'
    | 'Assigned'
    | 'Confirmed'
    | 'In Progress'
    | 'Completed'
    | 'Cancelled'
    | 'No-show'
    | 'Refund Pending'
    | 'Refunded';
  chief_complaint: string;
  triage_notes?: string;
  preferred_time: string;
  scheduled_at?: string;
  quoted_amount: number;
  currency: string;
  payment_id?: string;
  video_room_id?: string;
  created_at: string;
  updated_at?: string;

  // Joined display attributes
  patient_name?: string;
  relationship?: string;
  doctor_name?: string;
  doctor_workplace?: string;
  specialty_name?: string;
}

export interface MedicineLine {
  id?: string;
  brand_name: string;
  generic_name?: string;
  strength?: string;
  dosage_form?: string;
  dose: string;
  route?: string;
  frequency: string;
  duration: string;
  instructions?: string;
}

export interface TestOrder {
  id?: string;
  test_name: string;
  instructions: string;
  status?: string;
}

export interface Prescription {
  id: string;
  booking_id: string;
  encounter_id: string;
  prescription_ref: string;
  doctor_id: string;
  patient_profile_id: string;
  diagnosis: string;
  advice: string;
  follow_up_date: string;
  warnings: string;
  verification_hash: string;
  version: number;
  created_at: string;
  doctor_name?: string;
  doctor_bmdc?: string;
  doctor_degrees?: string;
  doctor_workplace?: string;
  patient_name?: string;
  patient_age?: number;
  patient_gender?: string;
  booking_ref?: string;
  medicines?: MedicineLine[];
  tests?: TestOrder[];
}

export interface PaymentTransaction {
  id: string;
  booking_id: string;
  invoice_ref: string;
  gateway: 'bkash' | 'sslcommerz';
  amount: number;
  currency: string;
  status: string;
  gateway_tx_id?: string;
  paid_at?: string;
  created_at: string;
  patient_name?: string;
  service_type?: string;
  booking_ref?: string;
}

export interface AuditEvent {
  id: string;
  user_id?: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  ip_address?: string;
  details_json: string;
  created_at: string;
}
