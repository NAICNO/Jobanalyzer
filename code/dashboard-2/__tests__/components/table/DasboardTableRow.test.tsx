import '@testing-library/jest-dom'
import { BrowserRouter } from 'react-router'
import { screen, within } from '@testing-library/react'
import { getCoreRowModel, useReactTable } from '@tanstack/react-table'

import { render } from '../../test-utils.tsx'
import { Cluster, DashboardTableItem } from '../../../src/types'
import { DashboardTable } from '../../../src/components/table'
import { getDashboardTableColumns } from '../../../src/util/TableUtils.ts'

const DashboardTableWrapper = ({data, cluster}: { data: DashboardTableItem[], cluster: Cluster }) => {
  const tableColumns = getDashboardTableColumns(cluster)

  const table = useReactTable({
    columns: tableColumns,
    data: data,
    getCoreRowModel: getCoreRowModel(),
  })

  return <DashboardTable table={table} cluster={cluster}/>
}

describe('DashboardTableRow', () => {
  it('applies correct style when cluster.uptime is true and cpuStatusCell value is not 0', () => {

    const data: DashboardTableItem[] = [
      {
        hostname: {text: 'ml1.hpc.uio.no', link: '/test/ml1.hpc.uio.no'},
        tag: 'ML Nodes',
        machine: '2x14 Intel Xeon Gold 5120 (hyperthreaded), 128GB, 4x NVIDIA RTX 2080 Ti @ 11GB',
        recent: 30,
        longer: 720,
        long: 1440,
        cpu_status: 1,
        gpu_status: 1,
        jobs_recent: 3,
        jobs_longer: 3,
        users_recent: 3,
        users_longer: 3,
        cpu_recent: 0,
        cpu_longer: 0,
        mem_recent: 38,
        mem_longer: 38,
        resident_recent: 0,
        resident_longer: 0,
        violators_long: 0,
        zombies_long: 0,
        gpu_recent: 0,
        gpu_longer: 0,
        gpumem_recent: 0,
        gpumem_longer: 0
      },
    ]

    const cluster: Cluster = {
      cluster: 'test',
      canonical: 'test',
      subclusters: [],
      uptime: true,
      violators: true,
      deadweight: true,
      defaultQuery: '*',
      hasDowntime: true,
      name: 'Test Cluster',
      description: 'Test Description',
      prefix: 'test',
      policy: 'test',
    }

    render(
      <BrowserRouter>
        <DashboardTableWrapper data={data} cluster={cluster}/>
      </BrowserRouter>
    )

    const table = screen.getByRole('table')
    const rowElements = within(table).getAllByRole('row')
    const dataRow = rowElements[2]
    expect(dataRow).toHaveStyle({backgroundColor: '#ff6347'})  // tomato
  })
})
