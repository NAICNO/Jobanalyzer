import { GenericCell, GpuFieldCell, HostNameFieldCell, WorkingFieldCell } from './components/table/cell'
import CellWithLink from './components/table/cell/CellWithLink.tsx'

export const APP_NAME = 'NAIC Jobanalyzer'
export const API_ENDPOINT = 'http://localhost:5173/api'

export const QueryKeys = {
  DASHBOARD_TABLE: 'DASHBOARD_TABLE',
  VIOLATIONS: 'VIOLATIONS',
  VIOLATOR: 'VIOLATOR',
  DEAD_WEIGHT: 'DEAD_WEIGHT',
  HOSTNAME_LIST: 'HOSTNAME_LIST',
  HOSTNAME: 'HOSTNAME',
}

export const FETCH_FREQUENCIES = [
  {text: 'Moment-to-moment (last 24h)', value: 'minutely'},
  {text: 'Daily, by hour', value: 'daily'},
  {text: 'Weekly, by hour', value: 'weekly'},
  {text: 'Monthly, by day', value: 'monthly'},
  {text: 'Quarterly, by day', value: 'quarterly'},
]

export const CELL_BACKGROUND_COLORS = {
  NA: 'transparent',
  DOWN: 'tomato', // #ff6347
  WORKING_HARD: 'deepskyblue', // #00bfff
  WORKING: 'lightskyblue', // #87cefa
  COASTING: 'lightcyan', // #e0ffff
}

export const CLUSTER_INFO: Record<string, Cluster> = {
  'ml': {
    cluster: 'ml',
    subclusters: [{name: 'nvidia', nodes: 'ml[1-3,6-9]'}],
    uptime: true,
    violators: true,
    deadweight: true,
    defaultQuery: '*',
    hasDowntime: true,
    name: 'ML nodes',
    description: 'UiO Machine Learning nodes',
    prefix: 'ml-',
    policy: 'Significant CPU usage without any GPU usage',
  },
  'fox': {
    cluster: 'fox',
    subclusters: [
      {name: 'cpu', nodes: 'c*'},
      {name: 'gpu', nodes: 'gpu*'},
      {name: 'int', nodes: 'int*'},
      {name: 'login', nodes: 'login*'},
    ],
    uptime: true,
    violators: false,
    deadweight: true,
    defaultQuery: 'login* or int*',
    name: 'Fox',
    hasDowntime: true,
    description: 'UiO \'Fox\' supercomputer',
    prefix: 'fox-',
    policy: '(To be determined)',
  },
  'saga': {
    cluster: 'saga',
    subclusters: [{name: 'login', nodes: 'login*'}],
    uptime: false,
    violators: false,
    deadweight: false,
    defaultQuery: 'login*',
    name: 'Saga',
    hasDowntime: false,
    description: 'Sigma2 \'Saga\' supercomputer',
    prefix: 'saga-',
    policy: '(To be determined)',
  }
}


export const HOSTNAMES_ALIAS = {
  'ml1.hpc.uio.no': 'ML1',
  'ml2.hpc.uio.no': 'ML2',
  'ml3.hpc.uio.no': 'ML3',
  'ml4.hpc.uio.no': 'ML4',
  'ml5.hpc.uio.no': 'ML5',
  'ml6.hpc.uio.no': 'ML6',
  'ml7.hpc.uio.no': 'ML7',
  'ml8.hpc.uio.no': 'ML8',
  'ml9.hpc.uio.no': 'ML9',
} as HostnameAlias


//Refer from code/dashboard/dashboard.js

export const DASHBOARD_COLUMN: { [K in keyof DashboardTableItem]: DashboardTableColumnHeader } = {
  hostname: {
    key: 'hostname',
    title: 'Hostname',
    sortable: true,
    renderFn: CellWithLink,
  },
  cpu_status: {
    key: 'cpu_status',
    title: 'CPU status',
    shortTitle: 'CPU',
    helpText: '0=up, 1=down',
    sortable: true,
    renderFn: GenericCell
  },
  gpu_status: {
    key: 'gpu_status',
    title: 'GPU status',
    shortTitle: 'GPU',
    helpText: '0=up, 1=down',
    sortable: true,
    renderFn: GpuFieldCell
  },
  users_recent: {
    key: 'users_recent',
    title: 'Users (recent)',
    shortTitle: 'Recent',
    helpText: 'Unique users running jobs',
    sortable: true,
    renderFn: GenericCell
  },
  users_longer: {
    key: 'users_longer',
    title: 'Users (longer)',
    shortTitle: 'Longer',
    helpText: 'Unique users running jobs',
    sortable: true,
    renderFn: GenericCell
  },
  jobs_recent: {
    key: 'jobs_recent',
    title: 'Jobs (recent)',
    shortTitle: 'Recent',
    helpText: 'Jobs big enough to count',
    sortable: true,
    renderFn: GenericCell
  },
  jobs_longer: {
    key: 'jobs_longer',
    title: 'Jobs (longer)',
    shortTitle: 'Longer',
    helpText: 'Jobs big enough to count',
    sortable: true,
    renderFn: GenericCell
  },
  cpu_recent: {
    key: 'cpu_recent',
    title: 'CPU % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
    sortable: true,
    renderFn: WorkingFieldCell,
  },
  cpu_longer: {
    key: 'cpu_longer',
    title: 'CPU % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
    sortable: true,
    renderFn: WorkingFieldCell,
  },
  resident_recent: {
    key: 'resident_recent',
    title: 'Resident% (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  resident_longer: {
    key: 'resident_longer',
    title: 'Resident% (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  mem_recent: {
    key: 'mem_recent',
    title: 'Virt % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  mem_longer: {
    key: 'mem_longer',
    title: 'Virt % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  gpu_recent: {
    key: 'gpu_recent',
    title: 'GPU % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  gpu_longer: {
    key: 'gpu_longer',
    title: 'GPU % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  gpumem_recent: {
    key: 'gpumem_recent',
    title: 'GPU Mem % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  gpumem_longer: {
    key: 'gpumem_longer',
    title: 'GPU Mem % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
    renderFn: WorkingFieldCell,
  },
  violators_long: {
    key: 'violators_long',
    title: 'Violators (new)',
    shortTitle: 'Viol. (new)',
    helpText: 'New jobs violating policy',
    renderFn: GenericCell,
  },
  zombies_long: {
    key: 'zombies_long',
    title: 'Zombies (new)',
    shortTitle: 'Zomb. (new)',
    helpText: 'New defunct and zombie jobs',
    renderFn: GenericCell
  },
  tag: {
    key: 'tag',
    title: 'Tag',
    sortable: true,
  },
  machine: {
    key: 'machine',
    title: 'Machine',
  },
  longer: {
    key: 'longer',
    title: 'Longer',
    sortable: true,
  },
  long: {
    key: 'long',
    title: 'Long',
    sortable: true,
  },
  recent: {
    key: 'recent',
    title: 'Recent',
    sortable: true,
  }
}

