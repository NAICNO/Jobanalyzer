import { VStack, HStack, Text, SimpleGrid, Box, Tag, Stat, Spinner } from '@chakra-ui/react'
import { useMemo } from 'react'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ICellRendererParams } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'
import {
  PieChart, Pie, Cell, Tooltip as RechartsTooltip, ResponsiveContainer,
  LineChart, Line, XAxis, YAxis, CartesianGrid, Legend, ReferenceLine,
} from 'recharts'

import type { JobsResponse, SampleGpuTimeseriesResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useClusterJobs, useClusterGpuTimeseries } from '../../hooks/v2/useClusterQueries'
import { getJobStateColor } from '../../util/formatters'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'
import { calculateCpuEfficiency, getEfficiencyColor } from '../../util/efficiency'

const CHAKRA_TO_HEX: Record<string, string> = {
  green: '#38A169',
  blue: '#3182CE',
  yellow: '#D69E2E',
  red: '#E53E3E',
  orange: '#DD6B20',
  purple: '#805AD5',
  gray: '#A0AEC0',
}

// Cell renderer for job state
const JobStateCellRenderer = (props: ICellRendererParams) => {
  const state = props.value || ''
  const color = getJobStateColor(state)

  return (
    <div style={{ display: 'flex', alignItems: 'center', height: '100%' }}>
      <Tag.Root size="sm" colorPalette={color}>
        <Tag.Label>{state}</Tag.Label>
      </Tag.Root>
    </div>
  )
}

// Cell renderer for success rate
const SuccessRateCellRenderer = (props: ICellRendererParams) => {
  const rate = props.value || 0
  const color = rate >= 80 ? 'green' : 'orange'

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
      <Tag.Root size="sm" colorPalette={color}>
        <Tag.Label>{rate}%</Tag.Label>
      </Tag.Root>
    </div>
  )
}

const formatDuration = (minutes: number) => {
  if (minutes < 60) return `${Math.round(minutes)}m`
  const hours = Math.floor(minutes / 60)
  const mins = Math.round(minutes % 60)
  return `${hours}h ${mins}m`
}

// Column definitions for Recent Jobs table
const createRecentJobsColDefs = (
  cluster: string
): ColDef[] => [
  {
    field: 'job_id',
    headerName: 'Job ID',
    flex: 1,
    minWidth: 85,
    cellStyle: { cursor: 'pointer', color: '#3182CE' },
    onCellClicked: (params) => {
      if (params.data?.job_id) {
        window.open(`/v2/${cluster}/jobs/${params.data.job_id}`, '_blank')
      }
    },
  },
  {
    field: 'user_name',
    headerName: 'User',
    flex: 1,
    minWidth: 100,
  },
  {
    field: 'job_state',
    headerName: 'State',
    flex: 1,
    minWidth: 120,
    cellRenderer: JobStateCellRenderer,
  },
  {
    headerName: 'Duration',
    flex: 1,
    minWidth: 100,
    valueGetter: (params) => {
      const job = params.data
      if (!job?.start_time || !job?.end_time) return 0
      const duration = (new Date(job.end_time).getTime() - new Date(job.start_time).getTime()) / (1000 * 60)
      return duration
    },
    valueFormatter: (params) => {
      const duration = params.value as number
      return duration > 0 ? formatDuration(duration) : 'N/A'
    },
  },
]

// Column definitions for Top Users table
const createTopUsersColDefs = (cluster: string): ColDef[] => [
  {
    field: 'user',
    headerName: 'User',
    flex: 2,
    minWidth: 100,
    cellStyle: { cursor: 'pointer', color: '#3182CE' },
    onCellClicked: (params) => {
      if (params.data?.user) {
        window.open(`/v2/${cluster}/jobs/query?user=${encodeURIComponent(params.data.user)}&searched=1`, '_blank')
      }
    },
  },
  {
    field: 'total',
    headerName: 'Total',
    flex: 1,
    minWidth: 80,
    cellStyle: { textAlign: 'center' },
  },
  {
    field: 'successRate',
    headerName: 'Success %',
    flex: 1,
    minWidth: 120,
    cellStyle: { textAlign: 'center' },
    cellRenderer: SuccessRateCellRenderer,
  },
]

interface Props {
  cluster: string
  enabled?: boolean
}

