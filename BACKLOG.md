# Known Limitations and Future Backlog - Upasham Plus

## 1. Known Limitations (Current Pilot Scope)
- **Launch Area:** The platform is targeted at Dhaka Metropolitan area. Sub-district (Upazila) and nationwide physical sample collection are reserved for Phase 2.
- **Provider Settlements:** Doctor payouts remain manual initially via periodic reconciliation exports in accordance with Section 2 rules.
- **Consultation Media Recording:** Video/audio media is not stored on server disks during this pilot release in compliance with Section 10 data minimization rules.
- **Native Mobile Apps:** Out of initial scope (responsive web application works seamlessly on mobile browsers).

---

## 2. Dormant Foundations Ready for Expansion (Section 13)
The current database schema and Go backend contain pre-built, dormant model schemas ready for immediate commercial activation:
1. **Home Sample Collection (`sample_jobs`, `lab_partners`)**:
   - Phlebotomist dispatching across Dhaka North/South.
   - Lab integration for CBC, Lipid Profile, HbA1c, and Liver Function tests.
2. **Caregiver & Home Nurse Support (`home_service_jobs`)**:
   - Wound dressing, IV cannulation, catheter care, post-operative nursing.
3. **Bundled Health Packages & Subscriptions (`packages`, `subscriptions`)**:
   - Chronic diabetes monthly care packages.
   - Elderly home monitoring packages.
