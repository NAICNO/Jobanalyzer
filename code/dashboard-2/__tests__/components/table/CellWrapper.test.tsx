import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import CellWrapper from '../../../src/components/table/CellWrapper'

describe('CellWrapper', () => {
  it('renders correctly', () => {
    const { asFragment } = render(<CellWrapper><></></CellWrapper>);
    expect(asFragment()).toMatchSnapshot();
  });
});