export const VIOLATING_USER_SUMMARY_COLUMN: { [K in keyof ViolatingUser]: ViolatingUserTableColumnHeader } = {
  user: {
    key: 'user',
    title: 'User',
    sortable: true,
    renderFn: CellWithLink
  },
  count: {
    key: 'count',
    title: 'No. violations',
    sortable: true,
    renderFn: GenericCell
  },
  earliest: {
    key: 'earliest',
    title: 'First seen',
    sortable: true,
    renderFn: GenericCell
  },
  latest: {
    key: 'latest',
    title: 'Last seen',
    sortable: true,
    renderFn: GenericCell
  },
}

export const VIOLATING_JOB_SUMMARY_COLUMN: { [K in keyof ViolatingJob]: ViolatingJobTableColumnHeader } = {
  hostname: {
    key: 'hostname',
    title: 'Hostname',
    sortable: true,
    renderFn: HostNameFieldCell
  },
  user: {
    key: 'user',
    title: 'User',
    sortable: true,
    renderFn: GenericCell
  },
  id: {
    key: 'id',
    title: 'Job',
    sortable: true,
    renderFn: GenericCell
  },
  cmd: {
    key: 'cmd',
    title: 'Command',
    sortable: true,
    renderFn: GenericCell
  },
  'started-on-or-before': {
    key: 'started-on-or-before',
    title: 'First seen',
    sortable: true,
    renderFn: GenericCell
  },
  'last-seen': {
    key: 'last-seen',
    title: 'Last seen',
    sortable: true,
    renderFn: GenericCell
  },
  'cpu-peak': {
    key: 'cpu-peak',
    title: 'CPU peak',
    sortable: true,
    renderFn: GenericCell
  },
  'rcpu-avg': {
    key: 'rcpu-avg',
    title: 'CPU% avg',
    sortable: true,
    renderFn: GenericCell
  },
  'rcpu-peak': {
    key: 'rcpu-peak',
    title: 'CPU% peak',
    sortable: true,
    renderFn: GenericCell
  },
  'rmem-avg': {
    key: 'rmem-avg',
    title: 'Virt% avg',
    sortable: true,
    renderFn: GenericCell
  },
  'rmem-peak': {
    key: 'rmem-peak',
    title: 'Virt% peak',
    sortable: true,
    renderFn: GenericCell
  },
  'first-violation': {
    key: 'first-violation',
    title: 'First violation',
    sortable: true,
    renderFn: GenericCell
  },
  policyName: {
    key: 'policy',
    title: 'Policy',
    sortable: true,
    renderFn: GenericCell
  },
}

export const POLICIES: Policies = {
  'ml': [
    {
      name: 'ml-cpuhog',
      trigger: 'Job uses more than 10% of system\'s CPU at peak, runs for at least 10 minutes, and uses no GPU at all',
      problem: 'ML nodes are for GPU jobs.  Job is in the way of other jobs that need GPU',
      remedy: 'Move your work to a GPU-less system such as Fox or Light-HPC',
    }
  ]
}

export const DEAD_WEIGHT_COLUMN: { [K in keyof DeadWeightTableItem]: DeadWeightTableColumnHeader } = {
  hostname: {
    key: 'hostname',
    title: 'Hostname',
    sortable: true,
    renderFn: GenericCell
  },
  user: {
    key: 'user',
    title: 'User',
    sortable: true,
    renderFn: GenericCell
  },
  id: {
    key: 'id',
    title: 'Job',
    sortable: true,
    renderFn: GenericCell
  },
  cmd: {
    key: 'cmd',
    title: 'Command',
    sortable: true,
    renderFn: GenericCell
  },
  'started-on-or-before': {
    key: 'started-on-or-before',
    title: 'First seen',
    sortable: true,
    renderFn: GenericCell
  },
  'first-violation': {
    key: 'first-violation',
    title: 'First violation',
    sortable: true,
    renderFn: GenericCell
  },
  'last-seen': {
    key: 'last-seen',
    title: 'Last seen',
    sortable: true,
    renderFn: GenericCell
  },
}

