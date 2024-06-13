import { GrNodes, GrServers } from 'react-icons/gr'
import { GiFox } from 'react-icons/gi'
import { MdOutlineQueryStats } from 'react-icons/md'
import * as yup from 'yup'

import {
  GenericCell,
  GpuFieldCell,
  JobQueryJobIdCell,
  WorkingFieldCell
} from './components/table/cell'
import CellWithLink from './components/table/cell/CellWithLink.tsx'
import { splitMultiPattern } from './util/query/HostGlobber.ts'
import CommandListCell from './components/table/cell/CommandListCell.tsx'
import {
  breakText,
  formatToUtcDateTimeString,
  parseRelativeDate,
  toPercentage,
  validateDateFormat
} from './util'
import { sortByDuration } from './util/TableUtils.ts'
import JobQueryValues from './types/JobQueryValues.ts'

export const APP_NAME = 'NAIC Jobanalyzer'

// URLs and API Endpoints to be moved to .env files once dev and prod environments are set up
export const APP_URL = 'https://naic-monitor.uio.no'
export const API_ENDPOINT = 'http://localhost:5173/api'
export const JOB_QUERY_API_ENDPOINT = 'http://localhost:5173/rest'

// The representation of "true" is a hack, but it's determined by the server, so live with it.
export const TRUE_VAL = 'xxxxxtruexxxxx'

export const EMPTY_ARRAY: any[] = []

export const DURATION_REGEX = /^(.*)d(.*)h(.*)m$/

export const PROFILE_NAMES = [
  {key: 'cpu', text: 'CPU'},
  {key: 'mem', text: 'RAM'},
  {key: 'gpu', text: 'GPU'},
  {key: 'gpumem', text: 'GPU RAM'}
]

export const QueryKeys = {
  DASHBOARD_TABLE: 'DASHBOARD_TABLE',
  VIOLATIONS: 'VIOLATIONS',
  VIOLATOR: 'VIOLATOR',
  DEAD_WEIGHT: 'DEAD_WEIGHT',
  HOSTNAME_LIST: 'HOSTNAME_LIST',
  HOSTNAME: 'HOSTNAME',
  JOB_QUERY: 'JOB_QUERY',
}

export const SIDEBAR_ITEMS: SidebarItem[] = [
  {
    type: 'link',
    path: '/dashboard/ml',
    matches: '/ml',
    text: 'ML Nodes',
    icon: GrNodes,
  },
  {
    type: 'link',
    path: '/dashboard/fox',
    matches: '/fox',
    text: 'Fox',
    icon: GiFox
  },
  {
    type: 'link',
    path: '/dashboard/saga',
    matches: '/saga',
    text: 'Saga',
    icon: GrServers
  },
  {
    type: 'separator'
  },
  {
    type: 'link',
    path: '/jobQuery',
    matches: '/jobQuery',
    text: 'Job Query',
    icon: MdOutlineQueryStats
  }
]

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
    renderFn: CommandListCell
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
    renderFn: GenericCell,
    minSize: 200,
  },
  user: {
    key: 'user',
    title: 'User',
    sortable: true,
    renderFn: GenericCell,
    minSize: 120,
  },
  id: {
    key: 'id',
    title: 'Job',
    sortable: true,
    renderFn: GenericCell,
    minSize: 100,
  },
  cmd: {
    key: 'cmd',
    title: 'Command',
    sortable: true,
    renderFn: CommandListCell,
    minSize: 600,
  },
  'started-on-or-before': {
    key: 'started-on-or-before',
    title: 'First seen',
    sortable: true,
    renderFn: GenericCell,
    minSize: 150,
  },
  'first-violation': {
    key: 'first-violation',
    title: 'First violation',
    sortable: true,
    renderFn: GenericCell,
    minSize: 150,
  },
  'last-seen': {
    key: 'last-seen',
    title: 'Last seen',
    sortable: true,
    renderFn: GenericCell,
    minSize: 150,
  },
}

export const CHART_SERIES_CONFIGS: Record<string, ChartSeriesConfig> = {
  rcpu: {
    dataKey: 'rcpu',
    label: 'CPU %',
    lineColor: '#36A2EB',
    strokeWidth: 2
  },
  rmem: {
    dataKey: 'rmem',
    label: 'VIRT %',
    lineColor: '#FF6384',
    strokeWidth: 2
  },
  rres: {
    dataKey: 'rres',
    label: 'RAM %',
    lineColor: '#FF9F40',
    strokeWidth: 2
  },
  rgpu: {
    dataKey: 'rgpu',
    label: 'GPU %',
    lineColor: '#FFCD56',
    strokeWidth: 2
  },
  rgpumem: {
    dataKey: 'rgpumem',
    label: 'VRAM %',
    lineColor: '#4BC0C0',
    strokeWidth: 2
  },
  downhost: {
    dataKey: 'downhost',
    label: 'DOWN',
    lineColor: '#4b74c0',
    strokeWidth: 2
  },
  downgpu: {
    dataKey: 'downgpu',
    label: 'GPU_DOWN',
    lineColor: '#9966FF',
    strokeWidth: 2
  }
}

