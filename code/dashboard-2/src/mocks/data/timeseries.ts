import { CLUSTER_NAME, NODE_NAMES } from './cluster'

// We use our own "wire format" types here because the generated types use
// `Date` for time fields, but MSW returns raw JSON (ISO strings).
// The app's response transformers handle the string -> Date conversion.

type WireSampleProcessAcc = {
  time: string; memory_resident: number; memory_virtual: number;
  memory_util: number; cpu_avg: number; cpu_util: number;
  cpu_time: number; processes_avg: number;
}

type WireSampleGpuBase = {
  time: string; failing: number; fan: number; compute_mode: string;
  performance_state: number; memory: number; memory_util: number;
  memory_clock: number; ce_util: number; ce_clock: number;
  temperature: number; power: number; power_limit: number;
}

type WireSampleGpuTimeseries = { uuid: string; index: number; data: WireSampleGpuBase[] }
type WireSampleDisk = {
  time: string; reads_completed: number; reads_merged: number;
  sectors_read: number; ms_spent_reading: number;
  writes_completed: number; writes_merged: number;
  sectors_written: number; ms_spent_writing: number;
  ios_currently_in_progress: number; ms_spent_doing_ios: number;
  weighted_ms_spent_doing_ios: number; discards_completed: number;
  discards_merged: number; sectors_discarded: number;
  ms_spent_discarding: number; flush_requests_completed: number;
  ms_spent_flushing: number; delta_time_in_s?: number;
}
type WireSampleDiskTimeseries = { name: string; major: number; minor: number; field_names?: string[]; data: WireSampleDisk[] }
type WireSampleProcessGpuAcc = { time: string; gpu_memory: number; gpu_util: number; gpu_memory_util: number; pids: number[] }

// Re-export cluster name for consumers that need it alongside timeseries data
export { CLUSTER_NAME }

// ---------------------------------------------------------------------------
// Seeded PRNG (linear congruential generator)
// ---------------------------------------------------------------------------

function createRng(seed: number) {
  let state = seed
  return () => {
    // Parameters from Numerical Recipes
    state = (state * 1664525 + 1013904223) & 0xffffffff
    return (state >>> 0) / 0xffffffff
  }
}

function hashString(s: string): number {
  let hash = 5381
  for (let i = 0; i < s.length; i++) {
    hash = ((hash << 5) + hash + s.charCodeAt(i)) & 0xffffffff
  }
  return hash >>> 0
}

function seedForNode(nodeName: string): number {
  return hashString(nodeName)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Clamp a number between min and max. */
function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

/** Linear interpolation. */
function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t
}

/** Generate an array of ISO 8601 timestamps at `intervalMinutes` intervals,
 *  ending approximately at "now" and spanning `hours` hours. */
function generateTimestamps(hours: number, intervalMinutes: number): string[] {
  const now = new Date()
  // Round down to nearest interval
  now.setMinutes(
    Math.floor(now.getMinutes() / intervalMinutes) * intervalMinutes,
    0,
    0,
  )
  const count = Math.floor((hours * 60) / intervalMinutes)
  const startMs = now.getTime() - count * intervalMinutes * 60_000
  const timestamps: string[] = []
  for (let i = 0; i <= count; i++) {
    timestamps.push(new Date(startMs + i * intervalMinutes * 60_000).toISOString())
  }
  return timestamps
}

/** Classify a node name into a category. */
type NodeCategory = 'cpu' | 'gpu' | 'highmem' | 'down'

const DOWN_NODES = new Set(['cn018', 'cn019'])

function classifyNode(name: string): NodeCategory {
  if (DOWN_NODES.has(name)) return 'down'
  if (name.startsWith('gpu')) return 'gpu'
  if (name.startsWith('hi')) return 'highmem'
  return 'cpu'
}

/** Diurnal pattern: lower at night (hours 0-6), higher during day.
 *  `hourOfDay` is the UTC hour. */
function diurnalFactor(hourOfDay: number): number {
  // Peak at 14:00, trough at 03:00
  return 0.5 + 0.5 * Math.sin(((hourOfDay - 3) / 24) * 2 * Math.PI - Math.PI / 2)
}

// ---------------------------------------------------------------------------
// GPU helpers
// ---------------------------------------------------------------------------

const GPU_NODES = NODE_NAMES.filter((n) => n.startsWith('gpu'))
const GPUS_PER_NODE = 4
const GPU_MEMORY_TOTAL_KIB = 80 * 1024 * 1024 // 80 GB in KiB

