/**
 * Mock partition data for demo mode.
 *
 * Each partition includes inline job objects to avoid circular dependencies
 * with jobs.ts. The job shapes conform to JobResponse from the generated types.
 */

import type { PartitionResponse, JobResponse } from '../../client/types.gen'
import { CLUSTER_NAME } from './cluster'
import { ALL_GPU_UUIDS } from './gpu-cards'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const NOW = new Date()

function hoursAgo(h: number): string {
  return new Date(NOW.getTime() - h * 3600_000).toISOString()
}

function minutesAgo(m: number): string {
  return new Date(NOW.getTime() - m * 60_000).toISOString()
}

/** Minimal running-job factory for partition queue data. */
function makeRunningJob(overrides: Partial<JobResponse> & { job_id: number; job_name: string; user_name: string; partition: string }): JobResponse {
  return {
    time: NOW.toISOString() as unknown as Date,
    cluster: CLUSTER_NAME,
    job_id: overrides.job_id,
    job_step: '',
    job_name: overrides.job_name,
    job_state: 'RUNNING',
    array_job_id: null,
    array_task_id: null,
    het_job_id: 0,
    het_job_offset: 0,
    user_name: overrides.user_name,
    account: overrides.account ?? 'research',
    start_time: (overrides.start_time as unknown as string ?? hoursAgo(4)) as unknown as Date,
    suspend_time: 0,
    submit_time: (overrides.submit_time as unknown as string ?? hoursAgo(5)) as unknown as Date,
    time_limit: overrides.time_limit ?? 86400,
    end_time: null,
    exit_code: null,
    partition: overrides.partition,
    reservation: '',
    nodes: overrides.nodes ?? null,
    priority: overrides.priority ?? 1000,
    distribution: 'Unknown',
    gres_detail: null,
    requested_resources: overrides.requested_resources ?? null,
    allocated_resources: overrides.allocated_resources ?? null,
    requested_cpus: overrides.requested_cpus ?? 32,
    requested_memory_per_node: overrides.requested_memory_per_node ?? 262144000,
    requested_node_count: overrides.requested_node_count ?? 1,
    minimum_cpus_per_node: null,
    used_gpu_uuids: overrides.used_gpu_uuids ?? null,
    sacct: null,
  }
}

/** Minimal pending-job factory. */
function makePendingJob(overrides: Partial<JobResponse> & { job_id: number; job_name: string; user_name: string; partition: string }): JobResponse {
  return {
    ...makeRunningJob(overrides),
    job_state: 'PENDING',
    start_time: null,
    nodes: null,
    used_gpu_uuids: null,
  }
}

// ---------------------------------------------------------------------------
// GPU Partition - running jobs
// ---------------------------------------------------------------------------

