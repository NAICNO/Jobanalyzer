/**
 * Mock process tree data for demo mode.
 *
 * Simulates a realistic ML training process tree for job 10001
 * running on gpu001 as user_alpha.
 */

import type {
  SystemProcessTreeResponse,
  ProcessTreeResponse,
  ProcessData,
  ProcessRelations,
  ProcessTreeMetaData,
  SampleProcessAccResponse,
} from '../../client/types.gen'

// ---------------------------------------------------------------------------
// Time helpers
// ---------------------------------------------------------------------------

const NOW = new Date()

function unixHoursAgo(h: number): number {
  return Math.floor((NOW.getTime() - h * 3600_000) / 1000)
}

// ---------------------------------------------------------------------------
// Sample data point factories
// ---------------------------------------------------------------------------

/** Generate N evenly-spaced sample data points over the last `spanMinutes` minutes. */
function makeSamples(
  count: number,
  spanMinutes: number,
  opts: {
    cpuAvg: number
    cpuUtil: number
    cpuTimeBase: number
    memResident: number
    memVirtual: number
    memUtil: number
    processesAvg: number
  },
): SampleProcessAccResponse[] {
  const samples: SampleProcessAccResponse[] = []
  const intervalMs = (spanMinutes * 60_000) / (count - 1 || 1)

  for (let i = 0; i < count; i++) {
    const jitter = (Math.sin(i * 1.7) * 0.05 + 1) // deterministic pseudo-jitter
    samples.push({
      time: new Date(NOW.getTime() - (spanMinutes * 60_000) + i * intervalMs).toISOString() as unknown as Date,
      memory_resident: Math.round(opts.memResident * jitter),
      memory_virtual: Math.round(opts.memVirtual * jitter),
      memory_util: Math.round(opts.memUtil * jitter * 10) / 10,
      cpu_avg: Math.round(opts.cpuAvg * jitter * 10) / 10,
      cpu_util: Math.round(opts.cpuUtil * jitter * 10) / 10,
      cpu_time: Math.round(opts.cpuTimeBase + i * 120 * jitter),
      processes_avg: opts.processesAvg,
    })
  }
  return samples
}

// ---------------------------------------------------------------------------
// Process definitions for job 10001 on gpu001
// ---------------------------------------------------------------------------

const USER = 'user_alpha'
const SPAN = 120 // last 2 hours in minutes

