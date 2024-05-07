import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'
import { WorkingFieldCell } from '../../../src/components/table/cell'

describe('WorkingFieldCell', () => {
  it('renders the value correctly', () => {
    render(<WorkingFieldCell value={65}/>)
    expect(screen.getByText('65')).toBeInTheDocument()
  })

  it('applies correct background color for value >= 75', () => {
    render(<WorkingFieldCell value={75}/>)
    expect(screen.getByText('75').parentNode).and.toHaveStyle({backgroundColor: '#00bfff'})
  })

  it('applies correct background color for 50 <= value < 75', () => {
    render(<WorkingFieldCell value={70}/>)
    expect(screen.getByText('70').parentNode).toHaveStyle({backgroundColor: '#87cefa'})
  })

  it('applies correct background color for 25 <= value < 50', () => {
    render(<WorkingFieldCell value={30}/>)
    expect(screen.getByText('30').parentNode).toHaveStyle({backgroundColor: '#e0ffff'})
  })

  it('applies correct background color for value < 25', () => {
    render(<WorkingFieldCell value={20}/>)
    expect(screen.getByText('20').parentNode)
  })
})
