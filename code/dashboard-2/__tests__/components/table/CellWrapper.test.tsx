import { CellWrapper } from '../../../src/components/table/cell'
import { render } from '../../test-utils.tsx'

describe('CellWrapper', () => {
  it('renders correctly', () => {
    const {asFragment} = render(<CellWrapper><></>
    </CellWrapper>)
    expect(asFragment()).toMatchSnapshot()
  })
})
