import { useMemo, useState } from 'react'
import { Accordion, Alert, Box, HStack, Spinner, Text, VStack } from '@chakra-ui/react'

import type { GetClusterByClusterNodesByNodenameMemoryTimeseriesResponse, SampleProcessAccResponse } from '../../client'
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid, Legend } from 'recharts'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeMemoryTimeseries } from '../../hooks/v2/useNodeQueries'

interface Props {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
  resolutionSec?: number
}

function toSeries(data: GetClusterByClusterNodesByNodenameMemoryTimeseriesResponse | undefined, nodename: string) {
  if (!data) return []
  const map = data as Record<string, SampleProcessAccResponse[]>
  const arr = map[nodename] ?? map[Object.keys(map)[0]] ?? []
  return arr
    .filter((s) => s && s.time)
    .map((s) => ({
      time: new Date(s.time),
      memory_resident_gb: s.memory_resident / (1024 * 1024),
      memory_virtual_gb: s.memory_virtual / (1024 * 1024),
      memory_util: s.memory_util,
    }))
    .sort((a, b) => a.time.getTime() - b.time.getTime())
}

export const NodeMemoryTimeseries = ({ cluster, nodename, initialCollapsed = false, resolutionSec }: Props) => {
  const client = useClusterClient(cluster)
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['memory'])
  const expanded = value.includes('memory')

  const { data, isLoading, isError, error } = useNodeMemoryTimeseries({ cluster, nodename, client, resolutionSec, enabled: expanded })

  const series = useMemo(() => toSeries(data, nodename), [data, nodename])

  return (
    <VStack align="start" w="100%" mt={2} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root value={value} onValueChange={(e) => setValue(e.value)} collapsible variant="plain">
          <Accordion.Item value="memory">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">Memory Usage</Text>
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody px={4}>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading memory timeseries…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load memory timeseries.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && series.length > 0 && (
                  <Box w="100%" h="240px" px={2}>
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={series} margin={{ left: 8, right: 8, top: 8, bottom: 8 }}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                          dataKey={(d) => d.time.getTime()}
                          type="number"
                          domain={['auto', 'auto']}
                          tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                        />
                        <YAxis yAxisId="gb" tickFormatter={(v) => `${v.toFixed(0)} GB`} />
                        <YAxis yAxisId="pct" orientation="right" domain={[0, 100]} tickFormatter={(v) => `${v}%`} />
                        <Tooltip
                          labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()}
                          formatter={(val: number, name: string) => {
                            if (name === 'memory_util') return [`${val.toFixed(1)}%`, 'Mem Util']
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
                )}
                {!isLoading && !isError && series.length === 0 && (
                  <Text color="gray.600">No memory data available.</Text>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
