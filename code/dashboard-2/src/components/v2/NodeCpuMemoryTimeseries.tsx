import { useMemo, useState } from 'react'
import { Accordion, Alert, Box, HStack, SimpleGrid, Spinner, Text, VStack } from '@chakra-ui/react'

import type {
  GetClusterByClusterNodesByNodenameCpuTimeseriesResponse,
  GetClusterByClusterNodesByNodenameMemoryTimeseriesResponse,
  SampleProcessAccResponse,
} from '../../client'
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid, Legend } from 'recharts'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeCpuTimeseries, useNodeMemoryTimeseries } from '../../hooks/v2/useNodeQueries'

interface Props {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
  resolutionSec?: number
}

function toCpuSeries(data: GetClusterByClusterNodesByNodenameCpuTimeseriesResponse | undefined, nodename: string) {
  if (!data) return []
  const map = data as Record<string, SampleProcessAccResponse[]>
  const arr = map[nodename] ?? map[Object.keys(map)[0]] ?? []
  return arr
    .filter((s) => s && s.time)
    .map((s) => ({ time: new Date(s.time).getTime(), cpu_util: s.cpu_util, cpu_avg: s.cpu_avg }))
    .sort((a, b) => a.time - b.time)
}

function toMemSeries(data: GetClusterByClusterNodesByNodenameMemoryTimeseriesResponse | undefined, nodename: string) {
  if (!data) return []
  const map = data as Record<string, SampleProcessAccResponse[]>
  const arr = map[nodename] ?? map[Object.keys(map)[0]] ?? []
  return arr
    .filter((s) => s && s.time)
    .map((s) => ({
      time: new Date(s.time).getTime(),
      memory_resident_gb: s.memory_resident / (1024 * 1024),
      memory_virtual_gb: s.memory_virtual / (1024 * 1024),
      memory_util: s.memory_util,
    }))
    .sort((a, b) => a.time - b.time)
}

export const NodeCpuMemoryTimeseries = ({ cluster, nodename, initialCollapsed = false, resolutionSec }: Props) => {
  const client = useClusterClient(cluster)
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['cpu-mem'])
  const expanded = value.includes('cpu-mem')

  const { data: cpuData, isLoading: cpuLoading, isError: cpuError, error: cpuErr } = useNodeCpuTimeseries({ cluster, nodename, client, resolutionSec, enabled: expanded })
  const { data: memData, isLoading: memLoading, isError: memError, error: memErr } = useNodeMemoryTimeseries({ cluster, nodename, client, resolutionSec, enabled: expanded })

  const cpuSeries = useMemo(() => toCpuSeries(cpuData, nodename), [cpuData, nodename])
  const memSeries = useMemo(() => toMemSeries(memData, nodename), [memData, nodename])

  const isLoading = cpuLoading || memLoading
  const isError = cpuError || memError
  const errorMsg = cpuErr instanceof Error ? cpuErr.message : memErr instanceof Error ? memErr.message : 'Failed to load timeseries.'

  return (
    <VStack align="start" w="100%" mt={2} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root value={value} onValueChange={(e) => setValue(e.value)} collapsible variant="plain">
          <Accordion.Item value="cpu-mem">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">CPU & Memory Usage</Text>
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody px={4}>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading timeseries…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>{errorMsg}</Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && (
                  <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
                    {cpuSeries.length > 0 ? (
                      <VStack align="start" gap={1}>
                        <Text fontSize="sm" fontWeight="medium" color="fg.muted">CPU</Text>
                        <Box w="100%" h="220px">
                          <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={cpuSeries} margin={{ left: 0, right: 8, top: 8, bottom: 8 }}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis
                                dataKey="time"
                                type="number"
                                domain={['auto', 'auto']}
                                tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                              />
                              <YAxis domain={[0, 100]} tickFormatter={(v) => `${v}%`} />
                              <Tooltip
                                labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()}
                                formatter={(val: number, name: string) => [`${val.toFixed(1)}%`, name === 'cpu_util' ? 'Util' : 'Avg']}
                              />
                              <Legend formatter={(v) => v === 'cpu_util' ? 'Util' : 'Avg'} />
                              <Line type="monotone" dataKey="cpu_util" stroke="#3182CE" dot={false} strokeWidth={2} />
                              <Line type="monotone" dataKey="cpu_avg" stroke="#A0AEC0" dot={false} strokeWidth={1.5} />
                            </LineChart>
                          </ResponsiveContainer>
                        </Box>
                      </VStack>
                    ) : (
                      <Text color="gray.600">No CPU data available.</Text>
                    )}

                    {memSeries.length > 0 ? (
                      <VStack align="start" gap={1}>
                        <Text fontSize="sm" fontWeight="medium" color="fg.muted">Memory</Text>
                        <Box w="100%" h="220px">
                          <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={memSeries} margin={{ left: 0, right: 8, top: 8, bottom: 8 }}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis
                                dataKey="time"
                                type="number"
                                domain={['auto', 'auto']}
                                tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                              />
                              <YAxis yAxisId="gb" tickFormatter={(v) => `${v.toFixed(0)} GB`} />
                              <YAxis yAxisId="pct" orientation="right" domain={[0, 100]} tickFormatter={(v) => `${v}%`} />
                              <Tooltip
                                labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()}
                                formatter={(val: number, name: string) => {
                                  if (name === 'memory_util') return [`${val.toFixed(1)}%`, 'Util']
                                  return [`${val.toFixed(1)} GB`, name === 'memory_resident_gb' ? 'Resident' : 'Virtual']
                                }}
                              />
                              <Legend formatter={(v) => v === 'memory_resident_gb' ? 'Resident' : v === 'memory_virtual_gb' ? 'Virtual' : 'Util %'} />
                              <Line yAxisId="gb" type="monotone" dataKey="memory_resident_gb" stroke="#3182CE" dot={false} strokeWidth={2} />
                              <Line yAxisId="gb" type="monotone" dataKey="memory_virtual_gb" stroke="#A0AEC0" dot={false} strokeWidth={1.5} />
                              <Line yAxisId="pct" type="monotone" dataKey="memory_util" stroke="#38A169" dot={false} strokeWidth={1.5} />
                            </LineChart>
                          </ResponsiveContainer>
                        </Box>
                      </VStack>
                    ) : (
                      <Text color="gray.600">No memory data available.</Text>
                    )}
                  </SimpleGrid>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
