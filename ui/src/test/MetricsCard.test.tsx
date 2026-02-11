import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MetricsCard from '../components/MetricsCard';
import type { Signals, Baseline } from '../types';

describe('MetricsCard', () => {
  const mockSignals: Signals = {
    uptimeSeconds: 86400, // 1 day
    loadAverage1m: 1.5,
    memoryUsedPct: 50.5,
    diskUsedPct: 75.2,
    processCount: 123,
  };

  const mockBaselines: Baseline[] = [
    {
      metric: 'loadAverage1m',
      mean: 1.2,
      stdDev: 0.3,
      min: 0.5,
      max: 2.0,
    },
    {
      metric: 'memoryUsedPct',
      mean: 48.0,
      stdDev: 5.0,
      min: 40,
      max: 60,
    },
  ];

  it('should render "No signal data available" when signals is undefined', () => {
    render(<MetricsCard />);
    expect(screen.getByText('No signal data available')).toBeInTheDocument();
  });

  it('should render all signal metrics when signals provided', () => {
    render(<MetricsCard signals={mockSignals} />);

    expect(screen.getByText('Uptime')).toBeInTheDocument();
    expect(screen.getByText('1d 0h 0m')).toBeInTheDocument();

    expect(screen.getByText('Load Average (1m)')).toBeInTheDocument();
    expect(screen.getByText('1.50')).toBeInTheDocument();

    expect(screen.getByText('Memory Used')).toBeInTheDocument();
    expect(screen.getByText('50.5%')).toBeInTheDocument();

    expect(screen.getByText('Disk Used')).toBeInTheDocument();
    expect(screen.getByText('75.2%')).toBeInTheDocument();

    expect(screen.getByText('Process Count')).toBeInTheDocument();
    expect(screen.getByText('123')).toBeInTheDocument();
  });

  it('should format uptime correctly', () => {
    const signals = { ...mockSignals, uptimeSeconds: 90061 }; // 1d 1h 1m 1s
    render(<MetricsCard signals={signals} />);
    expect(screen.getByText('1d 1h 1m')).toBeInTheDocument();
  });

  it('should display delta from baseline when baselines provided', () => {
    render(<MetricsCard signals={mockSignals} baselines={mockBaselines} />);

    // Load average: 1.5 vs baseline 1.2 = +0.3
    expect(screen.getByText(/\+0\.3 from baseline/)).toBeInTheDocument();

    // Memory: 50.5 vs baseline 48.0 = +2.5
    expect(screen.getByText(/\+2\.5 from baseline/)).toBeInTheDocument();
  });

  it('should not display delta when no baselines provided', () => {
    render(<MetricsCard signals={mockSignals} />);
    
    expect(screen.queryByText(/from baseline/)).not.toBeInTheDocument();
  });

  it('should handle zero uptime', () => {
    const signals = { ...mockSignals, uptimeSeconds: 0 };
    render(<MetricsCard signals={signals} />);
    expect(screen.getByText('0d 0h 0m')).toBeInTheDocument();
  });

  it('should render negative delta with minus sign', () => {
    const signals = { ...mockSignals, loadAverage1m: 0.8 };
    render(<MetricsCard signals={signals} baselines={mockBaselines} />);

    // 0.8 vs baseline 1.2 = -0.4
    expect(screen.getByText(/-0\.4 from baseline/)).toBeInTheDocument();
  });
});