function gpuUuid(nodeName: string, gpuIndex: number): string {
  const nodeNum = nodeName.replace(/\D/g, '').padStart(3, '0')
  const gpuNum = String(gpuIndex).padStart(2, '0')
  return `GPU-gpu${nodeNum}-${gpuNum}-0000-0000-000000000000`
}

// ---------------------------------------------------------------------------
// CPU Timeseries
// ---------------------------------------------------------------------------

/** Generate CPU-focused timeseries for a single node. */
export function generateNodeCpuTimeseries(
  nodeName: string,
  hours: number = 24,
): WireSampleProcessAcc[] {
  const category = classifyNode(nodeName)
  if (category === 'down') return []

  const rng = createRng(seedForNode(nodeName + ':cpu'))
  const timestamps = generateTimestamps(hours, 5)

  // Pre-generate phase offsets for naturalSignal (each call uses 3 rng values)
  // We call naturalSignal fresh each time, but seeded rng ensures determinism.
  // To get consistent phase offsets, generate them once.
  const phaseOffsets = [rng() * Math.PI * 2, rng() * Math.PI * 2, rng() * Math.PI * 2]
  const memPhaseOffsets = [rng() * Math.PI * 2, rng() * Math.PI * 2, rng() * Math.PI * 2]

  // Utilization ranges by category
  const [cpuMin, cpuMax] =
    category === 'gpu'
      ? [10, 50]
      : category === 'highmem'
        ? [30, 60]
        : [20, 90]

  // Memory: proportional to node type, 40-80% utilization
  const memoryTotalKib =
    category === 'highmem'
      ? 2 * 1024 * 1024 * 1024 // 2 TiB
      : category === 'gpu'
        ? 512 * 1024 * 1024 // 512 GiB
        : 256 * 1024 * 1024 // 256 GiB

  let cpuTimeAccum = 0

  return timestamps.map((time, i) => {
    const t = i / (timestamps.length - 1 || 1)
    const hour = new Date(time).getUTCHours()

    // CPU utilization: sinusoidal + diurnal + noise
    let cpuSignal = 0
    const freqs = [1, 2.3, 5.7]
    for (let f = 0; f < freqs.length; f++) {
      cpuSignal += (1 / (f + 1)) * Math.sin(2 * Math.PI * freqs[f] * t + phaseOffsets[f])
    }
    const maxAmp = freqs.reduce((sum, _, f) => sum + 1 / (f + 1), 0)
    cpuSignal = (cpuSignal / maxAmp + 1) / 2
    cpuSignal *= diurnalFactor(hour)
    cpuSignal += (rng() - 0.5) * 0.1
    cpuSignal = clamp(cpuSignal, 0, 1)

    const cpuUtil = lerp(cpuMin, cpuMax, cpuSignal)
    cpuTimeAccum += (cpuUtil / 100) * 300 // 5 minutes in seconds

    // Memory utilization: smoother, step-like
    let memSignal = 0
    for (let f = 0; f < freqs.length; f++) {
      memSignal += (1 / (f + 1)) * Math.sin(2 * Math.PI * freqs[f] * 0.3 * t + memPhaseOffsets[f])
    }
    memSignal = (memSignal / maxAmp + 1) / 2
    memSignal = clamp(memSignal * 0.4 + 0.4, 0.4, 0.8) // Range 40-80%
    memSignal += (rng() - 0.5) * 0.02

    const memoryUtil = clamp(memSignal * 100, 40, 80)
    const memoryResident = Math.round(memoryTotalKib * (memoryUtil / 100))
    const memoryVirtual = Math.round(memoryResident * (1.2 + rng() * 0.3))

    const processesAvg = category === 'gpu' ? 5 + rng() * 10 : 20 + rng() * 80

    return {
      time,
      memory_resident: memoryResident,
      memory_virtual: memoryVirtual,
      memory_util: Math.round(memoryUtil * 100) / 100,
      cpu_avg: Math.round(cpuUtil * 0.95 * 100) / 100,
      cpu_util: Math.round(cpuUtil * 100) / 100,
      cpu_time: Math.round(cpuTimeAccum),
      processes_avg: Math.round(processesAvg * 100) / 100,
    }  })
}

