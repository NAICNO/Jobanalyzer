import {
  AccessorKeyColumnDef,
  createColumnHelper,
  GroupColumnDef,
  SortingFn
} from '@tanstack/react-table'

import {
  DASHBOARD_COLUMN,
  DEAD_WEIGHT_COLUMN, DURATION_REGEX, JOB_QUERY_RESULTS_COLUMN,
  VIOLATING_JOB_SUMMARY_COLUMN,
  VIOLATING_USER_SUMMARY_COLUMN
} from '../Constants.ts'
import {
  Cluster,
  DashboardTableItem,
  DeadWeightTableItem, JobQueryResultsTableItem,
  ViolatingJobTableItem,
  ViolatingUserTableItem,
} from '../types'

export const getDashboardTableColumns = (selectedCluster: Cluster) => {
  const columns: (AccessorKeyColumnDef<DashboardTableItem, any> | GroupColumnDef<DashboardTableItem>)[] = [
    createDashboardTableColumn('hostname'),
  ]

  const addGroupedColumns = (id: string, header: string, keys: (keyof DashboardTableItem)[]) => {
    columns.push(dashboardTableColumnHelper.group({
      id,
      header,
      columns: keys.map(key => createDashboardTableColumn(key, true)),
    }))
  }

  if (selectedCluster.uptime) {
    addGroupedColumns('uptime', 'Uptime', ['cpu_status', 'gpu_status'])
  }

  // Unique users in the period.  This will never be greater than jobs; a user can have
  // several jobs, but not zero, and jobs can only have one user.
  addGroupedColumns('users', 'User', ['users_recent', 'users_longer'])

  // Unique jobs running within the period.
  addGroupedColumns('jobs', 'Jobs', ['jobs_recent', 'jobs_longer'])

  // Relative to system information.
  addGroupedColumns('cpu', 'CPU %', ['cpu_recent', 'cpu_longer'])
  addGroupedColumns('resident', 'Resident %', ['resident_recent', 'resident_longer'])
  addGroupedColumns('mem', 'Virt %', ['mem_recent', 'mem_longer'])
  addGroupedColumns('gpu', 'GPU %', ['gpu_recent', 'gpu_longer'])
  addGroupedColumns('gpu_mem', 'GPU Mem %', ['gpumem_recent', 'gpumem_longer'])

  // Number of *new* violators and zombies encountered in the period, as of the last
  // generated report.  This currently changes rarely.
  if (selectedCluster.violators) {
    columns.push(
      createDashboardTableColumn('violators_long')
    )
  }

  if (selectedCluster.deadweight) {
    columns.push(
      createDashboardTableColumn('zombies_long')
    )
  }
  return columns
}

const dashboardTableColumnHelper = createColumnHelper<DashboardTableItem>()

function createDashboardTableColumn<K extends keyof DashboardTableItem>(key: K, isGrouped = false) {
  // Ensure that column definition exists in the constants
  const columnDef = DASHBOARD_COLUMN[key]
  if (!columnDef) {
    throw new Error(`Column definition for key '${key}' not found.`)
  }

  const header = isGrouped ? (columnDef.shortTitle ?? columnDef.title) : columnDef.title

  // Accessor and Header are always used, but other properties like cell and meta are added as needed
  return dashboardTableColumnHelper.accessor(key, {
    cell: props => {
      if (columnDef.renderFn) {
        return columnDef.renderFn({value: props.getValue()})
      }
      return props.getValue()
    },
    header: header,
    meta: columnDef,
  })
}

export const getViolatingUserTableColumns = () => {
  const columns: AccessorKeyColumnDef<ViolatingUserTableItem, any>[] = [
    createViolatingUserTableColumn('user'),
    createViolatingUserTableColumn('count'),
    createViolatingUserTableColumn('latest'),
    createViolatingUserTableColumn('earliest'),
  ]
  return columns
}

const violatingUserTableColumnHelper = createColumnHelper<ViolatingUserTableItem>()

function createViolatingUserTableColumn<K extends keyof ViolatingUserTableItem>(key: K) {
  // Ensure that column definition exists in the constants
  const columnDef = VIOLATING_USER_SUMMARY_COLUMN[key]
  if (!columnDef) {
    throw new Error(`Column definition for key '${key}' not found.`)
  }

  // Accessor and Header are always used, but other properties like cell and meta are added as needed
  return violatingUserTableColumnHelper.accessor(key, {
    cell: props => {
      if (columnDef.renderFn) {
        return columnDef.renderFn({value: props.getValue()})
      }
      return props.getValue()
    },
    header: columnDef.title,
    meta: columnDef,
  })
}

export const getViolatingJobTableColumns = () => {
  const columns: AccessorKeyColumnDef<ViolatingJobTableItem, any>[] = [
    createViolatingJobTableColumn('hostname'),
    createViolatingJobTableColumn('user'),
    createViolatingJobTableColumn('id'),
    createViolatingJobTableColumn('cmd'),
    createViolatingJobTableColumn('started-on-or-before'),
    createViolatingJobTableColumn('last-seen'),
    createViolatingJobTableColumn('cpu-peak'),
    createViolatingJobTableColumn('rcpu-avg'),
    createViolatingJobTableColumn('rcpu-peak'),
    createViolatingJobTableColumn('rmem-avg'),
    createViolatingJobTableColumn('rmem-peak'),
  ]
  return columns
}

