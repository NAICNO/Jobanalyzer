import { useMemo, useState } from 'react'
import { Accordion, Alert, Box, HStack, Spinner, Text, VStack } from '@chakra-ui/react'

import type { GetClusterByClusterNodesByNodenameGpuTimeseriesResponse, SampleGpuTimeseriesResponse, SampleGpuBaseResponse } from '../../client'
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid, Legend } from 'recharts'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeGpuTimeseries } from '../../hooks/v2/useNodeQueries'

interface Props {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
  resolutionSec?: number
}

const GPU_COLORS = ['#3182CE', '#E53E3E', '#38A169', '#805AD5', '#DD6B20', '#319795', '#D53F8C', '#718096']

interface GpuDataPoint {
  time: number
  [key: string]: number
}

function toUtilSeries(data: GetClusterByClusterNodesByNodenameGpuTimeseriesResponse | undefined, nodename: string) {
  if (!data) return { utilData: [] as GpuDataPoint[], memData: [] as GpuDataPoint[], tempData: [] as GpuDataPoint[], gpuLabels: [] as string[] }

  const map = data as Record<string, SampleGpuTimeseriesResponse[]>
  const gpuArr = map[nodename] ?? map[Object.keys(map)[0]] ?? []

  if (gpuArr.length === 0) return { utilData: [], memData: [], tempData: [], gpuLabels: [] }

  const gpuLabels = gpuArr.map((g) => `GPU ${g.index}`)

  // Build a time-indexed map for each metric
  const utilByTime = new Map<number, GpuDataPoint>()
  const memByTime = new Map<number, GpuDataPoint>()
  const tempByTime = new Map<number, GpuDataPoint>()

  gpuArr.forEach((gpu) => {
    const key = `gpu_${gpu.index}`
    gpu.data.forEach((sample: SampleGpuBaseResponse) => {
      const t = new Date(sample.time).getTime()

      if (!utilByTime.has(t)) utilByTime.set(t, { time: t })
      utilByTime.get(t)![key] = sample.ce_util

      if (!memByTime.has(t)) memByTime.set(t, { time: t })
      memByTime.get(t)![key] = sample.memory_util

      if (!tempByTime.has(t)) tempByTime.set(t, { time: t })
      tempByTime.get(t)![key] = sample.temperature
    })
  })

  const sort = (m: Map<number, GpuDataPoint>) => Array.from(m.values()).sort((a, b) => a.time - b.time)

  return {
    utilData: sort(utilByTime),
    memData: sort(memByTime),
    tempData: sort(tempByTime),
    gpuLabels,
  }
}

function GpuChart({ data, gpuLabels, yLabel, yDomain, yFormatter, title }: {
  data: GpuDataPoint[]
  gpuLabels: string[]
  yLabel: string
  yDomain: [number, number] | [string, string]
  yFormatter: (v: number) => string
  title: string
}) {
  if (data.length === 0) return null
  return (
    <VStack align="start" w="100%" gap={1}>
      <Text fontSize="sm" fontWeight="medium" color="fg.muted">{title}</Text>
      <Box w="100%" h="200px">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data} margin={{ left: 8, right: 8, top: 4, bottom: 4 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="time"
              type="number"
              domain={['auto', 'auto']}
              tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
            />
            <YAxis domain={yDomain} tickFormatter={yFormatter} />
            <Tooltip
              labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()}
              formatter={(val: number, name: string) => {
                const idx = parseInt(name.replace('gpu_', ''))
                return [yFormatter(val), `GPU ${idx}`]
              }}
            />
            <Legend formatter={(v) => { const idx = parseInt(v.replace('gpu_', '')); return gpuLabels[idx] ?? v }} />
            {gpuLabels.map((_, i) => (
              <Line key={i} type="monotone" dataKey={`gpu_${i}`} stroke={GPU_COLORS[i % GPU_COLORS.length]} dot={false} strokeWidth={1.5} />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </Box>
    </VStack>
  )
}

export const NodeGpuTimeseries = ({ cluster, nodename, initialCollapsed = false, resolutionSec }: Props) => {
  const client = useClusterClient(cluster)
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['gpu-ts'])
  const expanded = value.includes('gpu-ts')

  const { data, isLoading, isError, error } = useNodeGpuTimeseries({ cluster, nodename, client, resolutionSec, enabled: expanded })

  const { utilData, memData, tempData, gpuLabels } = useMemo(() => toUtilSeries(data, nodename), [data, nodename])

  const hasData = utilData.length > 0 || memData.length > 0 || tempData.length > 0

  return (
    <VStack align="start" w="100%" mt={2} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root value={value} onValueChange={(e) => setValue(e.value)} collapsible variant="plain">
          <Accordion.Item value="gpu-ts">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">GPU Performance</Text>
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody px={4}>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading GPU timeseries…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load GPU timeseries.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && hasData && (
                  <VStack w="100%" gap={4}>
                    <GpuChart data={utilData} gpuLabels={gpuLabels} title="Compute Utilization" yLabel="%" yDomain={[0, 100]} yFormatter={(v) => `${v}%`} />
                    <GpuChart data={memData} gpuLabels={gpuLabels} title="Memory Utilization" yLabel="%" yDomain={[0, 100]} yFormatter={(v) => `${v}%`} />
                    <GpuChart data={tempData} gpuLabels={gpuLabels} title="Temperature" yLabel="°C" yDomain={['auto', 'auto']} yFormatter={(v) => `${v}°C`} />
                  </VStack>
                )}
                {!isLoading && !isError && !hasData && (
                  <Text color="gray.600">No GPU timeseries data available.</Text>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
