import { useMemo, useState, memo } from 'react'
import {
  VStack,
  HStack,
  Box,
  Text,
  Spinner,
  Alert,
  Card,
  SimpleGrid,
  Badge,
  Button,
  Group,
  Progress,
} from '@chakra-ui/react'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Brush,
} from 'recharts'
import type { Client } from '../../client/client/types.gen'
import type { GpuCardResponse, NodeInfoResponse } from '../../client/types.gen'

import { TimeRangePicker, type TimeRange } from '../TimeRangePicker'
import { useJobGpuTimeseries } from '../../hooks/useJobTimeseries'
import { useMultiNodeInfo } from '../../hooks/v2/useNodeQueries'
import {
  transformGpuTimeseries,
  calculateGpuStats,
  extractGpuNodeMapping,
} from '../../util/timeseriesTransformers'
import { formatMemory, formatEfficiency } from '../../util/formatters'
import { timeRangeToTimestamps } from '../../util/timeRangeUtils'
import { getEfficiencyColor, getEfficiencyLabel } from '../../util/efficiency'

interface Props {
  cluster: string;
  jobId: number;
  client: Client | null;
  gpuUuids: string[];
}

const RESOLUTION_OPTIONS = [
  { label: '1s', value: 1 },
  { label: '10s', value: 10 },
  { label: '30s', value: 30 },
  { label: '1m', value: 60 },
  { label: '5m', value: 300 },
]

// Color palette for different GPUs
const GPU_COLORS = [
  '#3182CE', // blue
  '#38A169', // green
  '#E53E3E', // red
  '#805AD5', // purple
  '#DD6B20', // orange
  '#319795', // teal
  '#D69E2E', // yellow
  '#C53030', // dark red
]

interface TooltipPayload {
  name: string;
  value: number;
  color: string;
  unit?: string;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: TooltipPayload[];
  label?: string;
}

/**
 * Custom tooltip for GPU charts
 */
const CustomTooltip = ({ active, payload, label }: CustomTooltipProps) => {
  if (!active || !payload || !payload.length) return null

  return (
    <Card.Root size="sm" p={3} bg="white" borderWidth="1px">
      <Text fontSize="sm" fontWeight="semibold" mb={2}>
        {label}
      </Text>
      {payload.map((entry, index) => (
        <Text key={index} fontSize="sm" color={entry.color}>
          {entry.name}:{' '}
          {typeof entry.value === 'number'
            ? entry.value.toFixed(2)
            : entry.value}
          {entry.unit}
        </Text>
      ))}
    </Card.Root>
  )
}

