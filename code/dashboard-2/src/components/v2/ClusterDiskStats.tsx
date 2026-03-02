import { VStack, Text, SimpleGrid, Box, Spinner } from '@chakra-ui/react'
import type { ReactNode } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, Brush } from 'recharts'
import type { NodeDiskTimeseriesResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useClusterDiskTimeseries } from '../../hooks/v2/useClusterQueries'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'
import { transformClusterDiskstatsTimeseries } from '../../util/timeseriesTransformers'

const DATA_RESOLUTION = 300 // 5 minutes

const ChartCard = ({ title, isLoading, hasData, children }: {
  title: string
  isLoading: boolean
  hasData: boolean
  children: ReactNode
}) => (
  <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white" role="img" aria-label={`${title} chart`}>
    <VStack align="start" gap={2} w="100%">
      <Text fontSize="sm" fontWeight="semibold" color="gray.700">{title}</Text>
      {isLoading ? (
        <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
          <Spinner size="lg"/>
        </Box>
      ) : !hasData ? (
        <Box w="100%" h="300px" display="flex" alignItems="center" justifyContent="center">
          <Text fontSize="sm" color="gray.500">No disk data available</Text>
        </Box>
      ) : children}
    </VStack>
  </Box>
)

export const ClusterDiskStats = ({ cluster, enabled = true }: { cluster: string; enabled?: boolean }) => {
  const { timeRange } = useClusterOverviewContext()
  const client = useClusterClient(cluster)

  const diskTimeseriesQ = useClusterDiskTimeseries({
    cluster,
    client,
    startTimeInS: timeRange.startTimeInS,
    endTimeInS: timeRange.endTimeInS,
    resolutionInS: DATA_RESOLUTION,
    enabled,
  })

  const diskData = diskTimeseriesQ.data as NodeDiskTimeseriesResponse | undefined
  const diskTimeSeriesData = transformClusterDiskstatsTimeseries(diskData)
  const hasDiskData = diskTimeSeriesData.length > 0
  const loading = diskTimeseriesQ.isLoading

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="xl" fontWeight="bold">
        Disk I/O Metrics
      </Text>

      <SimpleGrid columns={{base: 1, lg: 2}} gap={4} w="100%">
        {/* 1. Disk I/O Utilization */}
        <ChartCard title="Disk I/O Utilization Trend" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis domain={[0, 100]} tick={{fontSize: 10}} label={{value: 'Util %', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="avgUtilization" stroke="#d69e2e" name="Avg Utilization" strokeWidth={2} dot={false}/>
              <Line type="monotone" dataKey="maxUtilization" stroke="#e53e3e" name="Max Utilization" strokeWidth={1} strokeDasharray="5 5" dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 2. Disk IOPS */}
        <ChartCard title="Disk IOPS Trend" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'IOPS', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="readIOPS" stroke="#3182ce" name="Read IOPS" strokeWidth={2} dot={false}/>
              <Line type="monotone" dataKey="writeIOPS" stroke="#dd6b20" name="Write IOPS" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 3. Throughput (MB/s) */}
        <ChartCard title="Disk Throughput" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'MB/s', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="readThroughputMB" stroke="#3182ce" name="Read MB/s" strokeWidth={2} dot={false}/>
              <Line type="monotone" dataKey="writeThroughputMB" stroke="#dd6b20" name="Write MB/s" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 4. Latency (ms/op) */}
        <ChartCard title="Disk Latency" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'ms/op', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="readLatencyMs" stroke="#3182ce" name="Read Latency" strokeWidth={2} dot={false}/>
              <Line type="monotone" dataKey="writeLatencyMs" stroke="#dd6b20" name="Write Latency" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 5. Queue Depth */}
        <ChartCard title="Queue Depth" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'Depth', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="queueDepth" stroke="#805ad5" name="Queue Depth" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 6. Merge Rate */}
        <ChartCard title="I/O Merge Rate" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'merges/s', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="readMergeRate" stroke="#3182ce" name="Read Merges/s" strokeWidth={2} dot={false}/>
              <Line type="monotone" dataKey="writeMergeRate" stroke="#dd6b20" name="Write Merges/s" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 7. Average I/O Size */}
        <ChartCard title="Average I/O Size" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'KB/op', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="avgReadIOSizeKB" stroke="#3182ce" name="Read KB/op" strokeWidth={2} dot={false}/>
              <Line type="monotone" dataKey="avgWriteIOSizeKB" stroke="#dd6b20" name="Write KB/op" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* 8. Average Wait Time */}
        <ChartCard title="Average Wait Time" isLoading={loading} hasData={hasDiskData}>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={diskTimeSeriesData}>
              <CartesianGrid strokeDasharray="3 3"/>
              <XAxis dataKey="time" tick={{fontSize: 10}} interval="preserveStartEnd"/>
              <YAxis tick={{fontSize: 10}} label={{value: 'ms', angle: -90, position: 'insideLeft', style: {fontSize: 10}}}/>
              <Tooltip contentStyle={{fontSize: 12}}/>
              <Legend wrapperStyle={{fontSize: 12}}/>
              <Line type="monotone" dataKey="avgWaitTimeMs" stroke="#38a169" name="Avg Wait Time" strokeWidth={2} dot={false}/>
              <Brush dataKey="time" height={25} stroke="#718096"/>
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>
      </SimpleGrid>
    </VStack>
  )
}
