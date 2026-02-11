import React from 'react';
import { Monitor, Server } from 'lucide-react';
import { OS } from '../types';

interface OSIconProps {
  os?: OS;
  className?: string;
}

const OSIcon: React.FC<OSIconProps> = ({ os, className = 'w-5 h-5' }) => {
  switch (os) {
    case 'linux':
      // Linux Tux penguin representation using Server icon with custom color
      return <Server className={className} strokeWidth={2} />;
    case 'windows':
      // Windows representation using Monitor icon
      return <Monitor className={className} strokeWidth={2} />;
    case 'darwin':
      // macOS representation using Monitor icon with different styling
      return <Monitor className={className} strokeWidth={2.5} />;
    default:
      // Unknown OS - generic server icon
      return <Server className={className} strokeWidth={2} />;
  }
};

export default OSIcon;
