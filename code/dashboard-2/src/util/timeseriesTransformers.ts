import type { 
  SystemProcessTimeseriesResponse, 
  SampleProcessAccResponse,
  SampleProcessGpuAccResponse,
  JobNodeSampleProcessGpuTimeseriesResponse
} from '../client'

/**
 * Chart data point for CPU/Memory metrics
 */
export interface CpuMemoryDataPoint {
  time: Date
  timeStr: string // For display
  cpu_avg: number
  cpu_util: number
  cpu_time: number
  memory_resident: number
  memory_virtual: number
  memory_util: number
  processes_avg: number
}

/**
 * Chart data point for GPU metrics
 */
export interface GpuDataPoint {
  time: Date
  timeStr: string // For display
  gpu_memory: number
  gpu_util: number
  gpu_memory_util: number
  pids: number[]
}

/**
 * Transformed GPU timeseries by UUID
 */
export interface GpuTimeseriesByUuid {
  [uuid: string]: GpuDataPoint[]
}

/**
 * Transform SystemProcessTimeseriesResponse to flat array of CPU/Memory data points
 * Aggregates data across all nodes
 * 
 * @TODO Test with real API response from /jobs/{job_id}/process/timeseries endpoint
 * @TODO Test with multi-node job responses to verify proper aggregation across nodes
 * @TODO Write test cases for: empty response, single node, multiple nodes, missing data
 * @TODO Verify timestamp handling and timezone conversions
 */
export const transformProcessTimeseries = (
  response: SystemProcessTimeseriesResponse | undefined
): CpuMemoryDataPoint[] => {
  if (!response?.nodes) return []

  // Group samples by timestamp across all nodes
  const timeMap = new Map<number, SampleProcessAccResponse[]>()

  Object.values(response.nodes).forEach(nodeData => {
    nodeData.cpu_memory.forEach(sample => {
      const timestamp = new Date(sample.time).getTime()
      if (!timeMap.has(timestamp)) {
        timeMap.set(timestamp, [])
      }
      timeMap.get(timestamp)!.push(sample)
    })
  })

  // Aggregate samples at each timestamp
  const dataPoints: CpuMemoryDataPoint[] = []
  const sortedTimes = Array.from(timeMap.keys()).sort((a, b) => a - b)

  sortedTimes.forEach(timestamp => {
    const samples = timeMap.get(timestamp)!
    const count = samples.length

    // Average all metrics across nodes
    const aggregated = samples.reduce(
      (acc, sample) => ({
        cpu_avg: acc.cpu_avg + sample.cpu_avg,
        cpu_util: acc.cpu_util + sample.cpu_util,
        cpu_time: acc.cpu_time + sample.cpu_time,
        memory_resident: acc.memory_resident + sample.memory_resident,
        memory_virtual: acc.memory_virtual + sample.memory_virtual,
        memory_util: acc.memory_util + sample.memory_util,
        processes_avg: acc.processes_avg + sample.processes_avg,
      }),
      {
        cpu_avg: 0,
        cpu_util: 0,
        cpu_time: 0,
        memory_resident: 0,
        memory_virtual: 0,
        memory_util: 0,
        processes_avg: 0,
      }
    )

    const time = new Date(timestamp)
    dataPoints.push({
      time,
      timeStr: time.toLocaleTimeString(),
      cpu_avg: aggregated.cpu_avg / count,
      cpu_util: aggregated.cpu_util / count,
      cpu_time: aggregated.cpu_time / count,
      memory_resident: (aggregated.memory_resident / count) / 1024, // Convert KiB to MB
      memory_virtual: (aggregated.memory_virtual / count) / 1024, // Convert KiB to MB
      memory_util: aggregated.memory_util / count,
      processes_avg: aggregated.processes_avg / count,
    })
  })

  return dataPoints
}

/**
 * Transform GPU timeseries response to organized data by GPU UUID
 * Aggregates data across all nodes for each GPU
 * 
 * @TODO Test with real API response from /jobs/{job_id}/process/gpu/timeseries endpoint
 * @TODO Test with multi-GPU, multi-node scenarios to verify UUID-based grouping
 * @TODO Write test cases for: no GPUs, single GPU, multiple GPUs per node, GPUs across nodes
 * @TODO Verify PID deduplication logic works correctly
 */