/** Generate CPU timeseries for multiple nodes. */
export function generateCpuTimeseries(
  nodeNames: string[] = NODE_NAMES,
  hours: number = 24,
): Record<string, WireSampleProcessAcc[]> {
  const result: Record<string, WireSampleProcessAcc[]> = {}
  for (const name of nodeNames) {
    result[name] = generateNodeCpuTimeseries(name, hours)
  }
  return result
}

// ---------------------------------------------------------------------------
// Memory Timeseries
// ---------------------------------------------------------------------------

/** Generate memory-focused timeseries for a single node. */
export function generateNodeMemoryTimeseries(
  nodeName: string,
  hours: number = 24,
): WireSampleProcessAcc[] {
  const category = classifyNode(nodeName)
  if (category === 'down') return []

  const rng = createRng(seedForNode(nodeName + ':mem'))
  const timestamps = generateTimestamps(hours, 5)

  // Memory: higher utilization with step-function behavior
  const memoryTotalKib =
    category === 'highmem'
      ? 2 * 1024 * 1024 * 1024
      : category === 'gpu'
        ? 512 * 1024 * 1024
        : 256 * 1024 * 1024

  // Pre-generate step changes
  const numSteps = 3 + Math.floor(rng() * 5)
  const stepPositions: number[] = []
  const stepLevels: number[] = []
  for (let i = 0; i < numSteps; i++) {
    stepPositions.push(rng())
    stepLevels.push(0.5 + rng() * 0.4) // 50-90%
  }
  // Sort step positions
  const stepIndices = stepPositions.map((_, i) => i).sort((a, b) => stepPositions[a] - stepPositions[b])

  let cpuTimeAccum = 0

  return timestamps.map((time, i) => {
    const t = i / (timestamps.length - 1 || 1)

    // Memory utilization: step function with small noise
    let currentLevel = stepLevels[stepIndices[0]]
    for (const idx of stepIndices) {
      if (t >= stepPositions[idx]) {
        currentLevel = stepLevels[idx]
      }
    }
    const memoryUtil = clamp((currentLevel + (rng() - 0.5) * 0.02) * 100, 50, 95)
    const memoryResident = Math.round(memoryTotalKib * (memoryUtil / 100))
    const memoryVirtual = Math.round(memoryResident * (1.3 + rng() * 0.2))

    // CPU: lower for memory-focused view
    const cpuUtil = clamp(10 + rng() * 20, 10, 30)
    cpuTimeAccum += (cpuUtil / 100) * 300

    const processesAvg = 10 + rng() * 40

    return {
      time,
      memory_resident: memoryResident,
      memory_virtual: memoryVirtual,
      memory_util: Math.round(memoryUtil * 100) / 100,
      cpu_avg: Math.round(cpuUtil * 0.9 * 100) / 100,
      cpu_util: Math.round(cpuUtil * 100) / 100,
      cpu_time: Math.round(cpuTimeAccum),
      processes_avg: Math.round(processesAvg * 100) / 100,
    }  })
}

/** Generate memory timeseries for multiple nodes. */
export function generateMemoryTimeseries(
  nodeNames: string[] = NODE_NAMES,
  hours: number = 24,
): Record<string, WireSampleProcessAcc[]> {
  const result: Record<string, WireSampleProcessAcc[]> = {}
  for (const name of nodeNames) {
    result[name] = generateNodeMemoryTimeseries(name, hours)
  }
  return result
}

// ---------------------------------------------------------------------------
// GPU Timeseries
// ---------------------------------------------------------------------------

