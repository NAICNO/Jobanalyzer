import { VStack, Text, SimpleGrid, Box, Spinner } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import {
  getClusterByClusterNodesGpuTimeseriesOptions,
  getClusterByClusterNodesCpuTimeseriesOptions,
  getClusterByClusterNodesMemoryTimeseriesOptions
} from '../../client/@tanstack/react-query.gen'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import type { SampleGpuTimeseriesResponse, SampleProcessAccResponse } from '../../client'

interface Props {
  cluster: string
}

export const ClusterTimebasedActivity = ({cluster}: Props) => {
  const now = Math.floor(Date.now() / 1000)
  const last24h = now - (24 * 60 * 60)

  const gpuTimeseriesQ = useQuery({
    ...getClusterByClusterNodesGpuTimeseriesOptions({
      path: {cluster},
      query: {
        start_time_in_s: last24h,
        end_time_in_s: now,
        resolution_in_s: 3600, // 1 hour resolution
      },
    }),
    enabled: !!cluster,
  })

  const cpuTimeseriesQ = useQuery({
    ...getClusterByClusterNodesCpuTimeseriesOptions({
      path: {cluster},
      query: {
        start_time_in_s: last24h,
        end_time_in_s: now,
        resolution_in_s: 3600,
      },
    }),
    enabled: !!cluster,
  })

  const memoryTimeseriesQ = useQuery({
    ...getClusterByClusterNodesMemoryTimeseriesOptions({
      path: {cluster},
      query: {
        start_time_in_s: last24h,
        end_time_in_s: now,
        resolution_in_s: 3600,
      },
    }),
    enabled: !!cluster,
  })

  const gpuData = (gpuTimeseriesQ.data ?? {}) as Record<string, Array<SampleGpuTimeseriesResponse>>
  const cpuData = (cpuTimeseriesQ.data ?? {}) as Record<string, Array<SampleProcessAccResponse>>
  const memoryData = (memoryTimeseriesQ.data ?? {}) as Record<string, Array<SampleProcessAccResponse>>

  // Process GPU timeseries data
  const gpuTimeSeriesData: Array<{
    time: string
    avgUtil: number
    maxUtil: number
    gpuCount: number
  }> = []

  // Process CPU timeseries data
  const cpuTimeSeriesData: Array<{
    time: string
    avgUtil: number
    maxUtil: number
  }> = []

  // Process Memory timeseries data
  const memoryTimeSeriesData: Array<{
    time: string
    avgUtil: number
    maxUtil: number
  }> = []

  // Aggregate GPU by timestamp across all nodes
  const gpuTimestampMap = new Map<number, { totalUtil: number; maxUtil: number; count: number }>()

  for (const gpuArrays of Object.values(gpuData)) {
    if (Array.isArray(gpuArrays)) {
      for (const gpuTimeseries of gpuArrays) {
        if (Array.isArray(gpuTimeseries.data)) {
          for (const sample of gpuTimeseries.data) {
            const ts = sample.time.getTime() / 1000
            const util = sample.ce_util ?? 0

            if (!gpuTimestampMap.has(ts)) {
              gpuTimestampMap.set(ts, {totalUtil: 0, maxUtil: 0, count: 0})
            }

            const entry = gpuTimestampMap.get(ts)!
            entry.totalUtil += util
            entry.maxUtil = Math.max(entry.maxUtil, util)
            entry.count++
          }
        }
      }
    }
  }

  // Aggregate CPU by timestamp
  const cpuTimestampMap = new Map<number, { totalUtil: number; maxUtil: number; count: number }>()

  for (const samples of Object.values(cpuData)) {
    if (Array.isArray(samples)) {
      for (const sample of samples) {
        const ts = sample.time.getTime() / 1000
        const util = sample.cpu_util ?? 0

        if (!cpuTimestampMap.has(ts)) {
          cpuTimestampMap.set(ts, {totalUtil: 0, maxUtil: 0, count: 0})
        }

        const entry = cpuTimestampMap.get(ts)!
        entry.totalUtil += util
        entry.maxUtil = Math.max(entry.maxUtil, util)
        entry.count++
      }
    }
  }

  // Aggregate Memory by timestamp
  const memoryTimestampMap = new Map<number, { totalUtil: number; maxUtil: number; count: number }>()

  for (const samples of Object.values(memoryData)) {
    if (Array.isArray(samples)) {
      for (const sample of samples) {
        const ts = sample.time.getTime() / 1000
        const util = sample.memory_util ?? 0

        if (!memoryTimestampMap.has(ts)) {
          memoryTimestampMap.set(ts, {totalUtil: 0, maxUtil: 0, count: 0})
        }

        const entry = memoryTimestampMap.get(ts)!
        entry.totalUtil += util
        entry.maxUtil = Math.max(entry.maxUtil, util)
        entry.count++
      }
    }
  }

  // Convert to chart data
  for (const [ts, data] of Array.from(gpuTimestampMap.entries()).sort((a, b) => a[0] - b[0])) {
    const date = new Date(ts * 1000)
    gpuTimeSeriesData.push({
      time: date.toLocaleTimeString('en-US', {hour: '2-digit', minute: '2-digit'}),
      avgUtil: Math.round(data.totalUtil / data.count),
      maxUtil: Math.round(data.maxUtil),
      gpuCount: data.count,
    })
  }

  for (const [ts, data] of Array.from(cpuTimestampMap.entries()).sort((a, b) => a[0] - b[0])) {
    const date = new Date(ts * 1000)
    cpuTimeSeriesData.push({
      time: date.toLocaleTimeString('en-US', {hour: '2-digit', minute: '2-digit'}),
      avgUtil: Math.round(data.totalUtil / data.count),
      maxUtil: Math.round(data.maxUtil),
    })
  }

  for (const [ts, data] of Array.from(memoryTimestampMap.entries()).sort((a, b) => a[0] - b[0])) {
    const date = new Date(ts * 1000)
    memoryTimeSeriesData.push({
      time: date.toLocaleTimeString('en-US', {hour: '2-digit', minute: '2-digit'}),
      avgUtil: Math.round(data.totalUtil / data.count),
      maxUtil: Math.round(data.maxUtil),
    })
  }

  const hasGpuData = gpuTimeSeriesData.length > 0
  const hasCpuData = cpuTimeSeriesData.length > 0
  const hasMemoryData = memoryTimeSeriesData.length > 0

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Resource Activity Over Time (Last 24 Hours)</Text>

      <SimpleGrid columns={{base: 1, lg: 2}} gap={4} w="100%">
        {/* GPU Utilization Over Time */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">GPU Utilization Trend</Text>
            {gpuTimeseriesQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg"/>
              </Box>
            ) : !hasGpuData ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="sm" color="gray.500">No GPU data available</Text>
              </Box>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={gpuTimeSeriesData}>
                  <CartesianGrid strokeDasharray="3 3"/>
                  <XAxis
                    dataKey="time"
                    tick={{fontSize: 10}}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    domain={[0, 100]}
                    tick={{fontSize: 10}}
                    label={{value: 'GPU Util %', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}
                  />
                  <Tooltip
                    contentStyle={{fontSize: 12}}
                    formatter={(value: number) => `${value}%`}
                  />
                  <Legend wrapperStyle={{fontSize: 12}}/>
                  <Line
                    type="monotone"
                    dataKey="avgUtil"
                    stroke="#3182ce"
                    name="Avg Utilization"
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="maxUtil"
                    stroke="#e53e3e"
                    name="Max Utilization"
                    strokeWidth={1}
                    strokeDasharray="5 5"
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            )}
          </VStack>
        </Box>

        {/* CPU Utilization Over Time */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">CPU Utilization Trend</Text>
            {cpuTimeseriesQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg"/>
              </Box>
            ) : !hasCpuData ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="sm" color="gray.500">No CPU data available</Text>
              </Box>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={cpuTimeSeriesData}>
                  <CartesianGrid strokeDasharray="3 3"/>
                  <XAxis
                    dataKey="time"
                    tick={{fontSize: 10}}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    domain={[0, 100]}
                    tick={{fontSize: 10}}
                    label={{value: 'CPU Util %', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}
                  />
                  <Tooltip
                    contentStyle={{fontSize: 12}}
                    formatter={(value: number) => `${value}%`}
                  />
                  <Legend wrapperStyle={{fontSize: 12}}/>
                  <Line
                    type="monotone"
                    dataKey="avgUtil"
                    stroke="#38a169"
                    name="Avg Utilization"
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="maxUtil"
                    stroke="#dd6b20"
                    name="Max Utilization"
                    strokeWidth={1}
                    strokeDasharray="5 5"
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            )}
          </VStack>
        </Box>

        {/* Memory Utilization Over Time */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Memory Utilization Trend</Text>
            {memoryTimeseriesQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg"/>
              </Box>
            ) : !hasMemoryData ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="sm" color="gray.500">No memory data available</Text>
              </Box>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={memoryTimeSeriesData}>
                  <CartesianGrid strokeDasharray="3 3"/>
                  <XAxis
                    dataKey="time"
                    tick={{fontSize: 10}}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    domain={[0, 100]}
                    tick={{fontSize: 10}}
                    label={{value: 'Memory Util %', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}
                  />
                  <Tooltip
                    contentStyle={{fontSize: 12}}
                    formatter={(value: number) => `${value}%`}
                  />
                  <Legend wrapperStyle={{fontSize: 12}}/>
                  <Line
                    type="monotone"
                    dataKey="avgUtil"
                    stroke="#805ad5"
                    name="Avg Utilization"
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="maxUtil"
                    stroke="#d53f8c"
                    name="Max Utilization"
                    strokeWidth={1}
                    strokeDasharray="5 5"
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            )}
          </VStack>
        </Box>

        {/* GPU Count Over Time */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Active GPUs Trend</Text>
            {gpuTimeseriesQ.isLoading ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg"/>
              </Box>
            ) : !hasGpuData ? (
              <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
                <Text fontSize="sm" color="gray.500">No GPU data available</Text>
              </Box>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={gpuTimeSeriesData}>
                  <CartesianGrid strokeDasharray="3 3"/>
                  <XAxis
                    dataKey="time"
                    tick={{fontSize: 10}}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    tick={{fontSize: 10}}
                    label={{value: 'GPU Count', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}
                  />
                  <Tooltip
                    contentStyle={{fontSize: 12}}
                  />
                  <Legend wrapperStyle={{fontSize: 12}}/>
                  <Line
                    type="monotone"
                    dataKey="gpuCount"
                    stroke="#38a169"
                    name="Active GPUs"
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            )}
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
