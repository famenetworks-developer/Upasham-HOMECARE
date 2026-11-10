import React from 'react';
import officialLogo from '../assets/upasham-plus-logo-english.png';

interface LogoProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showTagline?: boolean;
}

const logoWidths = {
  sm: 96,
  md: 138,
  lg: 210,
};

/** The official Upasham Plus brand artwork supplied by the client. */
export const Logo: React.FC<LogoProps> = ({
  className = '',
  size = 'md',
  showTagline = true,
}) => (
  <span
    className={`official-logo ${className}`.trim()}
    data-tagline={showTagline ? 'visible' : 'embedded'}
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      flexShrink: 0,
      lineHeight: 0,
    }}
  >
    <img
      src={officialLogo}
      alt="Upasham Plus — Doctors from far, not so far"
      width={logoWidths[size]}
      height="auto"
      decoding="async"
      style={{ display: 'block', width: logoWidths[size], maxWidth: '100%', height: 'auto' }}
    />
  </span>
);