/** Generate GPU timeseries for a single node. Returns empty array for non-GPU nodes. */
export function generateNodeGpuTimeseries(
  nodeName: string,
  hours: number = 24,
): WireSampleGpuTimeseries[] {
  if (!nodeName.startsWith('gpu')) return []

  const rng = createRng(seedForNode(nodeName + ':gpu'))
  const timestamps = generateTimestamps(hours, 5)

  // Decide which GPUs are "bursty" (1-2 per node)
  const burstyGpus = new Set<number>()
  const numBursty = 1 + (rng() > 0.5 ? 1 : 0)
  while (burstyGpus.size < numBursty) {
    burstyGpus.add(Math.floor(rng() * GPUS_PER_NODE))
  }

  const gpus: WireSampleGpuTimeseries[] = []

  for (let gpuIdx = 0; gpuIdx < GPUS_PER_NODE; gpuIdx++) {
    const gpuRng = createRng(seedForNode(nodeName + ':gpu:' + gpuIdx))
    const isBursty = burstyGpus.has(gpuIdx)

    // Pre-generate phase offsets
    const phases = [gpuRng() * Math.PI * 2, gpuRng() * Math.PI * 2, gpuRng() * Math.PI * 2]

    const data: (Omit<WireSampleGpuBase, 'time'> & { time: string })[] = timestamps.map((time, i) => {
      const t = i / (timestamps.length - 1 || 1)

      let ceUtil: number
      if (isBursty) {
        // Bimodal: either near-zero or high
        const burstSignal = Math.sin(2 * Math.PI * 3 * t + phases[0])
        ceUtil = burstSignal > 0.2 ? lerp(70, 100, gpuRng()) : lerp(0, 15, gpuRng())
      } else {
        // Steady high utilization (training workload)
        const signal = Math.sin(2 * Math.PI * t + phases[0])
        ceUtil = lerp(70, 95, (signal + 1) / 2) + (gpuRng() - 0.5) * 5
      }
      ceUtil = clamp(ceUtil, 0, 100)

      // Memory usage: high for active GPUs
      const memFraction = ceUtil > 10
        ? lerp(0.6, 0.95, gpuRng())
        : lerp(0.05, 0.2, gpuRng())
      const memory = Math.round(GPU_MEMORY_TOTAL_KIB * memFraction)
      const memoryUtil = Math.round(memFraction * 10000) / 100

      // Temperature correlates with utilization
      const temperature = Math.round(35 + (ceUtil / 100) * 45 + (gpuRng() - 0.5) * 5)

      // Power correlates with utilization
      const powerLimit = 350 + Math.round(gpuRng() * 50) // 350-400W
      const power = Math.round(100 + (ceUtil / 100) * 250 + (gpuRng() - 0.5) * 20)

      // Fan speed correlates with temperature
      const fan = Math.round(clamp(20 + (temperature - 35) * 1.5 + (gpuRng() - 0.5) * 5, 15, 100))

      // Performance state: lower number = higher performance
      const performanceState = ceUtil > 50 ? 0 : ceUtil > 20 ? 2 : 8

      return {
        time,
        failing: 0,
        fan,
        compute_mode: 'Default',
        performance_state: performanceState,
        memory,
        memory_util: memoryUtil,
        memory_clock: ceUtil > 10 ? 1593 : 405,
        ce_util: Math.round(ceUtil * 100) / 100,
        ce_clock: ceUtil > 10 ? 1410 : 210,
        temperature: clamp(temperature, 30, 85),
        power: clamp(power, 50, 400),
        power_limit: powerLimit,
      }
    })

    gpus.push({
      uuid: gpuUuid(nodeName, gpuIdx),
      index: gpuIdx,
      data,
    })
  }

  return gpus
}

/** Generate GPU timeseries for multiple nodes. Only GPU nodes produce data. */
export function generateGpuTimeseries(
  nodeNames: string[] = GPU_NODES,
  hours: number = 24,
): Record<string, WireSampleGpuTimeseries[]> {
  const result: Record<string, WireSampleGpuTimeseries[]> = {}
  for (const name of nodeNames) {
    if (!name.startsWith('gpu')) continue
    result[name] = generateNodeGpuTimeseries(name, hours)
  }
  return result
}

// ---------------------------------------------------------------------------
// Disk Timeseries
// ---------------------------------------------------------------------------

const DISK_CONFIGS: { name: string; major: number; minor: number }[] = [
  { name: 'sda', major: 8, minor: 0 },
  { name: 'nvme0n1', major: 259, minor: 0 },
]

