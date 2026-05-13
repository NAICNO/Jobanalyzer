/**
 * Mock detailed job data for demo mode.
 *
 * Provides SAcct data and JobReport for job 10001 (llm-finetune-7b),
 * the primary example job running on gpu001 with 4x A100 GPUs.
 */

import type { JobResponse, SAcctResponse, JobReport } from '../../client/types.gen'
import { CLUSTER_NAME } from './cluster'
import { ALL_GPU_UUIDS } from './gpu-cards'

// ---------------------------------------------------------------------------
// Time helpers
// ---------------------------------------------------------------------------

const NOW = new Date()

function hoursAgo(h: number): string {
  return new Date(NOW.getTime() - h * 3600_000).toISOString()
}

// ---------------------------------------------------------------------------
// The detailed job (full JobResponse with sacct)
// ---------------------------------------------------------------------------

const sacctData: SAcctResponse = {
  AllocTRES: 'billing=32,cpu=32,gres/gpu=4,mem=256G,node=1',
  ElapsedRaw: 21600,         // 6 hours
  SystemCPU: 2850,           // ~47 minutes system CPU
  UserCPU: 645000,           // ~179 hours user CPU (32 cores * 6h * ~93% util)
  AveVMSize: 25165824,       // ~24 GiB average virtual memory
  MaxVMSize: 33554432,       // ~32 GiB peak virtual memory
  AveCPU: 20160,             // average CPU seconds per task
  MinCPU: 18500,             // minimum CPU seconds per task
  AveRSS: 18874368,          // ~18 GiB average RSS
  MaxRSS: 22020096,          // ~21 GiB peak RSS
  AveDiskRead: 107374182400, // ~100 GiB read (dataset loading)
  AveDiskWrite: 21474836480, // ~20 GiB written (checkpoints)
}

export const mockDetailedJob: JobResponse = {
  time: NOW.toISOString() as unknown as Date,
  cluster: CLUSTER_NAME,
  job_id: 10001,
  job_step: '',
  job_name: 'llm-finetune-7b',
  job_state: 'RUNNING',
  array_job_id: null,
  array_task_id: null,
  het_job_id: 0,
  het_job_offset: 0,
  user_name: 'user_alpha',
  account: 'ml_group',
  start_time: hoursAgo(6) as unknown as Date,
  suspend_time: 0,
  submit_time: hoursAgo(7) as unknown as Date,
  time_limit: 172800,         // 48 hours
  end_time: null,
  exit_code: null,
  partition: 'gpu',
  reservation: '',
  nodes: ['gpu001'],
  priority: 4294901735,
  distribution: 'block:block',
  gres_detail: null,
  requested_resources: 'gpu:4',
  allocated_resources: 'gpu:4',
  requested_cpus: 32,
  requested_memory_per_node: 268435456, // 256 GiB in KiB
  requested_node_count: 1,
  minimum_cpus_per_node: null,
  used_gpu_uuids: ALL_GPU_UUIDS['gpu001'],
  sacct: sacctData,
}

// ---------------------------------------------------------------------------
// Job report for job 10001
// ---------------------------------------------------------------------------

export const mockJobReport: JobReport = {
  cpu_avg: { mean: 87.5, stddev: 4.2 },
  cpu_util: { mean: 92.3, stddev: 6.1 },
  resident_memory: { mean: 18874368, stddev: 2097152 },    // ~18 GiB +/- 2 GiB
  virtual_memory: { mean: 25165824, stddev: 3145728 },     // ~24 GiB +/- 3 GiB
  num_threads: { mean: 42, stddev: 3 },
  data_read: { mean: 107374182400, stddev: 5368709120 },   // ~100 GiB +/- 5 GiB
  data_written: { mean: 21474836480, stddev: 2147483648 }, // ~20 GiB +/- 2 GiB
  data_cancelled: { mean: 0, stddev: 0 },
  requested_cpus: 32,
  requested_memory_per_node: 268435456, // 256 GiB in KiB
  requested_gpus: 4,
  used_gpu_uuids: ALL_GPU_UUIDS['gpu001'],
  nodes: ['gpu001'],
  warnings: [
    'GPU memory utilization below 60% on GPU-gpu001-03-0000-0000-000000000000 (avg 54.2%)',
  ],
}

