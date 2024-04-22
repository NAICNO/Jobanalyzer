import '@testing-library/jest-dom'
import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { GpuFieldCell } from '../../../src/components/table/GpuFieldCell'

describe('GpuFieldCell', () => {
  it('renders correctly when value is not 0 or undefined', () => {
    const value = 1;
    const { asFragment } = render(<GpuFieldCell value={value} />);
    expect(asFragment()).toMatchSnapshot();
  });

  it('renders correctly when value is 0', () => {
    const value = 0;
    const { asFragment } = render(<GpuFieldCell value={value} />);
    expect(asFragment()).toMatchSnapshot();
  });

  it('renders correctly when value is undefined', () => {
    const value = undefined;
    const { asFragment } = render(<GpuFieldCell value={value} />);
    expect(asFragment()).toMatchSnapshot();
  });
});