/** Generate disk timeseries for a single node. */
export function generateNodeDiskTimeseries(
  nodeName: string,
  hours: number = 24,
): WireSampleDiskTimeseries[] {
  const category = classifyNode(nodeName)
  if (category === 'down') return []

  const timestamps = generateTimestamps(hours, 5)

  return DISK_CONFIGS.map((disk) => {
    const rng = createRng(seedForNode(nodeName + ':disk:' + disk.name))

    // Pre-generate spike positions (checkpoint writes every ~1-3 hours)
    const numSpikes = Math.floor(hours / 2) + Math.floor(rng() * Math.floor(hours / 3))
    const spikePositions = new Set<number>()
    for (let s = 0; s < numSpikes; s++) {
      spikePositions.add(Math.floor(rng() * timestamps.length))
    }

    const data: (Omit<WireSampleDisk, 'time'> & { time: string })[] = timestamps.map((time, i) => {
      const isSpike = spikePositions.has(i)
      const spikeMultiplier = isSpike ? 10 + rng() * 40 : 1

      // NVMe is faster than HDD
      const diskSpeedFactor = disk.name === 'nvme0n1' ? 3 : 1

      const baseReads = Math.round((50 + rng() * 100) * diskSpeedFactor)
      const baseWrites = Math.round((30 + rng() * 80) * diskSpeedFactor)

      const readsCompleted = Math.round(baseReads * (1 + (rng() - 0.5) * 0.3))
      const writesCompleted = Math.round(baseWrites * spikeMultiplier)
      const sectorsRead = readsCompleted * (8 + Math.floor(rng() * 16))
      const sectorsWritten = writesCompleted * (8 + Math.floor(rng() * 32)) * (isSpike ? 4 : 1)

      return {
        time,
        reads_completed: readsCompleted,
        reads_merged: Math.round(readsCompleted * rng() * 0.3),
        sectors_read: sectorsRead,
        ms_spent_reading: Math.round(readsCompleted * (0.5 + rng() * 2)),
        writes_completed: writesCompleted,
        writes_merged: Math.round(writesCompleted * rng() * 0.2),
        sectors_written: sectorsWritten,
        ms_spent_writing: Math.round(writesCompleted * (0.3 + rng() * 1.5) * spikeMultiplier),
        ios_currently_in_progress: isSpike ? Math.floor(rng() * 8) + 1 : Math.floor(rng() * 2),
        ms_spent_doing_ios: Math.round((readsCompleted + writesCompleted) * (0.2 + rng())),
        weighted_ms_spent_doing_ios: Math.round(
          (readsCompleted + writesCompleted) * (0.3 + rng() * 1.5),
        ),
        discards_completed: Math.floor(rng() * 5),
        discards_merged: 0,
        sectors_discarded: Math.floor(rng() * 1024),
        ms_spent_discarding: Math.floor(rng() * 10),
        flush_requests_completed: Math.floor(rng() * 3),
        ms_spent_flushing: Math.floor(rng() * 5),
        delta_time_in_s: 300,
      }
    })

    return {
      name: disk.name,
      major: disk.major,
      minor: disk.minor,
      data,
    }
  })
}

/** Generate disk timeseries for multiple nodes. */
export function generateDiskTimeseries(
  nodeNames: string[] = NODE_NAMES,
  hours: number = 24,
): Record<string, WireSampleDiskTimeseries[]> {
  const result: Record<string, WireSampleDiskTimeseries[]> = {}
  for (const name of nodeNames) {
    result[name] = generateNodeDiskTimeseries(name, hours)
  }
  return result
}

// ---------------------------------------------------------------------------
// Job Process Timeseries (for Resource Timeline tab)
// ---------------------------------------------------------------------------

/** Generate job-level CPU/memory timeseries for a single node.
 *  Returns the shape expected by SystemProcessTimeseriesResponse.nodes[node]. */