const gpuRunningJobs: JobResponse[] = [
  makeRunningJob({
    job_id: 10001,
    job_name: 'llm-finetune-7b',
    user_name: 'user_alpha',
    partition: 'gpu',
    account: 'ml_group',
    start_time: hoursAgo(6) as unknown as Date,
    submit_time: hoursAgo(7) as unknown as Date,
    nodes: ['gpu001'],
    requested_cpus: 32,
    requested_memory_per_node: 262144000,
    requested_node_count: 1,
    requested_resources: 'gpu:4',
    allocated_resources: 'gpu:4',
    used_gpu_uuids: ALL_GPU_UUIDS['gpu001'],
    time_limit: 172800,
    priority: 4294901735,
  }),
  makeRunningJob({
    job_id: 10002,
    job_name: 'diffusion-train-v2',
    user_name: 'user_beta',
    partition: 'gpu',
    account: 'cv_lab',
    start_time: hoursAgo(3) as unknown as Date,
    submit_time: hoursAgo(4) as unknown as Date,
    nodes: ['gpu002'],
    requested_cpus: 32,
    requested_memory_per_node: 262144000,
    requested_resources: 'gpu:4',
    allocated_resources: 'gpu:4',
    used_gpu_uuids: ALL_GPU_UUIDS['gpu002'],
    time_limit: 86400,
    priority: 4294901600,
  }),
  makeRunningJob({
    job_id: 10003,
    job_name: 'protein-fold-sim',
    user_name: 'user_gamma',
    partition: 'gpu',
    account: 'bio_group',
    start_time: hoursAgo(12) as unknown as Date,
    submit_time: hoursAgo(14) as unknown as Date,
    nodes: ['gpu003', 'gpu004'],
    requested_cpus: 64,
    requested_memory_per_node: 262144000,
    requested_node_count: 2,
    requested_resources: 'gpu:8',
    allocated_resources: 'gpu:8',
    used_gpu_uuids: [...ALL_GPU_UUIDS['gpu003'], ...ALL_GPU_UUIDS['gpu004']],
    time_limit: 259200,
    priority: 4294901500,
  }),
  makeRunningJob({
    job_id: 10004,
    job_name: 'rl-agent-atari',
    user_name: 'user_delta',
    partition: 'gpu',
    account: 'rl_lab',
    start_time: hoursAgo(1) as unknown as Date,
    submit_time: hoursAgo(2) as unknown as Date,
    nodes: ['gpu005'],
    requested_cpus: 16,
    requested_memory_per_node: 131072000,
    requested_resources: 'gpu:2',
    allocated_resources: 'gpu:2',
    used_gpu_uuids: ALL_GPU_UUIDS['gpu005'].slice(0, 2),
    time_limit: 43200,
    priority: 4294901400,
  }),
  makeRunningJob({
    job_id: 10005,
    job_name: 'gan-synth-faces',
    user_name: 'user_epsilon',
    partition: 'gpu',
    account: 'cv_lab',
    start_time: hoursAgo(8) as unknown as Date,
    submit_time: hoursAgo(9) as unknown as Date,
    nodes: ['gpu006'],
    requested_cpus: 32,
    requested_memory_per_node: 262144000,
    requested_resources: 'gpu:4',
    allocated_resources: 'gpu:4',
    used_gpu_uuids: ALL_GPU_UUIDS['gpu006'],
    time_limit: 172800,
    priority: 4294901300,
  }),
  makeRunningJob({
    job_id: 10009,
    job_name: 'whisper-finetune',
    user_name: 'user_alpha',
    partition: 'gpu',
    account: 'ml_group',
    start_time: hoursAgo(2) as unknown as Date,
    submit_time: hoursAgo(3) as unknown as Date,
    nodes: ['gpu007'],
    requested_cpus: 16,
    requested_memory_per_node: 131072000,
    requested_resources: 'gpu:2',
    allocated_resources: 'gpu:2',
    used_gpu_uuids: ALL_GPU_UUIDS['gpu007'].slice(0, 2),
    time_limit: 86400,
    priority: 4294901200,
  }),
]

// ---------------------------------------------------------------------------
// GPU Partition - pending jobs
// ---------------------------------------------------------------------------

const gpuPendingJobs: JobResponse[] = [
  makePendingJob({
    job_id: 10006,
    job_name: 'vit-pretrain-large',
    user_name: 'user_beta',
    partition: 'gpu',
    account: 'cv_lab',
    submit_time: hoursAgo(1) as unknown as Date,
    requested_cpus: 64,
    requested_memory_per_node: 524288000,
    requested_node_count: 2,
    requested_resources: 'gpu:8',
    time_limit: 259200,
    priority: 4294901100,
  }),
  makePendingJob({
    job_id: 10007,
    job_name: 'bert-qa-finetune',
    user_name: 'user_gamma',
    partition: 'gpu',
    account: 'nlp_group',
    submit_time: minutesAgo(30) as unknown as Date,
    requested_cpus: 16,
    requested_memory_per_node: 131072000,
    requested_resources: 'gpu:2',
    time_limit: 43200,
    priority: 4294901000,
  }),
  makePendingJob({
    job_id: 10008,
    job_name: 'clip-eval-imagenet',
    user_name: 'user_delta',
    partition: 'gpu',
    account: 'cv_lab',
    submit_time: minutesAgo(15) as unknown as Date,
    requested_cpus: 8,
    requested_memory_per_node: 65536000,
    requested_resources: 'gpu:1',
    time_limit: 7200,
    priority: 4294900900,
  }),
]

