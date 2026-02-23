import type {
  SystemProcessTimeseriesResponse,
  SampleProcessAccResponse,
  SampleProcessGpuAccResponse,
  JobNodeSampleProcessGpuTimeseriesResponse,
  SampleDiskTimeseriesResponse,
  GetClusterByClusterNodesByNodenameDiskstatsTimeseriesResponse,
  NodeDiskTimeseriesResponse,
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
 * Chart data point for Disk I/O metrics
 */
export interface DiskDataPoint {
  time: Date
  timeMs: number
  timeStr: string // For display
  // IOPS (rate calculations)
  read_iops: number
  write_iops: number
  // Throughput in MB/s
  read_throughput_mb: number
  write_throughput_mb: number
  // Latency in ms (average per operation)
  read_latency_ms: number
  write_latency_ms: number
  // Queue depth
  ios_in_progress: number
  // Utilization %
  utilization_pct: number
}

/**
 * Disk metadata and timeseries data
 */
export interface DiskData {
  name: string
  major: number
  minor: number
  data: DiskDataPoint[]
}

/**
 * Transformed disk timeseries by disk name
 */
export interface DiskTimeseriesByName {
  [diskName: string]: DiskData
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
 * Extract a mapping of GPU UUID to node name from the raw timeseries response.
 * This information is discarded by transformGpuTimeseries, so this function
 * provides it separately for display and downstream fetching.
 */
export const extractGpuNodeMapping = (
  response: JobNodeSampleProcessGpuTimeseriesResponse | undefined
): Record<string, string> => {
  if (!response?.nodes) return {}

  const mapping: Record<string, string> = {}
  Object.entries(response.nodes).forEach(([nodeName, nodeData]) => {
    Object.keys(nodeData.gpus).forEach((uuid) => {
      mapping[uuid] = nodeName
    })
  })
  return mapping
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

/**
 * Transform disk timeseries response to organized data by disk name
 * Calculates rate-based metrics (IOPS, throughput, latency, utilization) from cumulative counters
 * 
 * @param response - API response from diskstats/timeseries endpoint
 * @param nodename - Node name to extract data for
 * @returns Object mapping disk names to their transformed timeseries data
 * 
 * @TODO Test with real API response from /nodes/{nodename}/diskstats/timeseries endpoint
 * @TODO Test with multi-disk scenarios (sda, sdb, nvme0n1, etc.)
 * @TODO Write test cases for: no disks, single disk, sparse data, counter wraparound
 * @TODO Verify rate calculations match expected I/O metrics
 */
export const transformDiskstatsTimeseries = (
  response: GetClusterByClusterNodesByNodenameDiskstatsTimeseriesResponse | undefined,
  nodename: string
): DiskTimeseriesByName => {
  if (!response) return {}
  
  const map = response as Record<string, SampleDiskTimeseriesResponse[]>
  const diskArrays = map[nodename] ?? map[Object.keys(map)[0]] ?? []
  
  const result: DiskTimeseriesByName = {}

  diskArrays.forEach((disk) => {
    const points: DiskDataPoint[] = []

    for (let i = 0; i < disk.data.length; i++) {
      const current = disk.data[i]
      const prev = i > 0 ? disk.data[i - 1] : null

      if (!current?.time) continue

      const time = new Date(current.time)
      const timeMs = time.getTime()
      const timeStr = time.toLocaleTimeString()

      // The API may return either cumulative counters or per-interval deltas.
      // If delta_time_in_s > 0, the values are already deltas for that interval — divide to get rates.
      // Otherwise, compute deltas from the previous sample (cumulative counters).
      const deltaTimeSec = (current as { delta_time_in_s?: number }).delta_time_in_s

      let read_iops = 0
      let write_iops = 0
      let read_throughput_mb = 0
      let write_throughput_mb = 0
      let read_latency_ms = 0
      let write_latency_ms = 0
      let utilization_pct = 0

      if (deltaTimeSec && deltaTimeSec > 0) {
        // Values are per-interval deltas — divide by interval to get rates
        read_iops = current.reads_completed / deltaTimeSec
        write_iops = current.writes_completed / deltaTimeSec
        read_throughput_mb = (current.sectors_read * 512) / (1024 * 1024) / deltaTimeSec
        write_throughput_mb = (current.sectors_written * 512) / (1024 * 1024) / deltaTimeSec
        if (current.reads_completed > 0) {
          read_latency_ms = current.ms_spent_reading / current.reads_completed
        }
        if (current.writes_completed > 0) {
          write_latency_ms = current.ms_spent_writing / current.writes_completed
        }
        utilization_pct = Math.min(100, (current.ms_spent_doing_ios / (deltaTimeSec * 1000)) * 100)
      } else if (prev?.time) {
        // Cumulative counters — compute deltas from previous sample
        const timeDeltaSec = (timeMs - new Date(prev.time).getTime()) / 1000

        if (timeDeltaSec > 0) {
          const readsDelta = current.reads_completed - prev.reads_completed
          const writesDelta = current.writes_completed - prev.writes_completed
          read_iops = readsDelta / timeDeltaSec
          write_iops = writesDelta / timeDeltaSec

          const sectorsDelta = current.sectors_read - prev.sectors_read
          const sectorsWrittenDelta = current.sectors_written - prev.sectors_written
          read_throughput_mb = (sectorsDelta * 512) / (1024 * 1024) / timeDeltaSec
          write_throughput_mb = (sectorsWrittenDelta * 512) / (1024 * 1024) / timeDeltaSec

          if (readsDelta > 0) {
            read_latency_ms = (current.ms_spent_reading - prev.ms_spent_reading) / readsDelta
          }
          if (writesDelta > 0) {
            write_latency_ms = (current.ms_spent_writing - prev.ms_spent_writing) / writesDelta
          }

          const ioMsDelta = current.ms_spent_doing_ios - prev.ms_spent_doing_ios
          utilization_pct = Math.min(100, (ioMsDelta / (timeDeltaSec * 1000)) * 100)
        }
      }

      points.push({
        time,
        timeMs,
        timeStr,
        read_iops: Math.max(0, read_iops),
        write_iops: Math.max(0, write_iops),
        read_throughput_mb: Math.max(0, read_throughput_mb),
        write_throughput_mb: Math.max(0, write_throughput_mb),
        read_latency_ms: Math.max(0, read_latency_ms),
        write_latency_ms: Math.max(0, write_latency_ms),
        ios_in_progress: current.ios_currently_in_progress,
        utilization_pct: Math.max(0, utilization_pct),
      })
    }
    
    result[disk.name] = {
      name: disk.name,
      major: disk.major,
      minor: disk.minor,
      data: points,
    }
  })

  return result
}

/**
 * Calculate disk I/O summary statistics
 * 
 * @TODO Test with real transformed disk timeseries data
 * @TODO Write test cases for: empty data, single point, high I/O scenarios
 * @TODO Verify statistical calculations match expected values
 */
export const calculateDiskStats = (
  data: DiskDataPoint[]
): {
  avgReadIOPS: number
  maxReadIOPS: number
  avgWriteIOPS: number
  maxWriteIOPS: number
  avgReadThroughput: number
  maxReadThroughput: number
  avgWriteThroughput: number
  maxWriteThroughput: number
  avgUtilization: number
  maxUtilization: number
} => {
  if (data.length === 0) {
    return {
      avgReadIOPS: 0,
      maxReadIOPS: 0,
      avgWriteIOPS: 0,
      maxWriteIOPS: 0,
      avgReadThroughput: 0,
      maxReadThroughput: 0,
      avgWriteThroughput: 0,
      maxWriteThroughput: 0,
      avgUtilization: 0,
      maxUtilization: 0,
    }
  }

  const sum = data.reduce(
    (acc, point) => ({
      read_iops: acc.read_iops + point.read_iops,
      write_iops: acc.write_iops + point.write_iops,
      read_throughput: acc.read_throughput + point.read_throughput_mb,
      write_throughput: acc.write_throughput + point.write_throughput_mb,
      utilization: acc.utilization + point.utilization_pct,
    }),
    { read_iops: 0, write_iops: 0, read_throughput: 0, write_throughput: 0, utilization: 0 }
  )

  return {
    avgReadIOPS: sum.read_iops / data.length,
    maxReadIOPS: Math.max(...data.map(d => d.read_iops)),
    avgWriteIOPS: sum.write_iops / data.length,
    maxWriteIOPS: Math.max(...data.map(d => d.write_iops)),
    avgReadThroughput: sum.read_throughput / data.length,
    maxReadThroughput: Math.max(...data.map(d => d.read_throughput_mb)),
    avgWriteThroughput: sum.write_throughput / data.length,
    maxWriteThroughput: Math.max(...data.map(d => d.write_throughput_mb)),
    avgUtilization: sum.utilization / data.length,
    maxUtilization: Math.max(...data.map(d => d.utilization_pct)),
  }
}

/**
 * Chart data point for cluster-wide disk I/O aggregated across all nodes
 */
export interface ClusterDiskDataPoint {
  time: string
  avgUtilization: number
  maxUtilization: number
  readIOPS: number
  writeIOPS: number
}

/**
 * Transform cluster-wide disk stats timeseries response into chart-ready data.
 * Aggregates across all nodes and all disks at each timestamp.
 *
 * The response is `Record<nodename, SampleDiskTimeseriesResponse[]>` where each
 * SampleDiskTimeseriesResponse has a `data` array of raw disk samples.
 */
export const transformClusterDiskstatsTimeseries = (
  response: NodeDiskTimeseriesResponse | undefined
): ClusterDiskDataPoint[] => {
  if (!response) return []

  // Aggregate by timestamp across all nodes and disks
  const tsMap = new Map<number, {
    totalUtil: number
    maxUtil: number
    totalReadIOPS: number
    totalWriteIOPS: number
    count: number
  }>()

  for (const diskArrays of Object.values(response)) {
    if (!Array.isArray(diskArrays)) continue

    for (const disk of diskArrays) {
      if (!Array.isArray(disk.data)) continue

      for (let i = 0; i < disk.data.length; i++) {
        const current = disk.data[i]
        const prev = i > 0 ? disk.data[i - 1] : null
        if (!current?.time) continue

        const time = new Date(current.time)
        const ts = Math.floor(time.getTime() / 1000)

        const deltaTimeSec = (current as { delta_time_in_s?: number }).delta_time_in_s

        let utilization = 0
        let readIops = 0
        let writeIops = 0

        if (deltaTimeSec && deltaTimeSec > 0) {
          utilization = Math.min(100, (current.ms_spent_doing_ios / (deltaTimeSec * 1000)) * 100)
          readIops = current.reads_completed / deltaTimeSec
          writeIops = current.writes_completed / deltaTimeSec
        } else if (prev?.time) {
          const timeDeltaSec = (time.getTime() - new Date(prev.time).getTime()) / 1000
          if (timeDeltaSec > 0) {
            const ioMsDelta = current.ms_spent_doing_ios - prev.ms_spent_doing_ios
            utilization = Math.min(100, (ioMsDelta / (timeDeltaSec * 1000)) * 100)
            readIops = (current.reads_completed - prev.reads_completed) / timeDeltaSec
            writeIops = (current.writes_completed - prev.writes_completed) / timeDeltaSec
          }
        }

        utilization = Math.max(0, utilization)
        readIops = Math.max(0, readIops)
        writeIops = Math.max(0, writeIops)

        if (!tsMap.has(ts)) {
          tsMap.set(ts, { totalUtil: 0, maxUtil: 0, totalReadIOPS: 0, totalWriteIOPS: 0, count: 0 })
        }
        const entry = tsMap.get(ts)!
        entry.totalUtil += utilization
        entry.maxUtil = Math.max(entry.maxUtil, utilization)
        entry.totalReadIOPS += readIops
        entry.totalWriteIOPS += writeIops
        entry.count++
      }
    }
  }

  return Array.from(tsMap.entries())
    .sort((a, b) => a[0] - b[0])
    .map(([ts, data]) => ({
      time: new Date(ts * 1000).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
      avgUtilization: data.count > 0 ? Math.round(data.totalUtil / data.count) : 0,
      maxUtilization: Math.round(data.maxUtil),
      readIOPS: data.count > 0 ? Math.round(data.totalReadIOPS / data.count) : 0,
      writeIOPS: data.count > 0 ? Math.round(data.totalWriteIOPS / data.count) : 0,
    }))
}