export const getUserViolatingJobTableColumns = () => {
  const columns: AccessorKeyColumnDef<ViolatingJobTableItem, any>[] = [
    createViolatingJobTableColumn('hostname'),
    createViolatingJobTableColumn('id'),
    createViolatingJobTableColumn('policyName'),
    createViolatingJobTableColumn('started-on-or-before'),
    createViolatingJobTableColumn('last-seen'),
    createViolatingJobTableColumn('rcpu-avg'),
    createViolatingJobTableColumn('rcpu-peak'),
    createViolatingJobTableColumn('rmem-avg'),
    createViolatingJobTableColumn('cmd'),
  ]
  return columns
}

const violatingJobTableColumnHelper = createColumnHelper<ViolatingJobTableItem>()

function createViolatingJobTableColumn<K extends keyof ViolatingJobTableItem>(key: K) {
  // Ensure that column definition exists in the constants
  const columnDef = VIOLATING_JOB_SUMMARY_COLUMN[key]
  if (!columnDef) {
    throw new Error(`Column definition for key '${key}' not found.`)
  }

  // Accessor and Header are always used, but other properties like cell and meta are added as needed
  return violatingJobTableColumnHelper.accessor(key, {
    cell: props => {
      if (columnDef.renderFn) {
        return columnDef.renderFn({value: props.getValue()})
      }
      return props.getValue()
    },
    header: columnDef.title,
    meta: columnDef,
  })
}

export const getDeadWeightTableColumns = () => {
  const columns: AccessorKeyColumnDef<DeadWeightTableItem, any>[] = [
    createDeadWeightTableColumn('hostname'),
    createDeadWeightTableColumn('user'),
    createDeadWeightTableColumn('id'),
    createDeadWeightTableColumn('cmd'),
    createDeadWeightTableColumn('started-on-or-before'),
    createDeadWeightTableColumn('last-seen'),
  ]
  return columns
}

const deadWeightTableColumnHelper = createColumnHelper<DeadWeightTableItem>()

function createDeadWeightTableColumn<K extends keyof DeadWeightTableItem>(key: K) {
  // Ensure that column definition exists in the constants
  const columnDef = DEAD_WEIGHT_COLUMN[key]
  if (!columnDef) {
    throw new Error(`Column definition for key '${key}' not found.`)
  }

  // Accessor and Header are always used, but other properties like cell and meta are added as needed
  return deadWeightTableColumnHelper.accessor(key, {
    cell: props => {
      if (columnDef.renderFn) {
        return columnDef.renderFn({value: props.getValue()})
      }
      return props.getValue()
    },
    header: columnDef.title,
    meta: columnDef,
    minSize: columnDef.minSize,
  })
}

export const getJobQueryResultsTableColumns = () => {
  const columns: AccessorKeyColumnDef<JobQueryResultsTableItem, any>[] = [
    createJobQueryResultsTableColumn('job'),
    createJobQueryResultsTableColumn('user'),
    createJobQueryResultsTableColumn('host'),
    createJobQueryResultsTableColumn('duration'),
    createJobQueryResultsTableColumn('start'),
    createJobQueryResultsTableColumn('end'),
    createJobQueryResultsTableColumn('cpu-peak'),
    createJobQueryResultsTableColumn('res-peak'),
    createJobQueryResultsTableColumn('mem-peak'),
    createJobQueryResultsTableColumn('gpu-peak'),
    createJobQueryResultsTableColumn('gpumem-peak'),
    createJobQueryResultsTableColumn('cmd'),
  ]
  return columns
}

const jobQueryResultsTableColumnHelper = createColumnHelper<JobQueryResultsTableItem>()

function createJobQueryResultsTableColumn<K extends keyof JobQueryResultsTableItem>(key: K) {
  // Ensure that column definition exists in the constants
  const columnDef = JOB_QUERY_RESULTS_COLUMN[key]
  if (!columnDef) {
    throw new Error(`Column definition for key '${key}' not found.`)
  }

  // Accessor and Header are always used, but other properties like cell and meta are added as needed
  return jobQueryResultsTableColumnHelper.accessor(key, {
    cell: props => {
      let value = props.getValue()
      if (columnDef.formatterFns) {
        value = columnDef.formatterFns.reduce((acc, fn) => fn(acc), value)
      }
      if (columnDef.renderFn) {
        return columnDef.renderFn({value})
      }
      return value
    },
    header: columnDef.title,
    meta: columnDef,
    minSize: columnDef.minSize,
    ...(columnDef.sortingFn && { // Only add sortingFn if it exists (not undefined)
      sortingFn: columnDef.sortingFn,
    })
  })
}

export const sortByDuration: SortingFn<JobQueryResultsTableItem> = (rowA, rowB, _columnId) => {
  const durationA = rowA.original.duration
  const durationB = rowB.original.duration
  const m1 = DURATION_REGEX.exec(durationA)
  const m2 = DURATION_REGEX.exec(durationB)

  if (m1 && m2) {
    const x = parseInt(m1[1]) * 24 * 60 + parseInt(m1[2]) * 60 + parseInt(m1[3])
    const y = parseInt(m2[1]) * 24 * 60 + parseInt(m2[2]) * 60 + parseInt(m2[3])
    if (x < y) {
      return -1
    }
    if (x > y) {
      return 1
    }
  }
  return 0
}

// failing cpus are sorted higher
// failing gpus are sorted higher
// then non-failing systems
// within each group, by hostname
export const sortDashboardTableRows = (a: DashboardTableItem, b: DashboardTableItem, uptime: boolean) => {
  if (uptime) {
    if (a.cpu_status != b.cpu_status) {
      return b.cpu_status - a.cpu_status
    }
    if (a.gpu_status != b.gpu_status) {
      return b.gpu_status - a.gpu_status
    }
  }
  if (a.hostname < b.hostname) {
    return -1
  }
  if (a.hostname > b.hostname) {
    return 1
  }
  return 0
}