export const JOB_QUERY_RESULTS_COLUMN: { [K in keyof JobQueryResultsTableItem]: JobQueryResultsTableColumnHeader } = {
  job: {
    key: 'job',
    title: 'Job#',
    sortable: true,
    renderFn: JobQueryJobIdCell,
  },
  user: {
    key: 'user',
    title: 'User',
    sortable: true,
    renderFn: GenericCell,
  },
  host: {
    key: 'host',
    title: 'Node',
    sortable: true,
    renderFn: GenericCell,
  },
  duration: {
    key: 'duration',
    title: 'Duration',
    sortable: true,
    renderFn: GenericCell,
    sortingFn: sortByDuration,
  },
  start: {
    key: 'start',
    title: 'Start',
    sortable: true,
    formatterFns: [formatToUtcDateTimeString],
    renderFn: GenericCell,
    minSize: 150,
  },
  end: {
    key: 'end',
    title: 'End',
    sortable: true,
    formatterFns: [formatToUtcDateTimeString],
    renderFn: GenericCell,
    minSize: 150,
  },
  cpuPeak: {
    key: 'cpuPeak',
    title: 'Peak #cores',
    sortable: true,
    formatterFns: [toPercentage],
    renderFn: GenericCell,
  },
  resPeak: {
    key: 'resPeak',
    title: 'Peak resident GB',
    sortable: true,
    renderFn: GenericCell,
  },
  memPeak: {
    key: 'memPeak',
    title: 'Peak virtual GB',
    sortable: true,
    renderFn: GenericCell,
  },
  gpuPeak: {
    key: 'gpuPeak',
    title: 'Peak GPU cards',
    sortable: true,
    formatterFns: [toPercentage],
    renderFn: GenericCell,
  },
  gpumemPeak: {
    key: 'gpumemPeak',
    title: 'Peak GPU RAM GB',
    sortable: true,
    renderFn: GenericCell,
  },
  cmd: {
    key: 'cmd',
    title: 'Command',
    sortable: true,
    formatterFns: [breakText],
    renderFn: GenericCell,
  },
}

export const initialFormValues: JobQueryValues = {
  clusterName: '',
  usernames: '',
  nodeNames: '',
  jobIds: '',
  fromDate: '',
  toDate: '',
  minRuntime: '',
  minPeakCpuCores: '',
  minPeakResidentGb: '',
  gpuUsage: 'either',
}

export const JOB_QUERY_GPU_OPTIONS: SimpleRadioInputOption[] = [
  {value: 'either', text: 'Either'},
  {value: 'some-gpu', text: 'Some'},
  {value: 'no-gpu', text: 'None'},
]

export const JOB_QUERY_VALIDATION_SCHEMA = yup.object({
  clusterName: yup.string()
    .required('Cluster name is required'),
  usernames: yup.string()
    .notRequired()
    .test('is-valid-username',
      'Usernames must be comma separated and cannot contain spaces.', (value) => {
        // Allow empty input
        if (!value) return true

        // Split by comma to allow multiple usernames
        const usernames = value.split(',')

        // Validate each username
        return usernames.every(username => {
          // Check if username has no spaces and is not an empty string after trim
          return /^\S+$/.test(username.trim()) && username.trim().length > 0
        })
      }),
  nodeNames: yup.string()
    .notRequired()
    .default('')
    .test(
      'node-names-validation',
      'Invalid node name pattern',
      value => {
        try {
          splitMultiPattern(value || '')
          return true
        } catch (error) {
          const errorMessage = (error as Error)?.message || 'Invalid node name pattern'
          return new yup.ValidationError(errorMessage, null, 'nodeNames')
        }
      }
    ),
  jobIds: yup.string()
    .notRequired()
    .default('')
    .test(
      'job-ids-validation',
      'All job ids must be finite and positive integers',
      value => {
        // Allow empty field to be considered valid
        if (!value) return true

        // Split the string by commas and validate each ID
        const ids = value.split(',')
        return ids.every(id => {
          const num = parseFloat(id)
          return Number.isFinite(num) && num > 0 && Math.floor(num) === num
        })
      }
    ),
  fromDate: yup.string()
    .test('is-valid-date', 'Invalid `from` time, format is YYYY-MM-DD or Nw or Nd', value => validateDateFormat(value)),
  toDate: yup.string()
    .test('is-valid-date', 'Invalid `to` time, format is YYYY-MM-DD or Nw or Nd', value => validateDateFormat(value))
    .test('is-same-or-after-from-date', 'To date must be the same or later than from date', function (value) {
      const {fromDate} = this.parent
      if (!value || !fromDate) return true // If either date is not set, skip this test

      const fromDateMoment = parseRelativeDate(fromDate)
      const toDateMoment = parseRelativeDate(value)

      if (fromDateMoment.isValid() && toDateMoment.isValid()) {
        return toDateMoment.isSameOrAfter(fromDateMoment)
      }

      return true // Skip this test if either date is not valid
    }),
  minRuntime: yup.string()
    .notRequired()
    .default('')
    .matches(
      /^(\d+w)?(\d+d)?(\d+h)?(\d+m)?$/,
      'Invalid min-runtime format. Enter the minimum runtime as a combination of weeks (w), days (d), hours (h), and minutes (m), like \'2w3d\', \'4h\', or \'5m\'. At least one unit must be provided.'
    ),
  minPeakCpuCores: yup.number()
    .notRequired()
    .default(null)
    .integer()
    .moreThan(-1, 'CPU cores cannot be negative')
    .typeError('Enter a number for CPU cores'),
  minPeakResidentGb: yup.number()
    .notRequired()
    .default(null)
    .integer()
    .moreThan(-1, 'Resident GB cannot be negative')
    .typeError('Enter a number for Resident GB'),
  gpuUsage: yup.string()
    .oneOf(JOB_QUERY_GPU_OPTIONS.map(option => option.value), 'Invalid GPU option')
    .required('GPU usage is required')
})