// ---------------------------------------------------------------------------
// Additional completed job for variety
// ---------------------------------------------------------------------------

const completedSacctData: SAcctResponse = {
  AllocTRES: 'billing=64,cpu=64,gres/gpu=0,mem=512G,node=2',
  ElapsedRaw: 43200,         // 12 hours
  SystemCPU: 5400,
  UserCPU: 2764800,          // 64 cores * 12h * ~100%
  AveVMSize: 419430400,      // ~400 GiB
  MaxVMSize: 524288000,      // ~500 GiB
  AveCPU: 43200,
  MinCPU: 42000,
  AveRSS: 314572800,         // ~300 GiB
  MaxRSS: 419430400,         // ~400 GiB
  AveDiskRead: 536870912000, // ~500 GiB
  AveDiskWrite: 107374182400,
}

const completedJob: JobResponse = {
  time: NOW.toISOString() as unknown as Date,
  cluster: CLUSTER_NAME,
  job_id: 10050,
  job_step: '',
  job_name: 'weather-sim-global',
  job_state: 'COMPLETED',
  array_job_id: null,
  array_task_id: null,
  het_job_id: 0,
  het_job_offset: 0,
  user_name: 'user_zeta',
  account: 'climate_lab',
  start_time: hoursAgo(18) as unknown as Date,
  suspend_time: 0,
  submit_time: hoursAgo(20) as unknown as Date,
  time_limit: 172800,
  end_time: hoursAgo(6) as unknown as Date,
  exit_code: 0,
  partition: 'cpu',
  reservation: '',
  nodes: ['cn001', 'cn002'],
  priority: 4294901500,
  distribution: 'cyclic',
  gres_detail: null,
  requested_resources: null,
  allocated_resources: null,
  requested_cpus: 64,
  requested_memory_per_node: 536870912,
  requested_node_count: 2,
  minimum_cpus_per_node: null,
  used_gpu_uuids: null,
  sacct: completedSacctData,
}

const completedJobReport: JobReport = {
  cpu_avg: { mean: 96.5, stddev: 2.1 },
  cpu_util: { mean: 98.0, stddev: 1.5 },
  resident_memory: { mean: 314572800, stddev: 52428800 },
  virtual_memory: { mean: 419430400, stddev: 52428800 },
  num_threads: { mean: 68, stddev: 4 },
  data_read: { mean: 536870912000, stddev: 26843545600 },
  data_written: { mean: 107374182400, stddev: 10737418240 },
  data_cancelled: { mean: 0, stddev: 0 },
  requested_cpus: 64,
  requested_memory_per_node: 536870912,
  requested_gpus: 0,
  used_gpu_uuids: [],
  nodes: ['cn001', 'cn002'],
  warnings: [],
}

// ---------------------------------------------------------------------------
// Lookup maps
// ---------------------------------------------------------------------------

const detailedJobs: Record<number, JobResponse> = {
  10001: mockDetailedJob,
  10050: completedJob,
}

const jobReports: Record<number, JobReport> = {
  10001: mockJobReport,
  10050: completedJobReport,
}

// ---------------------------------------------------------------------------
// Exports
// ---------------------------------------------------------------------------

/**
 * Look up a detailed job by ID (includes sacct data).
 * Returns undefined for unknown job IDs.
 */
export function getDetailedJob(jobId: number): JobResponse | undefined {
  return detailedJobs[jobId]
}

/**
 * Look up a job report by job ID.
 * Returns undefined for unknown job IDs.
 */
export function getJobReport(jobId: number): JobReport | undefined {
  return jobReports[jobId]
}