// ---------------------------------------------------------------------------
// CPU Partition jobs
// ---------------------------------------------------------------------------

const cpuRunningJobs: JobResponse[] = [
  makeRunningJob({
    job_id: 10020,
    job_name: 'weather-sim-global',
    user_name: 'user_zeta',
    partition: 'cpu',
    account: 'climate_lab',
    start_time: hoursAgo(18) as unknown as Date,
    submit_time: hoursAgo(20) as unknown as Date,
    nodes: ['cn001', 'cn002', 'cn003', 'cn004'],
    requested_cpus: 1024,
    requested_memory_per_node: 524288000,
    requested_node_count: 4,
    time_limit: 172800,
    priority: 4294901500,
  }),
  makeRunningJob({
    job_id: 10021,
    job_name: 'genome-assembly',
    user_name: 'user_eta',
    partition: 'cpu',
    account: 'bio_group',
    start_time: hoursAgo(5) as unknown as Date,
    submit_time: hoursAgo(6) as unknown as Date,
    nodes: ['cn005', 'cn006'],
    requested_cpus: 512,
    requested_memory_per_node: 262144000,
    requested_node_count: 2,
    time_limit: 86400,
    priority: 4294901400,
  }),
  makeRunningJob({
    job_id: 10022,
    job_name: 'cfd-turbulence',
    user_name: 'user_theta',
    partition: 'cpu',
    account: 'engineering',
    start_time: hoursAgo(10) as unknown as Date,
    submit_time: hoursAgo(11) as unknown as Date,
    nodes: ['cn007', 'cn008', 'cn009', 'cn010'],
    requested_cpus: 1024,
    requested_memory_per_node: 262144000,
    requested_node_count: 4,
    time_limit: 259200,
    priority: 4294901300,
  }),
]

const cpuPendingJobs: JobResponse[] = [
  makePendingJob({
    job_id: 10023,
    job_name: 'monte-carlo-pricing',
    user_name: 'user_iota',
    partition: 'cpu',
    account: 'finance',
    submit_time: hoursAgo(2) as unknown as Date,
    requested_cpus: 256,
    requested_memory_per_node: 131072000,
    requested_node_count: 1,
    time_limit: 43200,
    priority: 4294901200,
  }),
  makePendingJob({
    job_id: 10024,
    job_name: 'molecular-dynamics',
    user_name: 'user_gamma',
    partition: 'cpu',
    account: 'bio_group',
    submit_time: minutesAgo(45) as unknown as Date,
    requested_cpus: 512,
    requested_memory_per_node: 262144000,
    requested_node_count: 2,
    time_limit: 86400,
    priority: 4294901100,
  }),
]

// ---------------------------------------------------------------------------
// Interactive Partition jobs
// ---------------------------------------------------------------------------

const interactiveRunningJobs: JobResponse[] = [
  makeRunningJob({
    job_id: 10030,
    job_name: 'jupyter-notebook',
    user_name: 'user_alpha',
    partition: 'interactive',
    account: 'ml_group',
    start_time: hoursAgo(2) as unknown as Date,
    submit_time: hoursAgo(2) as unknown as Date,
    nodes: ['cn001'],
    requested_cpus: 4,
    requested_memory_per_node: 16384000,
    requested_resources: 'gpu:1',
    allocated_resources: 'gpu:1',
    used_gpu_uuids: ALL_GPU_UUIDS['gpu001']?.slice(0, 1) ?? [],
    time_limit: 28800,
    priority: 4294901700,
  }),
]

const interactivePendingJobs: JobResponse[] = [
  makePendingJob({
    job_id: 10031,
    job_name: 'vscode-remote',
    user_name: 'user_beta',
    partition: 'interactive',
    account: 'cv_lab',
    submit_time: minutesAgo(5) as unknown as Date,
    requested_cpus: 4,
    requested_memory_per_node: 16384000,
    time_limit: 28800,
    priority: 4294901600,
  }),
]

// ---------------------------------------------------------------------------
// Highmem Partition jobs
// ---------------------------------------------------------------------------

