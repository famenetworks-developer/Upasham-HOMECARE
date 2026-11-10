import React from 'react';
import { ArrowRight, Video } from 'lucide-react';
import { Language } from '../utils/translations';
import { defaultSiteContent, local, SiteContent } from '../content';
import heroCare from '../assets/upasham-hero-care.png';
interface Props { lang: Language; content?:SiteContent; onOpenBookingGeneral:()=>void; onScrollToSpecialists:()=>void }
export const PublicHero:React.FC<Props>=({lang,content=defaultSiteContent,onOpenBookingGeneral,onScrollToSpecialists})=>{const h=content.hero;return <section className="shukhee-hero" style={{background:content.theme.heroBackground}}><div className="container hero-grid"><div className="hero-copy animate-fade-in"><h1>{local(h.title,lang)}</h1><p>{local(h.subtitle,lang)}</p><div className="hero-actions"><button onClick={onOpenBookingGeneral} className="btn btn-deep-green">{local(h.primaryButton,lang)} <span className="price-chip">{h.price}</span><ArrowRight size={18}/></button><button onClick={onScrollToSpecialists} className="btn btn-white"><Video size={18}/>{local(h.secondaryButton,lang)}</button></div></div><div className="hero-visual"><img src={h.imageUrl||heroCare} alt={local(h.title,lang)}/></div></div></section>}