const processes: Record<string, ProcessData> = {
  '1000': {
    ppid: 0,
    user: 'root',
    cmd: 'slurmstepd',
    data: makeSamples(4, SPAN, {
      cpuAvg: 0.5,
      cpuUtil: 0.3,
      cpuTimeBase: 10,
      memResident: 8192,
      memVirtual: 32768,
      memUtil: 0.1,
      processesAvg: 1,
    }),
  },
  '1001': {
    ppid: 1000,
    user: USER,
    cmd: 'bash',
    data: makeSamples(4, SPAN, {
      cpuAvg: 0.2,
      cpuUtil: 0.1,
      cpuTimeBase: 5,
      memResident: 4096,
      memVirtual: 16384,
      memUtil: 0.05,
      processesAvg: 1,
    }),
  },
  '1002': {
    ppid: 1001,
    user: USER,
    cmd: 'python',
    data: makeSamples(5, SPAN, {
      cpuAvg: 85.0,
      cpuUtil: 92.0,
      cpuTimeBase: 18000,
      memResident: 12582912,   // ~12 GiB
      memVirtual: 25165824,    // ~24 GiB
      memUtil: 15.4,
      processesAvg: 1,
    }),
  },
  '1003': {
    ppid: 1002,
    user: USER,
    cmd: 'python: data_loader_0',
    data: makeSamples(5, SPAN, {
      cpuAvg: 45.0,
      cpuUtil: 55.0,
      cpuTimeBase: 6000,
      memResident: 2097152,    // ~2 GiB
      memVirtual: 4194304,
      memUtil: 2.6,
      processesAvg: 1,
    }),
  },
  '1004': {
    ppid: 1002,
    user: USER,
    cmd: 'python: data_loader_1',
    data: makeSamples(5, SPAN, {
      cpuAvg: 42.0,
      cpuUtil: 50.0,
      cpuTimeBase: 5800,
      memResident: 2097152,
      memVirtual: 4194304,
      memUtil: 2.6,
      processesAvg: 1,
    }),
  },
  '1005': {
    ppid: 1002,
    user: USER,
    cmd: 'python: training_worker_0',
    data: makeSamples(5, SPAN, {
      cpuAvg: 95.0,
      cpuUtil: 98.0,
      cpuTimeBase: 20000,
      memResident: 8388608,    // ~8 GiB
      memVirtual: 16777216,
      memUtil: 10.3,
      processesAvg: 1,
    }),
  },
  '1006': {
    ppid: 1002,
    user: USER,
    cmd: 'python: training_worker_1',
    data: makeSamples(5, SPAN, {
      cpuAvg: 94.0,
      cpuUtil: 97.5,
      cpuTimeBase: 19800,
      memResident: 8388608,
      memVirtual: 16777216,
      memUtil: 10.3,
      processesAvg: 1,
    }),
  },
  '1007': {
    ppid: 1002,
    user: USER,
    cmd: 'python: training_worker_2',
    data: makeSamples(4, SPAN, {
      cpuAvg: 93.0,
      cpuUtil: 96.0,
      cpuTimeBase: 19500,
      memResident: 8388608,
      memVirtual: 16777216,
      memUtil: 10.3,
      processesAvg: 1,
    }),
  },
  '1008': {
    ppid: 1002,
    user: USER,
    cmd: 'python: training_worker_3',
    data: makeSamples(4, SPAN, {
      cpuAvg: 92.0,
      cpuUtil: 95.0,
      cpuTimeBase: 19200,
      memResident: 8388608,
      memVirtual: 16777216,
      memUtil: 10.3,
      processesAvg: 1,
    }),
  },
  '1009': {
    ppid: 1002,
    user: USER,
    cmd: 'python: eval_worker',
    data: makeSamples(3, SPAN, {
      cpuAvg: 30.0,
      cpuUtil: 25.0,
      cpuTimeBase: 3000,
      memResident: 4194304,    // ~4 GiB
      memVirtual: 8388608,
      memUtil: 5.2,
      processesAvg: 1,
    }),
  },
  '1010': {
    ppid: 1002,
    user: USER,
    cmd: 'python: checkpoint_saver',
    data: makeSamples(3, SPAN, {
      cpuAvg: 5.0,
      cpuUtil: 3.0,
      cpuTimeBase: 600,
      memResident: 1048576,    // ~1 GiB
      memVirtual: 2097152,
      memUtil: 1.3,
      processesAvg: 1,
    }),
  },
}

// ---------------------------------------------------------------------------
// Relations (parent -> child edges)
// ---------------------------------------------------------------------------

const relations: ProcessRelations[] = [
  { relation_id: 'r-1000-1001', source: 1000, target: 1001 },
  { relation_id: 'r-1001-1002', source: 1001, target: 1002 },
  { relation_id: 'r-1002-1003', source: 1002, target: 1003 },
  { relation_id: 'r-1002-1004', source: 1002, target: 1004 },
  { relation_id: 'r-1002-1005', source: 1002, target: 1005 },
  { relation_id: 'r-1002-1006', source: 1002, target: 1006 },
  { relation_id: 'r-1002-1007', source: 1002, target: 1007 },
  { relation_id: 'r-1002-1008', source: 1002, target: 1008 },
  { relation_id: 'r-1002-1009', source: 1002, target: 1009 },
  { relation_id: 'r-1002-1010', source: 1002, target: 1010 },
]

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

const metadata: ProcessTreeMetaData = {
  total_processes: Object.keys(processes).length,
  start_time: unixHoursAgo(6),
  end_time: Math.floor(NOW.getTime() / 1000),
  root_pid: 1000,
  max_depth: 3,
}

// ---------------------------------------------------------------------------
// Assemble the tree for gpu001
// ---------------------------------------------------------------------------

const gpu001Tree: ProcessTreeResponse = {
  processes,
  relations,
  metadata,
}

// ---------------------------------------------------------------------------
// Exports
// ---------------------------------------------------------------------------

export const mockProcessTree: SystemProcessTreeResponse = {
  job: 10001,
  epoch: 0,
  nodes: {
    gpu001: gpu001Tree,
  },
}

/**
 * Look up a process tree by job ID.
 * Only job 10001 has data in demo mode; others return undefined.
 */
export function getProcessTree(jobId: number): SystemProcessTreeResponse | undefined {
  if (jobId === 10001) return mockProcessTree
  return undefined
}