const highmemRunningJobs: JobResponse[] = [
  makeRunningJob({
    job_id: 10040,
    job_name: 'graph-analytics',
    user_name: 'user_kappa',
    partition: 'highmem',
    account: 'data_science',
    start_time: hoursAgo(4) as unknown as Date,
    submit_time: hoursAgo(5) as unknown as Date,
    nodes: ['hi001'],
    requested_cpus: 112,
    requested_memory_per_node: 2097152000,
    time_limit: 86400,
    priority: 4294901500,
  }),
]

const highmemPendingJobs: JobResponse[] = []

// ---------------------------------------------------------------------------
// Collect all GPU UUIDs in use across the gpu partition
// ---------------------------------------------------------------------------

const gpuInUseUuids: string[] = gpuRunningJobs.flatMap(
  (j) => (j.used_gpu_uuids as string[] | null) ?? [],
)

// ---------------------------------------------------------------------------
// Partition definitions
// ---------------------------------------------------------------------------

function makeNodes(prefix: string, start: number, end: number): string[] {
  return Array.from({ length: end - start + 1 }, (_, i) =>
    `${prefix}${String(start + i).padStart(3, '0')}`,
  )
}

const gpuPartition: PartitionResponse = {
  time: NOW.toISOString() as unknown as Date,
  cluster: CLUSTER_NAME,
  name: 'gpu',
  nodes: makeNodes('gpu', 1, 8),
  nodes_compact: ['gpu[001-008]'],
  jobs_pending: gpuPendingJobs,
  jobs_running: gpuRunningJobs,
  pending_max_submit_time: hoursAgo(1) as unknown as Date,
  running_latest_wait_time: 1800,
  total_cpus: 1024,
  total_gpus: 32,
  gpus_reserved: 28,
  gpus_in_use: gpuInUseUuids,
}

const cpuPartition: PartitionResponse = {
  time: NOW.toISOString() as unknown as Date,
  cluster: CLUSTER_NAME,
  name: 'cpu',
  nodes: makeNodes('cn', 1, 20),
  nodes_compact: ['cn[001-020]'],
  jobs_pending: cpuPendingJobs,
  jobs_running: cpuRunningJobs,
  pending_max_submit_time: hoursAgo(2) as unknown as Date,
  running_latest_wait_time: 900,
  total_cpus: 5120,
  total_gpus: 0,
  gpus_reserved: 0,
  gpus_in_use: [],
}

const interactivePartition: PartitionResponse = {
  time: NOW.toISOString() as unknown as Date,
  cluster: CLUSTER_NAME,
  name: 'interactive',
  nodes: ['cn001', 'cn002', 'gpu001'],
  nodes_compact: ['cn[001-002]', 'gpu001'],
  jobs_pending: interactivePendingJobs,
  jobs_running: interactiveRunningJobs,
  pending_max_submit_time: minutesAgo(5) as unknown as Date,
  running_latest_wait_time: 30,
  total_cpus: 640,
  total_gpus: 4,
  gpus_reserved: 1,
  gpus_in_use: ALL_GPU_UUIDS['gpu001']?.slice(0, 1) ?? [],
}

const highmemPartition: PartitionResponse = {
  time: NOW.toISOString() as unknown as Date,
  cluster: CLUSTER_NAME,
  name: 'highmem',
  nodes: ['hi001', 'hi002'],
  nodes_compact: ['hi[001-002]'],
  jobs_pending: highmemPendingJobs,
  jobs_running: highmemRunningJobs,
  pending_max_submit_time: hoursAgo(5) as unknown as Date,
  running_latest_wait_time: 600,
  total_cpus: 448,
  total_gpus: 0,
  gpus_reserved: 0,
  gpus_in_use: [],
}

// ---------------------------------------------------------------------------
// Exports
// ---------------------------------------------------------------------------

export const mockPartitions: PartitionResponse[] = [
  gpuPartition,
  cpuPartition,
  interactivePartition,
  highmemPartition,
]

export function getPartition(name: string): PartitionResponse | undefined {
  return mockPartitions.find((p) => p.name === name)
}
