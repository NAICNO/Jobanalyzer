import { AccessorKeyColumnDef, createColumnHelper } from '@tanstack/react-table'
import { DASHBOARD_COLUMN } from '../Constants.ts'

export const getTableColumns = (selectedCluster: Cluster) => {
  const columns: AccessorKeyColumnDef<DashboardTableItem, any>[] = [
    createColumn('hostname'),
  ]

  if (selectedCluster.uptime) {
    columns.push(
      createColumn('cpu_status'),
      createColumn('gpu_status')
    )
  }

  columns.push(
    // Unique users in the period.  This will never be greater than jobs; a user can have
    // several jobs, but not zero, and jobs can only have one user.
    createColumn('users_recent'),
    createColumn('users_longer'),

    // Unique jobs running within the period.
    createColumn('jobs_recent'),
    createColumn('jobs_longer'),

    // Relative to system information.
    createColumn('cpu_recent'),
    createColumn('cpu_longer'),
    createColumn('resident_recent'),
    createColumn('resident_longer'),
    createColumn('mem_recent'),
    createColumn('mem_longer'),
    createColumn('gpu_recent'),
    createColumn('gpu_longer'),
    createColumn('gpumem_recent'),
    createColumn('gpumem_longer'),
  )

  // Number of *new* violators and zombies encountered in the period, as of the last
  // generated report.  This currently changes rarely.
  if (selectedCluster.violators) {
    columns.push(
      createColumn('violators_long')
    )
  }

  if (selectedCluster.deadweight) {
    columns.push(
      createColumn('zombies_long')
    )
  }
  return columns
}

const columnHelper = createColumnHelper<DashboardTableItem>()

function createColumn<K extends keyof DashboardTableItem>(key: K) {
  // Ensure that column definition exists in the constants
  const columnDef = DASHBOARD_COLUMN[key]
  if (!columnDef) {
    throw new Error(`Column definition for key '${key}' not found.`)
  }

  // Accessor and Header are always used, but other properties like cell and meta are added as needed
  return columnHelper.accessor(key, {
    cell: info => info.getValue(),
    header: columnDef.title,
    meta: columnDef,
  })
}

