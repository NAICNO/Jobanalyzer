export const APP_NAME = 'NAIC Jobanalyzer'
export const API_ENDPOINT = 'http://localhost:5173/api'

export const QueryKeys = {
  DASHBOARD_TABLE: 'DASHBOARD_TABLE',
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
  },
  cpu_status: {
    key: 'cpu_status',
    title: 'CPU status',
    shortTitle: 'CPU',
    helpText: '0=up, 1=down',
    sortable: true,
  },
  gpu_status: {
    key: 'gpu_status',
    title: 'GPU status',
    shortTitle: 'GPU',
    helpText: '0=up, 1=down',
    sortable: true,
  },
  users_recent: {
    key: 'users_recent',
    title: 'Users (recent)',
    shortTitle: 'Recent',
    helpText: 'Unique users running jobs',
    sortable: true,
  },
  users_longer: {
    key: 'users_longer',
    title: 'Users (longer)',
    shortTitle: 'Longer',
    helpText: 'Unique users running jobs',
    sortable: true,
  },
  jobs_recent: {
    key: 'jobs_recent',
    title: 'Jobs (recent)',
    shortTitle: 'Recent',
    helpText: 'Jobs big enough to count',
    sortable: true,
  },
  jobs_longer: {
    key: 'jobs_longer',
    title: 'Jobs (longer)',
    shortTitle: 'Longer',
    helpText: 'Jobs big enough to count',
    sortable: true,
  },
  cpu_recent: {
    key: 'cpu_recent',
    title: 'CPU % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
    sortable: true,
  },
  cpu_longer: {
    key: 'cpu_longer',
    title: 'CPU % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
    sortable: true,
  },
  resident_recent: {
    key: 'resident_recent',
    title: 'Resident% (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
  },
  resident_longer: {
    key: 'resident_longer',
    title: 'Resident% (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
  },
  mem_recent: {
    key: 'mem_recent',
    title: 'Virt % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
  },
  mem_longer: {
    key: 'mem_longer',
    title: 'Virt % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
  },
  gpu_recent: {
    key: 'gpu_recent',
    title: 'GPU % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
  },
  gpu_longer: {
    key: 'gpu_longer',
    title: 'GPU % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
  },
  gpumem_recent: {
    key: 'gpumem_recent',
    title: 'GPU Mem % (recent)',
    shortTitle: 'Recent',
    helpText: 'Running average',
  },
  gpumem_longer: {
    key: 'gpumem_longer',
    title: 'GPU Mem % (longer)',
    shortTitle: 'Longer',
    helpText: 'Running average',
  },
  violators_long: {
    key: 'violators_long',
    title: 'Violators (new)',
    shortTitle: 'Viol. (new)',
    helpText: 'New jobs violating policy',
  },
  zombies_long: {
    key: 'zombies_long',
    title: 'Zombies (new)',
    shortTitle: 'Zomb. (new)',
    helpText: 'New defunct and zombie jobs',
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
