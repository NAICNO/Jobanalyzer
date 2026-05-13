/**
 * Mock benchmark data for demo mode.
 *
 * Simulates lambda-labs GPU benchmark results comparing A100 and H100
 * GPUs across different ML workloads with various precisions and GPU counts.
 */

import type { BenchmarkRecord } from '../../types/benchmark'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const NOW_UNIX = Math.floor(Date.now() / 1000)

/** Create a benchmark record with sensible defaults. */
function makeBenchmarkRecord(overrides: Partial<BenchmarkRecord> & {
  system: string
  task_name: string
  metric_value: number
}): BenchmarkRecord {
  return {
    system: overrides.system,
    node: overrides.node ?? 'gpu001',
    number_of_gpus: overrides.number_of_gpus ?? 1,
    task_name: overrides.task_name,
    start_time: overrides.start_time ?? NOW_UNIX - 3600,
    end_time: overrides.end_time ?? NOW_UNIX - 3000,
    exit_code: overrides.exit_code ?? 'SUCCESS',
    slurm_job_id: overrides.slurm_job_id ?? 90000 + Math.floor(Math.random() * 1000),
    precision: overrides.precision ?? 'fp16',
    metric_name: overrides.metric_name ?? 'throughput',
    metric_value: overrides.metric_value,
  }
}

// ---------------------------------------------------------------------------
// Benchmark test definitions
// ---------------------------------------------------------------------------

/**
 * Throughput values (samples/sec) for each test/system/precision combination.
 * Values are realistic approximations based on public benchmarks.
 */

interface TestProfile {
  task_name: string
  /** throughput for A100 fp16 1-GPU */
  a100_fp16_1: number
  /** throughput for A100 fp32 1-GPU */
  a100_fp32_1: number
  /** throughput for H100 fp16 1-GPU */
  h100_fp16_1: number
  /** throughput for H100 fp32 1-GPU */
  h100_fp32_1: number
  /** Multi-GPU scaling factor (approximate for 4 GPUs) */
  scale4x: number
}

const tests: TestProfile[] = [
  {
    task_name: 'resnet50_infer',
    a100_fp16_1: 12850,
    a100_fp32_1: 5420,
    h100_fp16_1: 19200,
    h100_fp32_1: 8100,
    scale4x: 3.8,
  },
  {
    task_name: 'bert_train_fp16',
    a100_fp16_1: 345,
    a100_fp32_1: 142,
    h100_fp16_1: 520,
    h100_fp32_1: 215,
    scale4x: 3.6,
  },
  {
    task_name: 'gpt2_finetune',
    a100_fp16_1: 1120,
    a100_fp32_1: 480,
    h100_fp16_1: 1780,
    h100_fp32_1: 730,
    scale4x: 3.5,
  },
  {
    task_name: 'stable_diffusion_infer',
    a100_fp16_1: 28.5,
    a100_fp32_1: 12.1,
    h100_fp16_1: 45.2,
    h100_fp32_1: 19.0,
    scale4x: 3.7,
  },
  {
    task_name: 'whisper_large_infer',
    a100_fp16_1: 185,
    a100_fp32_1: 82,
    h100_fp16_1: 295,
    h100_fp32_1: 128,
    scale4x: 3.4,
  },
  {
    task_name: 'llama2_7b_finetune',
    a100_fp16_1: 42,
    a100_fp32_1: 18,
    h100_fp16_1: 68,
    h100_fp32_1: 29,
    scale4x: 3.6,
  },
  {
    task_name: 'yolov8_train',
    a100_fp16_1: 620,
    a100_fp32_1: 270,
    h100_fp16_1: 940,
    h100_fp32_1: 410,
    scale4x: 3.7,
  },
  {
    task_name: 'vit_large_infer',
    a100_fp16_1: 4200,
    a100_fp32_1: 1750,
    h100_fp16_1: 6500,
    h100_fp32_1: 2800,
    scale4x: 3.8,
  },
]

// ---------------------------------------------------------------------------
// Generate records
// ---------------------------------------------------------------------------

let slurmIdCounter = 90001

function nextSlurmId(): number {
  return slurmIdCounter++
}

const lambdalRecords: BenchmarkRecord[] = []

for (const test of tests) {
  const configs: Array<{
    system: string
    node: string
    precision: string
    gpus: number
    value: number
  }> = [
    // A100 single-GPU
    { system: 'NVIDIA A100-SXM4-80GB', node: 'gpu001', precision: 'fp16', gpus: 1, value: test.a100_fp16_1 },
    { system: 'NVIDIA A100-SXM4-80GB', node: 'gpu001', precision: 'fp32', gpus: 1, value: test.a100_fp32_1 },
    // A100 4-GPU
    { system: 'NVIDIA A100-SXM4-80GB', node: 'gpu002', precision: 'fp16', gpus: 4, value: Math.round(test.a100_fp16_1 * test.scale4x * 10) / 10 },
    { system: 'NVIDIA A100-SXM4-80GB', node: 'gpu002', precision: 'fp32', gpus: 4, value: Math.round(test.a100_fp32_1 * test.scale4x * 10) / 10 },
    // H100 single-GPU
    { system: 'NVIDIA H100-SXM5-80GB', node: 'gpu005', precision: 'fp16', gpus: 1, value: test.h100_fp16_1 },
    { system: 'NVIDIA H100-SXM5-80GB', node: 'gpu005', precision: 'fp32', gpus: 1, value: test.h100_fp32_1 },
    // H100 4-GPU
    { system: 'NVIDIA H100-SXM5-80GB', node: 'gpu006', precision: 'fp16', gpus: 4, value: Math.round(test.h100_fp16_1 * test.scale4x * 10) / 10 },
    { system: 'NVIDIA H100-SXM5-80GB', node: 'gpu006', precision: 'fp32', gpus: 4, value: Math.round(test.h100_fp32_1 * test.scale4x * 10) / 10 },
  ]

  for (const cfg of configs) {
    lambdalRecords.push(
      makeBenchmarkRecord({
        system: cfg.system,
        node: cfg.node,
        task_name: test.task_name,
        number_of_gpus: cfg.gpus,
        precision: cfg.precision,
        metric_value: cfg.value,
        metric_name: 'throughput',
        slurm_job_id: nextSlurmId(),
      }),
    )
  }
}

// Add one FAILED record for realism
lambdalRecords.push(
  makeBenchmarkRecord({
    system: 'NVIDIA A100-SXM4-80GB',
    node: 'gpu003',
    task_name: 'llama2_7b_finetune',
    number_of_gpus: 4,
    precision: 'fp16',
    metric_value: 0,
    metric_name: 'throughput',
    exit_code: 'FAILED',
    slurm_job_id: nextSlurmId(),
  }),
)

// ---------------------------------------------------------------------------
// Exports
// ---------------------------------------------------------------------------

/**
 * Benchmark records keyed by benchmark name.
 * The API endpoint is `/cluster/{cluster}/benchmarks/{benchmark_name}`.
 */
export const mockBenchmarks: Record<string, BenchmarkRecord[]> = {
  lambdal: lambdalRecords,
}

/**
 * Get benchmark records by name.
 * Returns an empty array for unknown benchmark names.
 */
export function getBenchmarkRecords(benchmarkName: string): BenchmarkRecord[] {
  return mockBenchmarks[benchmarkName] ?? []
}