export const transformGpuTimeseries = (
  response: JobNodeSampleProcessGpuTimeseriesResponse | undefined
): GpuTimeseriesByUuid => {
  if (!response?.nodes) return {}

  const gpuMap = new Map<string, Map<number, SampleProcessGpuAccResponse[]>>()

  // Collect samples grouped by UUID and timestamp
  Object.values(response.nodes).forEach(nodeData => {
    Object.entries(nodeData.gpus).forEach(([uuid, samples]) => {
      if (!gpuMap.has(uuid)) {
        gpuMap.set(uuid, new Map())
      }
      const uuidTimeMap = gpuMap.get(uuid)!

      samples.forEach(sample => {
        const timestamp = new Date(sample.time).getTime()
        if (!uuidTimeMap.has(timestamp)) {
          uuidTimeMap.set(timestamp, [])
        }
        uuidTimeMap.get(timestamp)!.push(sample)
      })
    })
  })

  // Transform to final structure
  const result: GpuTimeseriesByUuid = {}

  gpuMap.forEach((timeMap, uuid) => {
    const sortedTimes = Array.from(timeMap.keys()).sort((a, b) => a - b)
    
    result[uuid] = sortedTimes.map(timestamp => {
      const samples = timeMap.get(timestamp)!
      const count = samples.length

      // Average metrics across nodes (if GPU appears on multiple nodes)
      const aggregated = samples.reduce(
        (acc, sample) => ({
          gpu_memory: acc.gpu_memory + sample.gpu_memory,
          gpu_util: acc.gpu_util + sample.gpu_util,
          gpu_memory_util: acc.gpu_memory_util + sample.gpu_memory_util,
          pids: [...acc.pids, ...sample.pids],
        }),
        {
          gpu_memory: 0,
          gpu_util: 0,
          gpu_memory_util: 0,
          pids: [] as number[],
        }
      )

      const time = new Date(timestamp)
      return {
        time,
        timeStr: time.toLocaleTimeString(),
        gpu_memory: (aggregated.gpu_memory / count) / 1024, // Convert KiB to MB
        gpu_util: aggregated.gpu_util / count,
        gpu_memory_util: aggregated.gpu_memory_util / count,
        pids: [...new Set(aggregated.pids)], // Deduplicate PIDs
      }
    })
  })

  return result
}

/**
 * Calculate summary statistics from timeseries data
 * 
 * @TODO Test with real transformed timeseries data from various job types
 * @TODO Write test cases for: empty data, single point, sparse data, high-variance data
 * @TODO Verify statistical calculations (avg, max) match expected values
 */
export const calculateTimeseriesStats = (
  data: CpuMemoryDataPoint[]
): {
  avgCpuUtil: number
  maxCpuUtil: number
  avgMemoryResident: number
  maxMemoryResident: number
  avgProcesses: number
} => {
  if (data.length === 0) {
    return {
      avgCpuUtil: 0,
      maxCpuUtil: 0,
      avgMemoryResident: 0,
      maxMemoryResident: 0,
      avgProcesses: 0,
    }
  }

  const sum = data.reduce(
    (acc, point) => ({
      cpu_util: acc.cpu_util + point.cpu_util,
      memory_resident: acc.memory_resident + point.memory_resident,
      processes_avg: acc.processes_avg + point.processes_avg,
    }),
    { cpu_util: 0, memory_resident: 0, processes_avg: 0 }
  )

  const maxMemory = Math.max(...data.map(d => d.memory_resident))
  const maxCpu = Math.max(...data.map(d => d.cpu_util))

  return {
    avgCpuUtil: sum.cpu_util / data.length,
    maxCpuUtil: maxCpu,
    avgMemoryResident: sum.memory_resident / data.length,
    maxMemoryResident: maxMemory,
    avgProcesses: sum.processes_avg / data.length,
  }
}

/**
 * Calculate GPU summary statistics
 */
export const calculateGpuStats = (
  data: GpuDataPoint[]
): {
  avgGpuUtil: number
  maxGpuUtil: number
  avgGpuMemory: number
  maxGpuMemory: number
} => {
  if (data.length === 0) {
    return {
      avgGpuUtil: 0,
      maxGpuUtil: 0,
      avgGpuMemory: 0,
      maxGpuMemory: 0,
    }
  }

  const sum = data.reduce(
    (acc, point) => ({
      gpu_util: acc.gpu_util + point.gpu_util,
      gpu_memory: acc.gpu_memory + point.gpu_memory,
    }),
    { gpu_util: 0, gpu_memory: 0 }
  )

  return {
    avgGpuUtil: sum.gpu_util / data.length,
    maxGpuUtil: Math.max(...data.map(d => d.gpu_util)),
    avgGpuMemory: sum.gpu_memory / data.length,
    maxGpuMemory: Math.max(...data.map(d => d.gpu_memory)),
  }
}

/**
 * Downsample large datasets for better chart performance
 * Uses Largest-Triangle-Three-Buckets (LTTB) algorithm for better visual accuracy
 * 
 * @param data - Array of data points
 * @param maxPoints - Maximum number of points to keep (default: 500)
 * @returns Downsampled array
 */
export function downsampleTimeseries<T extends { time: Date }>(
  data: T[],
  maxPoints: number = 500
): T[] {
  if (data.length <= maxPoints) {
    return data
  }

  // Always keep first and last points
  const sampled: T[] = [data[0]]
  const bucketSize = (data.length - 2) / (maxPoints - 2)

  for (let i = 0; i < maxPoints - 2; i++) {
    const bucketStart = Math.floor(i * bucketSize) + 1
    const bucketEnd = Math.floor((i + 1) * bucketSize) + 1
    const bucketLength = bucketEnd - bucketStart

    // Calculate average point in this bucket
    let avgIndex = bucketStart
    if (bucketLength > 1) {
      avgIndex = bucketStart + Math.floor(bucketLength / 2)
    }

    sampled.push(data[avgIndex])
  }

  sampled.push(data[data.length - 1])
  return sampled
}

/**
 * Calculate optimal max points based on chart width
 * Ensures reasonable density while maintaining performance
 */
export function calculateOptimalMaxPoints(chartWidth: number): number {
  // Aim for roughly 2 points per pixel, capped at 1000
  return Math.min(Math.max(Math.floor(chartWidth / 2), 200), 1000)
}