export const ClusterJobAnalytics = ({ cluster, enabled }: Props) => {
  const client = useClusterClient(cluster)
  const { startTimeInS, endTimeInS, timeRange } = useClusterOverviewContext()

  const jobsQ = useClusterJobs({ cluster, client, startTimeInS, endTimeInS, enabled })

  // Adaptive resolution for GPU timeseries (matches CPU efficiency bucketing)
  const rangeSeconds = (endTimeInS ?? 0) - (startTimeInS ?? 0)
  const gpuResolution = rangeSeconds > 7 * 86400 ? 86400 : rangeSeconds > 86400 ? 14400 : 3600
  const gpuTimeseriesQ = useClusterGpuTimeseries({
    cluster, client, startTimeInS, endTimeInS, resolutionInS: gpuResolution, enabled,
  })
  const gpuData = (gpuTimeseriesQ.data ?? {}) as Record<string, Array<SampleGpuTimeseriesResponse>>

  const jobsData = (jobsQ.data as JobsResponse) ?? { jobs: [] }
  const jobs = jobsData.jobs ?? []

  // Calculate statistics
  const completedJobs = jobs.filter(j => j.job_state === 'COMPLETED')
  const failedJobs = jobs.filter(j => j.job_state === 'FAILED')
  const cancelledJobs = jobs.filter(j => j.job_state === 'CANCELLED')
  const timeoutJobs = jobs.filter(j => j.job_state === 'TIMEOUT')
  
  const totalFinished = completedJobs.length + failedJobs.length + cancelledJobs.length + timeoutJobs.length
  const successRate = totalFinished > 0 ? Math.round((completedJobs.length / totalFinished) * 100) : 0

  // Calculate completion times for completed jobs
  const completionTimes: number[] = []
  for (const job of completedJobs) {
    if (job.start_time && job.end_time) {
      const startMs = new Date(job.start_time).getTime()
      const endMs = new Date(job.end_time).getTime()
      const durationMin = (endMs - startMs) / (1000 * 60)
      if (durationMin > 0) {
        completionTimes.push(durationMin)
      }
    }
  }

  const avgCompletionTime = completionTimes.length > 0
    ? Math.round(completionTimes.reduce((a, b) => a + b, 0) / completionTimes.length)
    : 0

  const medianCompletionTime = completionTimes.length > 0
    ? Math.round(completionTimes.sort((a, b) => a - b)[Math.floor(completionTimes.length / 2)])
    : 0

  // Group by user
  const userStats: Record<string, { completed: number; failed: number; total: number }> = {}
  for (const job of jobs) {
    const user = job.user_name ?? 'Unknown'
    if (!userStats[user]) {
      userStats[user] = { completed: 0, failed: 0, total: 0 }
    }
    userStats[user].total++
    if (job.job_state === 'COMPLETED') userStats[user].completed++
    if (job.job_state === 'FAILED') userStats[user].failed++
  }

  // TODO: There is a row which has no username.  Usually it has the highest job count, which is suspicious.  Need to investigate if this is a data issue or a bug in the grouping logic.
  const topUsers = Object.entries(userStats)
    .map(([user, stats]) => ({
      user,
      ...stats,
      successRate: stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0
    }))
    .sort((a, b) => b.total - a.total)
    .slice(0, 10)

  // Recent jobs for table (last 20) - deduplicated by job_id, only with user
  const recentJobs = (() => {
    const jobsWithEndTimeAndUser = jobs.filter(j => j.end_time && j.user_name)

    // Group by job_id and take the most recent entry for each
    const uniqueJobsMap = new Map()
    for (const job of jobsWithEndTimeAndUser) {
      const existingJob = uniqueJobsMap.get(job.job_id)
      if (!existingJob) {
        uniqueJobsMap.set(job.job_id, job)
      } else {
        // Keep the job with the more recent end_time
        const existingTime = existingJob.end_time ? new Date(existingJob.end_time).getTime() : 0
        const currentTime = job.end_time ? new Date(job.end_time).getTime() : 0
        if (currentTime > existingTime) {
          uniqueJobsMap.set(job.job_id, job)
        }
      }
    }

    return Array.from(uniqueJobsMap.values())
      .sort((a, b) => {
        const aTime = a.end_time ? new Date(a.end_time).getTime() : 0
        const bTime = b.end_time ? new Date(b.end_time).getTime() : 0
        return bTime - aTime
      })
      .slice(0, 20)
  })()

  // Use column definitions from outside the component
  const recentJobsColDefs = useMemo<ColDef[]>(
    () => createRecentJobsColDefs(cluster),
    [cluster]
  )

  const topUsersColDefs = useMemo<ColDef[]>(
    () => createTopUsersColDefs(cluster),
    [cluster]
  )

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Job Analytics ({timeRange.label})</Text>

      {/* Summary Cards */}
      <SimpleGrid columns={{ base: 2, md: 4, lg: 6 }} gap={3} w="100%">
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Total Jobs</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {jobs.length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Completed</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color="green.600">
              {completedJobs.length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Failed</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color="red.600">
              {failedJobs.length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Success Rate</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold" color={successRate >= 80 ? 'green.600' : 'orange.600'}>
              {successRate}%
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Avg Duration</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {avgCompletionTime > 0 ? formatDuration(avgCompletionTime) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Median Duration</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {medianCompletionTime > 0 ? formatDuration(medianCompletionTime) : 'N/A'}
            </Stat.ValueText>
          </Stat.Root>
        </Box>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 3 }} gap={4} w="100%">
        {/* Job State Distribution */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%" h="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Job State Distribution</Text>
            {jobsQ.isLoading ? (
              <Box w="100%" flex={1} display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            ) : jobs.length === 0 ? (
              <Text fontSize="xs" color="gray.500">No job data available</Text>
            ) : (
              (() => {
                const stateDistribution: Record<string, number> = {}
                for (const job of jobs) {
                  const raw = job.job_state ?? 'UNKNOWN'
                  const state = raw.split(/[\s+]/)[0]
                  stateDistribution[state] = (stateDistribution[state] ?? 0) + 1
                }
                const pieData = Object.entries(stateDistribution)
                  .map(([state, count]) => ({
                    name: state,
                    value: count,
                    color: getJobStateColor(state),
                  }))
                  .sort((a, b) => b.value - a.value)

                return (
                  <HStack gap={1} w="100%" flex={1} align="center" justify="center">
                    <Box flex={1} minW={0}>
                      <ResponsiveContainer width="100%" height={300} minWidth={0}>
                        <PieChart>
                          <Pie
                            data={pieData}
                            cx="50%"
                            cy="50%"
                            outerRadius="80%"
                            dataKey="value"
                            nameKey="name"
                            label={(props) => {
                              const { cx, cy, midAngle, innerRadius, outerRadius, percent } = props as {
                                cx: number; cy: number; midAngle: number; innerRadius: number; outerRadius: number; percent: number
                              }
                              if (!percent || percent < 0.05) return null
                              const RADIAN = Math.PI / 180
                              const radius = innerRadius + (outerRadius - innerRadius) * 0.5
                              const x = cx + radius * Math.cos(-midAngle * RADIAN)
                              const y = cy + radius * Math.sin(-midAngle * RADIAN)
                              return (
                                <text x={x} y={y} fill="white" textAnchor="middle" dominantBaseline="central" fontSize={13} fontWeight="bold">
                                  {`${(percent * 100).toFixed(0)}%`}
                                </text>
                              )
                            }}
                            labelLine={false}
                          >
                            {pieData.map((entry, idx) => (
                              <Cell key={idx} fill={CHAKRA_TO_HEX[entry.color] ?? '#A0AEC0'} />
                            ))}
                          </Pie>
                          <RechartsTooltip contentStyle={{ fontSize: 12 }} />
                        </PieChart>
                      </ResponsiveContainer>
                    </Box>
                    <VStack align="start" gap={1} flexShrink={0}>
                      {pieData.map((d) => (
                        <Tag.Root key={d.name} size="sm" colorPalette={d.color}>
                          <Tag.Label>{d.name}: {d.value}</Tag.Label>
                        </Tag.Root>
                      ))}
                    </VStack>
                  </HStack>
                )
              })()
            )}
          </VStack>
        </Box>

        {/* Recent Jobs Table */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Recent Completed Jobs (Top 20)</Text>
            {jobsQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            ) : recentJobs.length === 0 ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="xs" color="gray.500">No completed jobs</Text>
              </Box>
            ) : (
              <Box w="100%" h="300px">
                <AgGridReact
                  theme={themeQuartz}
                  rowData={recentJobs}
                  columnDefs={recentJobsColDefs}
                  domLayout="normal"
                  rowHeight={32}
                  headerHeight={36}
                />
              </Box>
            )}
          </VStack>
        </Box>

        {/* User Statistics */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Top Users by Job Count (Top 10)</Text>
            {topUsers.length === 0 ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="xs" color="gray.500">No user data</Text>
              </Box>
            ) : (
              <Box w="100%" h="300px">
                <AgGridReact
                  theme={themeQuartz}
                  rowData={topUsers}
                  columnDefs={topUsersColDefs}
                  domLayout="normal"
                  rowHeight={32}
                  headerHeight={36}
                />
              </Box>
            )}
          </VStack>
        </Box>
      </SimpleGrid>

      {/* CPU Efficiency Trends */}
      {(() => {
        const jobsWithEfficiency = jobs
          .filter((j) => j.job_state === 'COMPLETED' && j.sacct && j.start_time && j.end_time)
          .map((j) => {
            const elapsed = j.sacct!.ElapsedRaw ?? 0
            const cpuEff = calculateCpuEfficiency(j, elapsed)
            return {
              endTime: new Date(j.end_time!).getTime(),
              cpuEfficiency: cpuEff,
            }
          })
          .filter((j): j is { endTime: number; cpuEfficiency: number } => j.cpuEfficiency !== null)
          .sort((a, b) => a.endTime - b.endTime)

        if (jobsWithEfficiency.length === 0) return null

        // Adaptive bucketing based on time range
        const rangeSeconds = (endTimeInS ?? 0) - (startTimeInS ?? 0)
        const bucketMs = rangeSeconds > 7 * 86400
          ? 24 * 60 * 60 * 1000
          : rangeSeconds > 86400
            ? 4 * 60 * 60 * 1000
            : 60 * 60 * 1000

        const buckets = new Map<number, { sum: number; count: number; lowCount: number }>()
        for (const job of jobsWithEfficiency) {
          const key = Math.floor(job.endTime / bucketMs) * bucketMs
          if (!buckets.has(key)) buckets.set(key, { sum: 0, count: 0, lowCount: 0 })
          const b = buckets.get(key)!
          b.sum += job.cpuEfficiency
          b.count++
          if (job.cpuEfficiency < 20) b.lowCount++
        }

        const trendData = Array.from(buckets.entries())
          .map(([ts, b]) => ({
            time: new Date(ts).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit' }),
            timestamp: ts,
            avgEfficiency: Math.round(b.sum / b.count),
            jobCount: b.count,
          }))
          .sort((a, b) => a.timestamp - b.timestamp)

        const overallAvgEff = Math.round(
          jobsWithEfficiency.reduce((s, j) => s + j.cpuEfficiency, 0) / jobsWithEfficiency.length
        )
        const lowEffCount = jobsWithEfficiency.filter((j) => j.cpuEfficiency < 20).length

        return (
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">CPU Efficiency Trends</Text>

            <HStack align="start" gap={4} w="100%" h="280px">
              <VStack gap={3} flexShrink={0} w="200px" h="100%">
                <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%" flex={1}>
                  <Stat.Root>
                    <Stat.Label fontSize="sm" color="gray.600">Avg CPU Efficiency</Stat.Label>
                    <Stat.ValueText fontSize="2xl" fontWeight="bold" color={`${getEfficiencyColor(overallAvgEff)}.600`}>
                      {overallAvgEff}%
                    </Stat.ValueText>
                  </Stat.Root>
                </Box>
                <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%" flex={1}>
                  <Stat.Root>
                    <Stat.Label fontSize="sm" color="gray.600">Low Efficiency Jobs (&lt;20%)</Stat.Label>
                    <Stat.ValueText fontSize="2xl" fontWeight="bold" color={lowEffCount > 0 ? 'red.600' : 'green.600'}>
                      {lowEffCount}
                    </Stat.ValueText>
                  </Stat.Root>
                </Box>
                <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%" flex={1}>
                  <Stat.Root>
                    <Stat.Label fontSize="sm" color="gray.600">Jobs Analyzed</Stat.Label>
                    <Stat.ValueText fontSize="2xl" fontWeight="bold">
                      {jobsWithEfficiency.length}
                    </Stat.ValueText>
                  </Stat.Root>
                </Box>
              </VStack>

              <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" flex={1} minW={0} h="100%">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={trendData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" tick={{ fontSize: 10 }} />
                    <YAxis
                      domain={[0, 100]}
                      tick={{ fontSize: 10 }}
                      label={{ value: 'Efficiency %', angle: -90, position: 'insideLeft', style: { fontSize: 10 } }}
                    />
                    <RechartsTooltip contentStyle={{ fontSize: 12 }} />
                    <Legend wrapperStyle={{ fontSize: 12 }} />
                    <ReferenceLine y={20} stroke="#E53E3E" strokeDasharray="5 5" label={{ value: 'Low', fontSize: 10 }} />
                    <ReferenceLine y={80} stroke="#38A169" strokeDasharray="5 5" label={{ value: 'Target', fontSize: 10 }} />
                    <Line
                      type="monotone"
                      dataKey="avgEfficiency"
                      stroke="#3182CE"
                      name="Avg CPU Efficiency"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </HStack>
          </VStack>
        )
      })()}

      {/* Cluster-wide GPU Utilization Trends */}
      {(() => {
        // Aggregate GPU utilization by adaptive time buckets
        const bucketMs = rangeSeconds > 7 * 86400
          ? 24 * 60 * 60 * 1000
          : rangeSeconds > 86400
            ? 4 * 60 * 60 * 1000
            : 60 * 60 * 1000

        const gpuBuckets = new Map<number, { totalUtil: number; maxUtil: number; count: number; activeCount: number }>()

        for (const gpuArrays of Object.values(gpuData)) {
          if (Array.isArray(gpuArrays)) {
            for (const gpuTimeseries of gpuArrays) {
              if (Array.isArray(gpuTimeseries.data)) {
                for (const sample of gpuTimeseries.data) {
                  const ts = new Date(sample.time).getTime()
                  const key = Math.floor(ts / bucketMs) * bucketMs
                  const util = sample.ce_util ?? 0

                  if (!gpuBuckets.has(key)) {
                    gpuBuckets.set(key, { totalUtil: 0, maxUtil: 0, count: 0, activeCount: 0 })
                  }
                  const b = gpuBuckets.get(key)!
                  b.totalUtil += util
                  b.maxUtil = Math.max(b.maxUtil, util)
                  b.count++
                  if (util > 0) b.activeCount++
                }
              }
            }
          }
        }

        if (gpuBuckets.size === 0) return null

        const gpuTrendData = Array.from(gpuBuckets.entries())
          .map(([ts, b]) => ({
            time: new Date(ts).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit' }),
            timestamp: ts,
            avgUtil: Math.round(b.totalUtil / b.count),
            peakUtil: Math.round(b.maxUtil),
            gpuCount: b.count,
          }))
          .sort((a, b) => a.timestamp - b.timestamp)

        const totalSamples = Array.from(gpuBuckets.values()).reduce((s, b) => s + b.count, 0)
        const totalUtil = Array.from(gpuBuckets.values()).reduce((s, b) => s + b.totalUtil, 0)
        const overallAvgUtil = totalSamples > 0 ? Math.round(totalUtil / totalSamples) : 0
        const peakUtil = Math.max(...gpuTrendData.map(d => d.peakUtil))
        // Estimate unique GPUs from the max count in any single bucket
        const gpusAnalyzed = Math.max(...Array.from(gpuBuckets.values()).map(b => b.count))

        return (
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Cluster-wide GPU Utilization Trends</Text>

            <HStack align="start" gap={4} w="100%" h="280px">
              <VStack gap={3} flexShrink={0} w="200px" h="100%">
                <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%" flex={1}>
                  <Stat.Root>
                    <Stat.Label fontSize="sm" color="gray.600">Avg GPU Utilization</Stat.Label>
                    <Stat.ValueText fontSize="2xl" fontWeight="bold" color={overallAvgUtil >= 80 ? 'green.600' : overallAvgUtil >= 50 ? 'orange.600' : 'red.600'}>
                      {overallAvgUtil}%
                    </Stat.ValueText>
                  </Stat.Root>
                </Box>
                <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%" flex={1}>
                  <Stat.Root>
                    <Stat.Label fontSize="sm" color="gray.600">Peak GPU Utilization</Stat.Label>
                    <Stat.ValueText fontSize="2xl" fontWeight="bold">
                      {peakUtil}%
                    </Stat.ValueText>
                  </Stat.Root>
                </Box>
                <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" w="100%" flex={1}>
                  <Stat.Root>
                    <Stat.Label fontSize="sm" color="gray.600">GPUs Analyzed</Stat.Label>
                    <Stat.ValueText fontSize="2xl" fontWeight="bold">
                      {gpusAnalyzed}
                    </Stat.ValueText>
                  </Stat.Root>
                </Box>
              </VStack>

              <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" flex={1} minW={0} h="100%">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={gpuTrendData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" tick={{ fontSize: 10 }} />
                    <YAxis
                      domain={[0, 100]}
                      tick={{ fontSize: 10 }}
                      label={{ value: 'GPU Util %', angle: -90, position: 'insideLeft', style: { fontSize: 10 } }}
                    />
                    <RechartsTooltip contentStyle={{ fontSize: 12 }} />
                    <Legend wrapperStyle={{ fontSize: 12 }} />
                    <ReferenceLine y={80} stroke="#38A169" strokeDasharray="5 5" label={{ value: 'Target', fontSize: 10 }} />
                    <Line
                      type="monotone"
                      dataKey="avgUtil"
                      stroke="#805AD5"
                      name="Avg GPU Utilization"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                    />
                    <Line
                      type="monotone"
                      dataKey="peakUtil"
                      stroke="#D69E2E"
                      name="Peak GPU Utilization"
                      strokeWidth={1}
                      strokeDasharray="4 4"
                      dot={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </HStack>
          </VStack>
        )
      })()}
    </VStack>
  )
}