export function generateJobProcessTimeseries(
  nodeName: string,
  hours: number = 6,
): { cpu_memory: WireSampleProcessAcc[]; gpus: Record<string, WireSampleProcessGpuAcc[]> } {
  const rng = createRng(seedForNode(nodeName + ':jobproc'))
  const timestamps = generateTimestamps(hours, 1) // 1-minute intervals for job-level data
  const phases = [rng() * Math.PI * 2, rng() * Math.PI * 2, rng() * Math.PI * 2]

  const cpu_memory: WireSampleProcessAcc[] = timestamps.map((time, i) => {
    const t = i / (timestamps.length - 1 || 1)

    // Training workload: high CPU with periodic spikes (data loading)
    let cpuSignal = 0.7 + 0.2 * Math.sin(2 * Math.PI * 4 * t + phases[0])
    cpuSignal += (rng() - 0.5) * 0.08
    cpuSignal = clamp(cpuSignal, 0.3, 1.0)

    const cpuUtil = cpuSignal * 95
    const cpuAvg = cpuUtil * 0.92

    // Memory: gradually increasing (training accumulation) with small noise
    const memBase = 0.4 + 0.3 * t // 40% -> 70% over time
    const memUtil = clamp((memBase + (rng() - 0.5) * 0.03) * 100, 35, 80)
    const memTotalKib = 512 * 1024 * 1024 // GPU node
    const memResident = Math.round(memTotalKib * (memUtil / 100))
    const memVirtual = Math.round(memResident * (1.25 + rng() * 0.15))

    return {
      time,
      memory_resident: memResident,
      memory_virtual: memVirtual,
      memory_util: Math.round(memUtil * 100) / 100,
      cpu_avg: Math.round(cpuAvg * 100) / 100,
      cpu_util: Math.round(cpuUtil * 100) / 100,
      cpu_time: Math.round(cpuUtil * 60 * (i + 1) / 100),
      processes_avg: Math.round((8 + rng() * 4) * 100) / 100,
    }
  })

  // GPU process-level timeseries
  const gpus: Record<string, WireSampleProcessGpuAcc[]> = {}
  if (nodeName.startsWith('gpu')) {
    for (let gpuIdx = 0; gpuIdx < GPUS_PER_NODE; gpuIdx++) {
      const uuid = gpuUuid(nodeName, gpuIdx)
      const gpuRng = createRng(seedForNode(nodeName + ':jobgpu:' + gpuIdx))
      const gpuPhase = gpuRng() * Math.PI * 2

      gpus[uuid] = timestamps.map((time, i) => {
        const t = i / (timestamps.length - 1 || 1)

        // Training: high, sustained GPU utilization with periodic dips (data loading)
        let gpuUtilSignal = 0.85 + 0.1 * Math.sin(2 * Math.PI * 6 * t + gpuPhase)
        gpuUtilSignal += (gpuRng() - 0.5) * 0.06
        gpuUtilSignal = clamp(gpuUtilSignal, 0.4, 1.0)

        const gpuUtil = gpuUtilSignal * 100
        const memFraction = clamp(0.6 + 0.2 * t + (gpuRng() - 0.5) * 0.05, 0.5, 0.95)
        const gpuMemory = Math.round(GPU_MEMORY_TOTAL_KIB * memFraction)

        return {
          time,
          gpu_memory: gpuMemory,
          gpu_util: Math.round(gpuUtil * 100) / 100,
          gpu_memory_util: Math.round(memFraction * 10000) / 100,
          pids: [1002, 1005, 1006, 1007, 1008],
        }
      })
    }
  }

  return { cpu_memory, gpus }
}

/** Generate the full job process GPU timeseries response.
 *  Returns the shape expected by JobNodeSampleProcessGpuTimeseriesResponse.nodes[node]. */
export function generateJobGpuTimeseries(
  nodeName: string,
  hours: number = 6,
): { gpus: Record<string, WireSampleProcessGpuAcc[]> } {
  const { gpus } = generateJobProcessTimeseries(nodeName, hours)
  return { gpus }
}

// ---------------------------------------------------------------------------
// Process GPU Utilization (latest snapshot)
// ---------------------------------------------------------------------------

/** Generate a latest-snapshot GPU utilization map for GPU nodes. */
export function generateNodeGpuUtil(
  nodeNames: string[] = GPU_NODES,
): Record<string, Record<string, WireSampleProcessGpuAcc>> {
  const result: Record<string, Record<string, WireSampleProcessGpuAcc>> = {}
  const now = new Date().toISOString()

  for (const nodeName of nodeNames) {
    if (!nodeName.startsWith('gpu')) continue

    const rng = createRng(seedForNode(nodeName + ':gpuutil'))
    const gpuMap: Record<string, WireSampleProcessGpuAcc> = {}

    for (let gpuIdx = 0; gpuIdx < GPUS_PER_NODE; gpuIdx++) {
      const uuid = gpuUuid(nodeName, gpuIdx)

      // Most GPUs are active with high utilization
      const isActive = rng() > 0.15
      const gpuUtil = isActive ? lerp(60, 98, rng()) : lerp(0, 10, rng())
      const memFraction = isActive ? lerp(0.5, 0.95, rng()) : lerp(0.01, 0.1, rng())
      const gpuMemory = Math.round(GPU_MEMORY_TOTAL_KIB * memFraction)

      // Generate 1-4 PIDs for active GPUs
      const numPids = isActive ? 1 + Math.floor(rng() * 3) : 0
      const pids: number[] = []
      for (let p = 0; p < numPids; p++) {
        pids.push(10000 + Math.floor(rng() * 90000))
      }

      gpuMap[uuid] = {
        time: now,
        gpu_memory: gpuMemory,
        gpu_util: Math.round(gpuUtil * 100) / 100,
        gpu_memory_util: Math.round(memFraction * 10000) / 100,
        pids,
      }
    }

    result[nodeName] = gpuMap
  }

  return result
}
