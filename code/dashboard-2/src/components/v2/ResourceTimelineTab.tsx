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
  Stat,
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
import { useJobProcessTimeseries } from '../../hooks/useJobTimeseries'
import {
  transformProcessTimeseries,
  calculateTimeseriesStats,
} from '../../util/timeseriesTransformers'
import { formatMemory } from '../../util/formatters'
import { timeRangeToTimestamps } from '../../util/timeRangeUtils'

interface Props {
  cluster: string;
  jobId: number;
  client: Client | null;
  jobStartTime?: Date | string | null;
  jobEndTime?: Date | string | null;
  elapsed?: number;
}

const RESOLUTION_OPTIONS = [
  { label: '1s', value: 1 },
  { label: '10s', value: 10 },
  { label: '30s', value: 30 },
  { label: '1m', value: 60 },
  { label: '5m', value: 300 },
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
 * Custom tooltip for timeline charts
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

const computeSmartDefault = (
  elapsed?: number,
  jobStartTime?: Date | string | null,
  jobEndTime?: Date | string | null,
): { timeRange: TimeRange; resolution: number } => {
  if (elapsed && elapsed > 86400 && jobStartTime) {
    const start = new Date(jobStartTime)
    const end = jobEndTime ? new Date(jobEndTime) : new Date()
    return {
      timeRange: { type: 'absolute', label: 'Full job', value: 'full-job', start, end, endAt: jobEndTime ? 'fixed' : 'now' },
      resolution: 300,
    }
  }
  if (elapsed && elapsed > 21600) return { timeRange: { type: 'relative', label: 'Last 1 day', value: '1d', endAt: 'now' }, resolution: 60 }
  if (elapsed && elapsed > 3600) return { timeRange: { type: 'relative', label: 'Last 3 hours', value: '3h', endAt: 'now' }, resolution: 30 }
  return { timeRange: { type: 'relative', label: 'Last 1 hour', value: '1h', endAt: 'now' }, resolution: 30 }
}

export const ResourceTimelineTab = memo(({ cluster, jobId, client, jobStartTime, jobEndTime, elapsed }: Props) => {
  const defaults = useMemo(() => computeSmartDefault(elapsed, jobStartTime, jobEndTime), [elapsed, jobStartTime, jobEndTime])
  const [timeRange, setTimeRange] = useState<TimeRange>(defaults.timeRange)
  const [resolution, setResolution] = useState(defaults.resolution)

  const handleFullJobDuration = () => {
    if (!jobStartTime) return
    const start = new Date(jobStartTime)
    const end = jobEndTime ? new Date(jobEndTime) : new Date()
    const durationSec = (end.getTime() - start.getTime()) / 1000
    setTimeRange({ type: 'absolute', label: 'Full job', value: 'full-job', start, end, endAt: jobEndTime ? 'fixed' : 'now' })
    if (durationSec > 86400) setResolution(300)
    else if (durationSec > 21600) setResolution(60)
    else if (durationSec > 3600) setResolution(30)
    else setResolution(10)
  }

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
  } = useJobProcessTimeseries({
    cluster,
    jobId,
    client,
    resolution,
    startTimeInS,
    endTimeInS,
  })

  const chartData = useMemo(
    () => transformProcessTimeseries(timeseriesData?.[0]),
    [timeseriesData],
  )

  const stats = useMemo(() => calculateTimeseriesStats(chartData), [chartData])

  return (
    <VStack w="100%" gap={6} align="start">
      {/* Controls */}
      <HStack justify="space-between" w="100%" flexWrap="wrap" gap={4}>
        <HStack gap={2}>
          <Text fontSize="lg" fontWeight="semibold">
            Resource Usage Over Time
          </Text>
          {isFetching && <Spinner size="sm" color="blue.500" />}
        </HStack>
        <HStack gap={4} flexWrap="wrap">
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
          {jobStartTime && (
            <Button size="sm" variant={timeRange.value === 'full-job' ? 'solid' : 'outline'} onClick={handleFullJobDuration}>
              Full Job
            </Button>
          )}
          <TimeRangePicker value={timeRange} onChange={setTimeRange} />
        </HStack>
      </HStack>

      {/* Content Area - conditionally render based on state */}
      {isError ? (
        <Alert.Root status="error" w="100%">
          <Alert.Indicator />
          <Alert.Description>
            Failed to load timeseries data: {error?.message || 'Unknown error'}
          </Alert.Description>
        </Alert.Root>
      ) : isLoading && !chartData.length ? (
        <Box
          w="100%"
          h="400px"
          display="flex"
          alignItems="center"
          justifyContent="center"
        >
          <Spinner size="xl" />
        </Box>
      ) : chartData.length === 0 ? (
        <Alert.Root status="info" w="100%">
          <Alert.Indicator />
          <Alert.Description>
            No timeseries data available for this job in the selected time
            range. Try adjusting the time window or resolution.
          </Alert.Description>
        </Alert.Root>
      ) : (
        <>
          {/* Summary Stats */}
          <SimpleGrid
            columns={{ base: 2, md: 4 }}
            gap={4}
            w="100%"
            opacity={isFetching ? 0.6 : 1}
            transition="opacity 0.3s"
          >
            <Card.Root size="sm">
              <Card.Body>
                <Stat.Root>
                  <Stat.Label fontSize="sm">Avg CPU Util</Stat.Label>
                  <Stat.ValueText fontSize="2xl" fontWeight="bold">
                    {stats.avgCpuUtil.toFixed(1)}%
                  </Stat.ValueText>
                  <Stat.HelpText fontSize="xs">
                    Peak: {stats.maxCpuUtil.toFixed(1)}%
                  </Stat.HelpText>
                </Stat.Root>
              </Card.Body>
            </Card.Root>

            <Card.Root size="sm">
              <Card.Body>
                <Stat.Root>
                  <Stat.Label fontSize="sm">Avg Memory</Stat.Label>
                  <Stat.ValueText fontSize="2xl" fontWeight="bold">
                    {formatMemory(stats.avgMemoryResident)}
                  </Stat.ValueText>
                  <Stat.HelpText fontSize="xs">
                    Peak: {formatMemory(stats.maxMemoryResident)}
                  </Stat.HelpText>
                </Stat.Root>
              </Card.Body>
            </Card.Root>

            <Card.Root size="sm">
              <Card.Body>
                <Stat.Root>
                  <Stat.Label fontSize="sm">Avg Processes</Stat.Label>
                  <Stat.ValueText fontSize="2xl" fontWeight="bold">
                    {stats.avgProcesses.toFixed(1)}
                  </Stat.ValueText>
                </Stat.Root>
              </Card.Body>
            </Card.Root>

            <Card.Root size="sm">
              <Card.Body>
                <Stat.Root>
                  <Stat.Label fontSize="sm">Data Points</Stat.Label>
                  <Stat.ValueText fontSize="2xl" fontWeight="bold">
                    {chartData.length}
                  </Stat.ValueText>
                </Stat.Root>
              </Card.Body>
            </Card.Root>
          </SimpleGrid>

          {/* Charts Grid - 2x2 layout */}
          <SimpleGrid
            columns={{ base: 1, lg: 2 }}
            gap={4}
            w="100%"
            opacity={isFetching ? 0.6 : 1}
            transition="opacity 0.3s"
          >
            {/* CPU Utilization Chart */}
            <Card.Root size="sm">
              <Card.Body>
                <Text fontSize="md" fontWeight="semibold" mb={4}>
                  CPU Utilization
                </Text>
                <ResponsiveContainer width="100%" height={220}>
                  <LineChart
                    data={chartData}
                    margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="timeStr"
                      tick={{ fontSize: 10 }}
                      interval="preserveStartEnd"
                    />
                    <YAxis
                      label={{
                        value: 'CPU %',
                        angle: -90,
                        position: 'insideLeft',
                        style: { fontSize: 11 },
                      }}
                      domain={[0, 100]}
                      tick={{ fontSize: 10 }}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    <Line
                      type="monotone"
                      dataKey="cpu_util"
                      stroke="#3182CE"
                      name="CPU Utilization"
                      dot={false}
                      strokeWidth={2}
                    />
                    <Line
                      type="monotone"
                      dataKey="cpu_avg"
                      stroke="#E53E3E"
                      name="CPU Average"
                      dot={false}
                      strokeWidth={2}
                    />
                    <Brush dataKey="timeStr" height={25} stroke="#718096" />
                  </LineChart>
                </ResponsiveContainer>
              </Card.Body>
            </Card.Root>

            {/* Memory Usage Chart */}
            <Card.Root size="sm">
              <Card.Body>
                <Text fontSize="md" fontWeight="semibold" mb={4}>
                  Memory Usage
                </Text>
                <ResponsiveContainer width="100%" height={220}>
                  <AreaChart
                    data={chartData}
                    margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="timeStr"
                      tick={{ fontSize: 10 }}
                      interval="preserveStartEnd"
                    />
                    <YAxis
                      label={{
                        value: 'Memory (GB)',
                        angle: -90,
                        position: 'insideLeft',
                        style: { fontSize: 11 },
                      }}
                      tickFormatter={(value) => formatMemory(value)}
                      tick={{ fontSize: 10 }}
                    />
                    <Tooltip
                      content={<CustomTooltip />}
                      formatter={(value: number | undefined) => [
                        formatMemory(value || 0),
                        '',
                      ]}
                    />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    <Area
                      type="monotone"
                      dataKey="memory_resident"
                      stackId="1"
                      stroke="#38A169"
                      fill="#38A169"
                      fillOpacity={0.6}
                      name="Resident Memory"
                    />
                    <Area
                      type="monotone"
                      dataKey="memory_virtual"
                      stackId="2"
                      stroke="#805AD5"
                      fill="#805AD5"
                      fillOpacity={0.4}
                      name="Virtual Memory"
                    />
                    <Brush dataKey="timeStr" height={25} stroke="#718096" />
                  </AreaChart>
                </ResponsiveContainer>
              </Card.Body>
            </Card.Root>

            {/* Memory Utilization % Chart */}
            <Card.Root size="sm">
              <Card.Body>
                <Text fontSize="md" fontWeight="semibold" mb={4}>
                  Memory Utilization %
                </Text>
                <ResponsiveContainer width="100%" height={220}>
                  <LineChart
                    data={chartData}
                    margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="timeStr"
                      tick={{ fontSize: 10 }}
                      interval="preserveStartEnd"
                    />
                    <YAxis
                      label={{
                        value: 'Memory %',
                        angle: -90,
                        position: 'insideLeft',
                        style: { fontSize: 11 },
                      }}
                      domain={[0, 100]}
                      tick={{ fontSize: 10 }}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    <Line
                      type="monotone"
                      dataKey="memory_util"
                      stroke="#319795"
                      name="Memory Utilization"
                      dot={false}
                      strokeWidth={2}
                    />
                    <Brush dataKey="timeStr" height={25} stroke="#718096" />
                  </LineChart>
                </ResponsiveContainer>
              </Card.Body>
            </Card.Root>

            {/* Process Count Chart */}
            <Card.Root size="sm">
              <Card.Body>
                <Text fontSize="md" fontWeight="semibold" mb={4}>
                  Process Count
                </Text>
                <ResponsiveContainer width="100%" height={220}>
                  <LineChart
                    data={chartData}
                    margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="timeStr"
                      tick={{ fontSize: 10 }}
                      interval="preserveStartEnd"
                    />
                    <YAxis
                      label={{
                        value: 'Processes',
                        angle: -90,
                        position: 'insideLeft',
                        style: { fontSize: 11 },
                      }}
                      tick={{ fontSize: 10 }}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    <Line
                      type="monotone"
                      dataKey="processes_avg"
                      stroke="#DD6B20"
                      name="Average Processes"
                      dot={false}
                      strokeWidth={2}
                    />
                    <Brush dataKey="timeStr" height={25} stroke="#718096" />
                  </LineChart>
                </ResponsiveContainer>
              </Card.Body>
            </Card.Root>
          </SimpleGrid>
        </>
      )}
    </VStack>
  )
})

ResourceTimelineTab.displayName = 'ResourceTimelineTab'
