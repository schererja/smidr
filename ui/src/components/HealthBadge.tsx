import React from 'react';
import { HealthState } from '../types';
import { Badge } from '@/ui/badge';
import { cn } from '@/lib/utils';

interface HealthBadgeProps {
  state: HealthState;
  size?: 'small' | 'medium' | 'large';
}

const HealthBadge: React.FC<HealthBadgeProps> = ({ state, size = 'medium' }) => {
  const sizeClasses = {
    small: 'text-xs px-2 py-0.5',
    medium: 'text-sm px-3 py-1',
    large: 'text-base px-4 py-1.5',
  };

  const stateClasses = {
    learning: 'bg-blue-100 text-blue-800 border-blue-300',
    healthy: 'bg-green-100 text-green-800 border-green-300',
    degraded: 'bg-orange-100 text-orange-800 border-orange-300',
    attention: 'bg-red-100 text-red-800 border-red-300',
    unknown: 'bg-gray-100 text-gray-800 border-gray-300',
  };

  return (
    <Badge
      variant="outline"
      className={cn(
        sizeClasses[size],
        stateClasses[state],
        'font-semibold uppercase'
      )}
    >
      {state}
    </Badge>
  );
};

export default HealthBadge;