export const GpuPerformanceTab = memo(({ cluster, jobId, client }: Props) => {
  const [timeRange, setTimeRange] = useState<TimeRange>({
    type: 'relative',
    label: 'Last 1 hour',
    value: '1h',
    endAt: 'now',
  })
  const [resolution, setResolution] = useState(30)
  const [selectedGpu, setSelectedGpu] = useState<string | 'all'>('all')

  // Convert time range to timestamps
  const { startTimeInS, endTimeInS } = useMemo(
    () => timeRangeToTimestamps(timeRange),
    [timeRange],
  )

  const {
    data: timeseriesData,
    isLoading,
    isError,
    error,
    isFetching,
  } = useJobGpuTimeseries({
    cluster,
    jobId,
    client,
    resolution,
    startTimeInS,
    endTimeInS,
  })

  const gpuData = useMemo(
    () => transformGpuTimeseries(timeseriesData?.[0]),
    [timeseriesData],
  )

  // Feature 3: Node-level GPU context (UUID → node name)
  const gpuNodeMapping = useMemo(
    () => extractGpuNodeMapping(timeseriesData?.[0]),
    [timeseriesData],
  )

  // Feature 1: GPU hardware info - fetch node info for all involved nodes
  const uniqueNodeNames = useMemo(
    () => [...new Set(Object.values(gpuNodeMapping))],
    [gpuNodeMapping],
  )

  const nodeInfoQueries = useMultiNodeInfo({
    cluster,
    nodenames: uniqueNodeNames,
    client,
    enabled: uniqueNodeNames.length > 0,
  })

  const gpuCardLookup = useMemo(() => {
    const lookup: Record<string, GpuCardResponse> = {}
    nodeInfoQueries.forEach((query) => {
      if (!query.data) return
      const nodeInfoMap = query.data as Record<string, NodeInfoResponse>
      Object.values(nodeInfoMap).forEach((info) => {
        info.cards?.forEach((card) => {
          lookup[card.uuid] = card
        })
      })
    })
    return lookup
  }, [nodeInfoQueries])

  const gpuStats = useMemo(() => {
    const statsMap: Record<string, ReturnType<typeof calculateGpuStats>> = {}
    Object.entries(gpuData).forEach(([uuid, data]) => {
      statsMap[uuid] = calculateGpuStats(data)
    })
    return statsMap
  }, [gpuData])

  // Feature 2: Process-level breakdown (aggregate all PIDs per GPU)
  const gpuAllPids = useMemo(() => {
    const pidsMap: Record<string, number[]> = {}
    Object.entries(gpuData).forEach(([uuid, data]) => {
      const allPids = new Set<number>()
      data.forEach((point) => {
        point.pids.forEach((pid) => allPids.add(pid))
      })
      pidsMap[uuid] = Array.from(allPids).sort((a, b) => a - b)
    })
    return pidsMap
  }, [gpuData])

  // Feature 4: GPU efficiency metric
  const gpuEfficiency = useMemo(() => {
    const uuids = Object.keys(gpuStats)
    if (uuids.length === 0) return null
    const totalAvgUtil = uuids.reduce((sum, uuid) => sum + gpuStats[uuid].avgGpuUtil, 0)
    return totalAvgUtil / uuids.length
  }, [gpuStats])

  const availableGpus = Object.keys(gpuData)

  // Prepare combined data for all GPUs view
  const combinedUtilData =
    selectedGpu === 'all'
      ? (() => {
        const timeMap = new Map<
            string,
            Record<string, string | Date | number>
          >()
        Object.entries(gpuData).forEach(([uuid, data]) => {
          data.forEach((point) => {
            const key = point.timeStr
            if (!timeMap.has(key)) {
              timeMap.set(key, { timeStr: key, time: point.time })
            }
              timeMap.get(key)![`gpu_util_${uuid.slice(-8)}`] = point.gpu_util
          })
        })
        return Array.from(timeMap.values()).sort(
          (a, b) => (a.time as Date).getTime() - (b.time as Date).getTime(),
        )
      })()
      : gpuData[selectedGpu]

  return (
    <VStack w="100%" gap={6} align="start">
      {/* Controls */}
      <HStack justify="space-between" w="100%" flexWrap="wrap" gap={4}>
        <HStack gap={2}>
          <Text fontSize="lg" fontWeight="semibold">
            GPU Performance
          </Text>
          <Badge colorPalette="blue" size="lg">
            {availableGpus.length} GPU{availableGpus.length > 1 ? 's' : ''}
          </Badge>
          {gpuEfficiency !== null && (
            <Badge colorPalette={getEfficiencyColor(gpuEfficiency)} size="lg">
              Efficiency: {formatEfficiency(gpuEfficiency)} ({getEfficiencyLabel(gpuEfficiency)})
            </Badge>
          )}
          {isFetching && <Spinner size="sm" color="blue.500" />}
        </HStack>

        <HStack gap={4} flexWrap="wrap">
          {/* GPU Selector */}
          <HStack gap={2}>
            <Text fontSize="sm" color="fg.muted">
              GPU:
            </Text>
            <Group attached>
              <Button
                size="sm"
                variant={selectedGpu === 'all' ? 'solid' : 'outline'}
                onClick={() => setSelectedGpu('all')}
              >
                All
              </Button>
              {availableGpus.map((uuid, idx) => (
                <Button
                  key={uuid}
                  size="sm"
                  variant={selectedGpu === uuid ? 'solid' : 'outline'}
                  onClick={() => setSelectedGpu(uuid)}
                >
                  GPU {idx + 1}
                </Button>
              ))}
            </Group>
          </HStack>

          {/* Resolution Selector */}
          <HStack gap={2}>
            <Text fontSize="sm" color="fg.muted">
              Resolution:
            </Text>
            <Group attached>
              {RESOLUTION_OPTIONS.map((opt) => (
                <Button
                  key={opt.value}
                  size="sm"
                  variant={resolution === opt.value ? 'solid' : 'outline'}
                  onClick={() => setResolution(opt.value)}
                >
                  {opt.label}
                </Button>
              ))}
            </Group>
          </HStack>

          {/* Time Range Picker */}
          <TimeRangePicker value={timeRange} onChange={setTimeRange} />
        </HStack>
      </HStack>

      {/* Content Area - conditionally render based on state */}
      {isError ? (
        <Alert.Root status="error" w="100%">
          <Alert.Indicator />
          <Alert.Description>
            Failed to load GPU timeseries data:{' '}
            {error?.message || 'Unknown error'}
          </Alert.Description>
        </Alert.Root>
      ) : isLoading && availableGpus.length === 0 ? (
        <Box
          w="100%"
          h="400px"
          display="flex"
          alignItems="center"
          justifyContent="center"
        >
          <Spinner size="xl" />
        </Box>
      ) : availableGpus.length === 0 ? (
        <Alert.Root status="info" w="100%">
          <Alert.Indicator />
          <Alert.Description>
            No GPU timeseries data available for this job in the selected time
            range. Try adjusting the time window or resolution.
          </Alert.Description>
        </Alert.Root>
      ) : (
        <>
          {/* Summary Cards for Each GPU */}
          <SimpleGrid
            columns={{
              base: 1,
              md: 2,
              lg: availableGpus.length > 2 ? availableGpus.length : 2,
            }}
            gap={4}
            w="100%"
            opacity={isFetching ? 0.6 : 1}
            transition="opacity 0.3s"
          >
            {availableGpus.map((uuid, idx) => {
              const stats = gpuStats[uuid]
              const card = gpuCardLookup[uuid]
              const pids = gpuAllPids[uuid]
              return (
                <Card.Root
                  size="sm"
                  key={uuid}
                  borderWidth="2px"
                  borderColor={selectedGpu === uuid ? 'blue.500' : 'border'}
                >
                  <Card.Body>
                    <VStack align="start" gap={2}>
                      <HStack justify="space-between" w="100%">
                        <VStack align="start" gap={0}>
                          <Text fontSize="sm" fontWeight="semibold">
                            GPU {idx + 1}
                          </Text>
                          {gpuNodeMapping[uuid] && (
                            <Text fontSize="xs" color="fg.muted">
                              Node: {gpuNodeMapping[uuid]}
                            </Text>
                          )}
                        </VStack>
                        <Badge colorPalette="blue" size="sm">
                          {uuid.slice(-8)}
                        </Badge>
                      </HStack>

                      {/* Hardware info */}
                      {card && (
                        <Box w="100%" pb={1} borderBottomWidth="1px" borderColor="border">
                          <Text fontSize="xs" fontWeight="semibold">
                            {card.manufacturer} {card.model}
                          </Text>
                          <HStack gap={3} flexWrap="wrap">
                            <Text fontSize="xs" color="fg.muted">
                              {card.architecture}
                            </Text>
                            <Text fontSize="xs" color="fg.muted">
                              {(card.memory / (1024 * 1024)).toFixed(0)} GB
                            </Text>
                            <Text fontSize="xs" color="fg.muted">
                              CE {card.max_ce_clock} MHz
                            </Text>
                            <Text fontSize="xs" color="fg.muted">
                              Mem {card.max_memory_clock} MHz
                            </Text>
                          </HStack>
                        </Box>
                      )}

                      <SimpleGrid columns={2} gap={3} w="100%">
                        <Box>
                          <Text fontSize="xs" color="fg.muted">
                            Avg Util
                          </Text>
                          <Text fontSize="lg" fontWeight="bold">
                            {stats.avgGpuUtil.toFixed(1)}%
                          </Text>
                        </Box>
                        <Box>
                          <Text fontSize="xs" color="fg.muted">
                            Peak Util
                          </Text>
                          <Text fontSize="lg" fontWeight="bold">
                            {stats.maxGpuUtil.toFixed(1)}%
                          </Text>
                        </Box>
                        <Box>
                          <Text fontSize="xs" color="fg.muted">
                            Avg Memory
                          </Text>
                          <Text fontSize="lg" fontWeight="bold">
                            {formatMemory(stats.avgGpuMemory)}
                          </Text>
                        </Box>
                        <Box>
                          <Text fontSize="xs" color="fg.muted">
                            Peak Memory
                          </Text>
                          <Text fontSize="lg" fontWeight="bold">
                            {formatMemory(stats.maxGpuMemory)}
                          </Text>
                        </Box>
                      </SimpleGrid>

                      {/* PIDs */}
                      {pids && pids.length > 0 && (
                        <Box w="100%" pt={1} borderTopWidth="1px" borderColor="border">
                          <Text fontSize="xs" color="fg.muted" mb={1}>
                            PIDs ({pids.length}):
                          </Text>
                          <HStack gap={1} flexWrap="wrap">
                            {pids.slice(0, 6).map((pid) => (
                              <Badge key={pid} size="xs" variant="outline" colorPalette="gray">
                                {pid}
                              </Badge>
                            ))}
                            {pids.length > 6 && (
                              <Text fontSize="xs" color="fg.muted">
                                +{pids.length - 6} more
                              </Text>
                            )}
                          </HStack>
                        </Box>
                      )}
                    </VStack>
                  </Card.Body>
                </Card.Root>
              )
            })}
          </SimpleGrid>

          {/* GPU Compute Efficiency */}
          {gpuEfficiency !== null && (
            <Card.Root size="sm" w="100%">
              <Card.Body>
                <HStack justify="space-between" mb={1}>
                  <Text fontSize="md" fontWeight="medium">
                    GPU Compute Efficiency
                  </Text>
                  <HStack gap={2}>
                    <Text fontSize="md" color="fg.muted">
                      {formatEfficiency(gpuEfficiency)}
                    </Text>
                    <Text fontSize="sm" color="fg.muted">
                      ({getEfficiencyLabel(gpuEfficiency)})
                    </Text>
                  </HStack>
                </HStack>
                <Progress.Root
                  value={gpuEfficiency}
                  max={100}
                  colorPalette={getEfficiencyColor(gpuEfficiency)}
                >
                  <Progress.Track>
                    <Progress.Range />
                  </Progress.Track>
                </Progress.Root>
                <Text fontSize="xs" color="fg.muted" mt={1}>
                  Average GPU compute utilization across all {Object.keys(gpuStats).length} GPU(s)
                </Text>
              </Card.Body>
            </Card.Root>
          )}

          {/* Charts Grid */}
          <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
            {/* GPU Compute Utilization Chart */}
            <Card.Root
              size="sm"
              w="100%"
              opacity={isFetching ? 0.6 : 1}
              transition="opacity 0.3s"
            >
              <Card.Body>
                <Text fontSize="md" fontWeight="semibold" mb={4}>
                  GPU Compute Utilization
                </Text>
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart
                    data={combinedUtilData}
                    margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="timeStr"
                      tick={{ fontSize: 12 }}
                      interval="preserveStartEnd"
                    />
                    <YAxis
                      label={{
                        value: 'GPU Utilization %',
                        angle: -90,
                        position: 'insideLeft',
                      }}
                      domain={[0, 100]}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <Legend />

                    {selectedGpu === 'all' ? (
                      availableGpus.map((uuid, idx) => (
                        <Line
                          key={uuid}
                          type="monotone"
                          dataKey={`gpu_util_${uuid.slice(-8)}`}
                          stroke={GPU_COLORS[idx % GPU_COLORS.length]}
                          name={`GPU ${idx + 1}`}
                          dot={false}
                          strokeWidth={2}
                        />
                      ))
                    ) : (
                      <Line
                        type="monotone"
                        dataKey="gpu_util"
                        stroke="#3182CE"
                        name="GPU Utilization"
                        dot={false}
                        strokeWidth={2}
                      />
                    )}

                    <Brush dataKey="timeStr" height={30} stroke="#718096" />
                  </LineChart>
                </ResponsiveContainer>
              </Card.Body>
            </Card.Root>

            {/* GPU Memory Usage Chart (single GPU) */}
            {selectedGpu !== 'all' && (
              <Card.Root
                size="sm"
                w="100%"
                opacity={isFetching ? 0.6 : 1}
                transition="opacity 0.3s"
              >
                <Card.Body>
                  <Text fontSize="md" fontWeight="semibold" mb={4}>
                    GPU Memory Usage
                  </Text>
                  <ResponsiveContainer width="100%" height={300}>
                    <AreaChart
                      data={gpuData[selectedGpu]}
                      margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timeStr"
                        tick={{ fontSize: 12 }}
                        interval="preserveStartEnd"
                      />
                      <YAxis
                        label={{
                          value: 'Memory',
                          angle: -90,
                          position: 'insideLeft',
                        }}
                        tickFormatter={(value) => formatMemory(value)}
                      />
                      <Tooltip
                        content={<CustomTooltip />}
                        formatter={(value: number | undefined) => [
                          formatMemory(value || 0),
                          '',
                        ]}
                      />
                      <Legend />
                      <Area
                        type="monotone"
                        dataKey="gpu_memory"
                        stroke="#38A169"
                        fill="#38A169"
                        fillOpacity={0.6}
                        name="GPU Memory"
                      />
                      <Brush dataKey="timeStr" height={30} stroke="#718096" />
                    </AreaChart>
                  </ResponsiveContainer>
                </Card.Body>
              </Card.Root>
            )}

            {/* GPU Memory Utilization % Chart (single GPU) */}
            {selectedGpu !== 'all' && (
              <Card.Root
                size="sm"
                w="100%"
                opacity={isFetching ? 0.6 : 1}
                transition="opacity 0.3s"
              >
                <Card.Body>
                  <Text fontSize="md" fontWeight="semibold" mb={4}>
                    GPU Memory Utilization %
                  </Text>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart
                      data={gpuData[selectedGpu]}
                      margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timeStr"
                        tick={{ fontSize: 12 }}
                        interval="preserveStartEnd"
                      />
                      <YAxis
                        label={{
                          value: 'Memory %',
                          angle: -90,
                          position: 'insideLeft',
                        }}
                        domain={[0, 100]}
                      />
                      <Tooltip content={<CustomTooltip />} />
                      <Legend />
                      <Line
                        type="monotone"
                        dataKey="gpu_memory_util"
                        stroke="#805AD5"
                        name="GPU Memory Utilization"
                        dot={false}
                        strokeWidth={2}
                      />
                      <Brush dataKey="timeStr" height={30} stroke="#718096" />
                    </LineChart>
                  </ResponsiveContainer>
                </Card.Body>
              </Card.Root>
            )}

            {/* Combined Memory Chart for All GPUs */}
            {selectedGpu === 'all' && (
              <Card.Root
                size="sm"
                w="100%"
                opacity={isFetching ? 0.6 : 1}
                transition="opacity 0.3s"
              >
                <Card.Body>
                  <Text fontSize="md" fontWeight="semibold" mb={4}>
                    GPU Memory Usage (All GPUs)
                  </Text>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart
                      data={(() => {
                        const timeMap = new Map<
                          string,
                          Record<string, string | Date | number>
                        >()
                        Object.entries(gpuData).forEach(([uuid, data]) => {
                          data.forEach((point) => {
                            const key = point.timeStr
                            if (!timeMap.has(key)) {
                              timeMap.set(key, {
                                timeStr: key,
                                time: point.time,
                              })
                            }
                            timeMap.get(key)![`gpu_memory_${uuid.slice(-8)}`] =
                              point.gpu_memory
                          })
                        })
                        return Array.from(timeMap.values()).sort(
                          (a, b) =>
                            (a.time as Date).getTime() -
                            (b.time as Date).getTime(),
                        )
                      })()}
                      margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timeStr"
                        tick={{ fontSize: 12 }}
                        interval="preserveStartEnd"
                      />
                      <YAxis
                        label={{
                          value: 'Memory',
                          angle: -90,
                          position: 'insideLeft',
                        }}
                        tickFormatter={(value) => formatMemory(value)}
                      />
                      <Tooltip
                        content={<CustomTooltip />}
                        formatter={(value: number | undefined) => [
                          formatMemory(value || 0),
                          '',
                        ]}
                      />
                      <Legend />

                      {availableGpus.map((uuid, idx) => (
                        <Line
                          key={uuid}
                          type="monotone"
                          dataKey={`gpu_memory_${uuid.slice(-8)}`}
                          stroke={GPU_COLORS[idx % GPU_COLORS.length]}
                          name={`GPU ${idx + 1}`}
                          dot={false}
                          strokeWidth={2}
                        />
                      ))}

                      <Brush dataKey="timeStr" height={30} stroke="#718096" />
                    </LineChart>
                  </ResponsiveContainer>
                </Card.Body>
              </Card.Root>
            )}

            {/* Combined Memory Utilization % Chart for All GPUs */}
            {selectedGpu === 'all' && (
              <Card.Root
                size="sm"
                w="100%"
                opacity={isFetching ? 0.6 : 1}
                transition="opacity 0.3s"
              >
                <Card.Body>
                  <Text fontSize="md" fontWeight="semibold" mb={4}>
                    GPU Memory Utilization % (All GPUs)
                  </Text>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart
                      data={(() => {
                        const timeMap = new Map<
                          string,
                          Record<string, string | Date | number>
                        >()
                        Object.entries(gpuData).forEach(([uuid, data]) => {
                          data.forEach((point) => {
                            const key = point.timeStr
                            if (!timeMap.has(key)) {
                              timeMap.set(key, {
                                timeStr: key,
                                time: point.time,
                              })
                            }
                            timeMap.get(key)![`gpu_mem_util_${uuid.slice(-8)}`] =
                              point.gpu_memory_util
                          })
                        })
                        return Array.from(timeMap.values()).sort(
                          (a, b) =>
                            (a.time as Date).getTime() -
                            (b.time as Date).getTime(),
                        )
                      })()}
                      margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timeStr"
                        tick={{ fontSize: 12 }}
                        interval="preserveStartEnd"
                      />
                      <YAxis
                        label={{
                          value: 'Memory %',
                          angle: -90,
                          position: 'insideLeft',
                        }}
                        domain={[0, 100]}
                      />
                      <Tooltip content={<CustomTooltip />} />
                      <Legend />

                      {availableGpus.map((uuid, idx) => (
                        <Line
                          key={uuid}
                          type="monotone"
                          dataKey={`gpu_mem_util_${uuid.slice(-8)}`}
                          stroke={GPU_COLORS[idx % GPU_COLORS.length]}
                          name={`GPU ${idx + 1}`}
                          dot={false}
                          strokeWidth={2}
                        />
                      ))}

                      <Brush dataKey="timeStr" height={30} stroke="#718096" />
                    </LineChart>
                  </ResponsiveContainer>
                </Card.Body>
              </Card.Root>
            )}
          </SimpleGrid>
        </>
      )}
    </VStack>
  )
})

GpuPerformanceTab.displayName = 'GpuPerformanceTab'
