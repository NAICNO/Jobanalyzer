import type { JobReport, JobResponse } from '../client/types.gen'

/**
 * Statistics data structure with mean and standard deviation
 */
export type Stats = {
  mean: number
  stddev: number
}

/**
 * Calculate CPU efficiency percentage
 * 
 * CPU Efficiency = (Actual CPU Time Used / (Allocated CPUs × Elapsed Time)) × 100
 * 
 * @param job - Job response data
 * @param elapsed - Elapsed time in seconds
 * @returns CPU efficiency percentage (0-100), or null if not calculable
 * 
 * @TODO Test with real API responses from different job types (single-node, multi-node, GPU jobs)
 * @TODO Write test cases for edge cases: zero elapsed time, missing sacct data, etc
 */
export const calculateCpuEfficiency = (job: JobResponse, elapsed: number): number | null => {
  if (!job.sacct || !job.requested_cpus || elapsed <= 0) return null
  
  const actualCpuTime = (job.sacct.SystemCPU ?? 0) + (job.sacct.UserCPU ?? 0)
  const allocatedTime = job.requested_cpus * elapsed
  
  if (allocatedTime === 0) return null
  
  return Math.min((actualCpuTime / allocatedTime) * 100, 100)
}

/**
 * Calculate Memory efficiency percentage
 * 
 * Memory Efficiency = (Peak Memory Used / Allocated Memory) × 100
 * 
 * @param job - Job response data
 * @returns Memory efficiency percentage (0-100), or null if not calculable
 * 
 * @TODO Test with real API responses including multi-node jobs with different memory configs
 * @TODO Write test cases for missing MaxRSS, zero memory allocation, over-subscription scenarios
 */
export const calculateMemoryEfficiency = (job: JobResponse): number | null => {
  if (!job.sacct?.MaxRSS || !job.requested_memory_per_node || !job.requested_node_count) {
    return null
  }
  
  const peakMemory = job.sacct.MaxRSS
  const allocatedMemory = job.requested_memory_per_node * job.requested_node_count
  
  if (allocatedMemory === 0) return null
  
  return Math.min((peakMemory / allocatedMemory) * 100, 100)
}

/**
 * Calculate GPU efficiency percentage
 * 
 * GPU Efficiency = Average GPU Utilization across all GPUs
 * 
 * @param report - Job report data
 * @returns GPU efficiency percentage (0-100), or null if not calculable
 */
export const calculateGpuEfficiency = (report?: JobReport): number | null => {
  if (!report?.used_gpu_uuids || report.used_gpu_uuids.length === 0) {
    return null
  }
  
  // Note: GPU utilization stats would need to come from the report
  // For now, this is a placeholder - we'd need GPU timeseries data
  // to calculate average GPU utilization
  return null
}

/**
 * Calculate Time efficiency percentage
 * 
 * Time Efficiency = (Elapsed Time / Time Limit) × 100
 * Capped at 100% - measures if job finished within reasonable time
 * 
 * @param elapsed - Elapsed time in seconds
 * @param timeLimit - Time limit in seconds
 * @returns Time efficiency percentage (0-100), or null if not calculable
 */
export const calculateTimeEfficiency = (elapsed: number, timeLimit: number): number | null => {
  if (timeLimit <= 0 || elapsed <= 0) return null
  
  return Math.min((elapsed / timeLimit) * 100, 100)
}

/**
 * Efficiency metrics for a job
 */
export type EfficiencyMetrics = {
  cpu: number | null
  memory: number | null
  gpu: number | null
  time: number | null
  overall: number | null
}

/**
 * Calculate overall efficiency score
 * 
 * Overall Score = weighted_average([
 *   CPU Efficiency × 0.4,
 *   Memory Efficiency × 0.3,
 *   GPU Efficiency × 0.2,  // if GPUs used, else redistribute weight
 *   Time Efficiency × 0.1
 * ])
 * 
 * @param metrics - Individual efficiency metrics
 * @returns Overall efficiency score (0-100), or null if not calculable
 */
export const calculateOverallEfficiency = (metrics: EfficiencyMetrics): number | null => {
  const weights = {
    cpu: 0.4,
    memory: 0.3,
    gpu: 0.2,
    time: 0.1,
  }
  
  let totalWeight = 0
  let weightedSum = 0
  
  if (metrics.cpu !== null) {
    weightedSum += metrics.cpu * weights.cpu
    totalWeight += weights.cpu
  }
  
  if (metrics.memory !== null) {
    weightedSum += metrics.memory * weights.memory
    totalWeight += weights.memory
  }
  
  if (metrics.gpu !== null) {
    weightedSum += metrics.gpu * weights.gpu
    totalWeight += weights.gpu
  } else {
    // Redistribute GPU weight to CPU and Memory
    if (metrics.cpu !== null) {
      weightedSum += metrics.cpu * (weights.gpu * 0.6)
      totalWeight += weights.gpu * 0.6
    }
    if (metrics.memory !== null) {
      weightedSum += metrics.memory * (weights.gpu * 0.4)
      totalWeight += weights.gpu * 0.4
    }
  }
  
  if (metrics.time !== null) {
    weightedSum += metrics.time * weights.time
    totalWeight += weights.time
  }
  
  if (totalWeight === 0) return null
  
  return weightedSum / totalWeight
}

