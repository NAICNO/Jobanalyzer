export interface BenchmarkRecord {
  system: string
  node: string
  number_of_gpus: number
  task_name: string
  start_time: number
  end_time: number
  exit_code: string | number
  slurm_job_id: number
  precision: string
  metric_name: string
  metric_value: number
}

export interface BenchmarkFilterState {
  metric: string
  selectedPrecisions: string[]
  selectedGpuCounts: number[]
  selectedTests: string[]
  referenceSystem: string
  comparisonType: 'relative' | 'absolute'
  systemFilter: string
}
