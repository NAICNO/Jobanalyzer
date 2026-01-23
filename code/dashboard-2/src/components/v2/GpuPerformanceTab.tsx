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

import { TimeRangePicker, type TimeRange } from '../TimeRangePicker'
import { useJobGpuTimeseries } from '../../hooks/useJobTimeseries'
import {
  transformGpuTimeseries,
  calculateGpuStats,
} from '../../util/timeseriesTransformers'
import { formatMemory } from '../../util/formatters'
import { timeRangeToTimestamps } from '../../util/timeRangeUtils'

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

  const gpuStats = useMemo(() => {
    const statsMap: Record<string, ReturnType<typeof calculateGpuStats>> = {}
    Object.entries(gpuData).forEach(([uuid, data]) => {
      statsMap[uuid] = calculateGpuStats(data)
    })
    return statsMap
  }, [gpuData])

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
                        <Text fontSize="sm" fontWeight="semibold">
                          GPU {idx + 1}
                        </Text>
                        <Badge colorPalette="blue" size="sm">
                          {uuid.slice(-8)}
                        </Badge>
                      </HStack>

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
                    </VStack>
                  </Card.Body>
                </Card.Root>
              )
            })}
          </SimpleGrid>

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
              <ResponsiveContainer width="100%" height={350}>
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
                    // Show all GPUs as separate lines
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
                    // Show single GPU
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

          {/* GPU Memory Usage Chart */}
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

          {/* GPU Memory Utilization % Chart */}
          {selectedGpu !== 'all' && (
            <Card.Root              size="sm"              w="100%"
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
                <ResponsiveContainer width="100%" height={350}>
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
        </>
      )}
    </VStack>
  )
})

GpuPerformanceTab.displayName = 'GpuPerformanceTab'