/**
 * Calculate all efficiency metrics for a job
 * 
 * @param job - Job response data
 * @param report - Job report data (optional)
 * @param elapsed - Elapsed time in seconds
 * @returns Complete efficiency metrics
 * 
 * @TODO Test with complete API response payloads from /jobs/{job_id} endpoint
 * @TODO Write test cases for: completed jobs, running jobs, failed jobs, cancelled jobs
 * @TODO Verify weighted efficiency calculations match expected outputs
 */
export const calculateEfficiencyMetrics = (
  job: JobResponse,
  report: JobReport | undefined,
  elapsed: number
): EfficiencyMetrics => {
  const cpu = calculateCpuEfficiency(job, elapsed)
  const memory = calculateMemoryEfficiency(job)
  const gpu = calculateGpuEfficiency(report)
  const time = calculateTimeEfficiency(elapsed, job.time_limit)
  
  const metrics: EfficiencyMetrics = {
    cpu,
    memory,
    gpu,
    time,
    overall: null,
  }
  
  metrics.overall = calculateOverallEfficiency(metrics)
  
  return metrics
}

/**
 * Get efficiency color based on percentage
 * 
 * @param efficiency - Efficiency percentage (0-100)
 * @returns Chakra UI color palette name
 */
export const getEfficiencyColor = (efficiency: number | null): string => {
  if (efficiency === null) return 'gray'
  if (efficiency >= 80) return 'green'
  if (efficiency >= 60) return 'blue'
  if (efficiency >= 40) return 'yellow'
  if (efficiency >= 20) return 'orange'
  return 'red'
}

/**
 * Get efficiency label based on percentage
 * 
 * @param efficiency - Efficiency percentage (0-100)
 * @returns Human-readable efficiency label
 */
export const getEfficiencyLabel = (efficiency: number | null): string => {
  if (efficiency === null) return 'Unknown'
  if (efficiency >= 80) return 'Excellent'
  if (efficiency >= 60) return 'Good'
  if (efficiency >= 40) return 'Fair'
  if (efficiency >= 20) return 'Poor'
  return 'Critical'
}

/**
 * Calculate CPU hours used
 * 
 * @param elapsed - Elapsed time in seconds
 * @param cpus - Number of CPUs
 * @returns CPU hours
 */
export const calculateCpuHours = (elapsed: number, cpus: number): number => {
  return (elapsed / 3600) * cpus
}

/**
 * Calculate GPU hours used
 * 
 * @param elapsed - Elapsed time in seconds
 * @param gpus - Number of GPUs
 * @returns GPU hours
 */
export const calculateGpuHours = (elapsed: number, gpus: number): number => {
  return (elapsed / 3600) * gpus
}

/**
 * Calculate wasted CPU hours
 * 
 * @param job - Job response data
 * @param elapsed - Elapsed time in seconds
 * @returns Wasted CPU hours
 * 
 * @TODO Test with real job data showing various efficiency levels (20%, 50%, 80%, 100%)
 * @TODO Write test cases for zero waste scenarios and extreme waste scenarios
 */
export const calculateWastedCpuHours = (job: JobResponse, elapsed: number): number => {
  const efficiency = calculateCpuEfficiency(job, elapsed)
  if (efficiency === null || !job.requested_cpus) return 0
  
  const totalCpuHours = calculateCpuHours(elapsed, job.requested_cpus)
  return totalCpuHours * (1 - efficiency / 100)
}

/**
 * Calculate wasted memory (in KiB)
 * 
 * @param job - Job response data
 * @returns Wasted memory in KiB
 * 
 * @TODO Test with real API responses from jobs with different memory usage patterns
 * @TODO Write test cases for multi-node jobs to ensure correct total memory calculation
 */
export const calculateWastedMemory = (job: JobResponse): number => {
  const efficiency = calculateMemoryEfficiency(job)
  if (efficiency === null || !job.requested_memory_per_node || !job.requested_node_count) {
    return 0
  }
  
  const totalMemory = job.requested_memory_per_node * job.requested_node_count
  return totalMemory * (1 - efficiency / 100)
}
