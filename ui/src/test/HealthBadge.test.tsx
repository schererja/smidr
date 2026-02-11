import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import HealthBadge from '../components/HealthBadge';
import type { HealthState } from '../types';

describe('HealthBadge', () => {
  it('should render learning state correctly', () => {
    render(<HealthBadge state="learning" />);
    expect(screen.getByText('learning')).toBeInTheDocument();
  });

  it('should render healthy state correctly', () => {
    render(<HealthBadge state="healthy" />);
    expect(screen.getByText('healthy')).toBeInTheDocument();
  });

  it('should render degraded state correctly', () => {
    render(<HealthBadge state="degraded" />);
    expect(screen.getByText('degraded')).toBeInTheDocument();
  });

  it('should render attention state correctly', () => {
    render(<HealthBadge state="attention" />);
    expect(screen.getByText('attention')).toBeInTheDocument();
  });

  it('should render unknown state correctly', () => {
    render(<HealthBadge state="unknown" />);
    expect(screen.getByText('unknown')).toBeInTheDocument();
  });

  it('should apply correct size classes', () => {
    const { rerender } = render(<HealthBadge state="healthy" size="small" />);
    expect(screen.getByText('healthy')).toBeInTheDocument();

    rerender(<HealthBadge state="healthy" size="medium" />);
    expect(screen.getByText('healthy')).toBeInTheDocument();

    rerender(<HealthBadge state="healthy" size="large" />);
    expect(screen.getByText('healthy')).toBeInTheDocument();
  });

  it('should default to medium size', () => {
    render(<HealthBadge state="healthy" />);
    expect(screen.getByText('healthy')).toBeInTheDocument();
  });
});
