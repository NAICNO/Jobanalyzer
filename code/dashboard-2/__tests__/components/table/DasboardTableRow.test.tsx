import '@testing-library/jest-dom'
import { render, screen, within } from '@testing-library/react'
import DashboardTableRow from '../../../src/components/table/DashboardTableRow'
import { Table, Tbody } from '@chakra-ui/react'


describe('DashboardTableRow', () => {
  it('applies correct style when cluster.uptime is true and cpuStatusCell value is not 0', () => {
    const row = {
      id: '1',
      getAllCells: () => [
        {
          id: 'cell1',
          column: {id: 'cpu_status', columnDef: {meta: {isNumeric: true}, cell: () => 'Cell 1'}},
          getValue: () => 1,
          getContext: () => {
          }
        },
        {
          id: 'cell2',
          column: {id: 'memory_status', columnDef: {meta: {isNumeric: false}, cell: () => 'Cell 2'}},
          getValue: () => 0,
          getContext: () => {
          }
        }
      ]
    }
    const cluster = {uptime: true}
    render(
      <Table>
        <Tbody>
          <DashboardTableRow row={row} cluster={cluster}/>
        </Tbody>
      </Table>
    )

    const table = screen.getByRole('table')
    const rowElement = within(table).getByRole('row')
    expect(rowElement).toHaveStyle({backgroundColor: '#ff6347'})  // tomato
  })
})
