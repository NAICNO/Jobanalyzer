import { CellWithLink } from '../../../src/components/table/cell'
import { MemoryRouter } from 'react-router'
import { render } from '../../test-utils.tsx'

const mockValue = {
  text: 'Test Link',
  link: '/test-link'
}

describe('CellWithLink', () => {
  it('renders correctly', () => {
    const {asFragment} = render(
      <MemoryRouter>
        <CellWithLink value={mockValue}/>
      </MemoryRouter>
    )
    expect(asFragment()).toMatchSnapshot()
  })
})
