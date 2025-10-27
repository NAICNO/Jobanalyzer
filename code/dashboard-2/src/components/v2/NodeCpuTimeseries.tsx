import { useMemo, useState } from 'react'
import { Accordion, Alert, Box, HStack, Spinner, Text, VStack } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'

import { getClusterByClusterNodesByNodenameCpuTimeseriesOptions } from '../../client/@tanstack/react-query.gen'
import type { GetClusterByClusterNodesByNodenameCpuTimeseriesResponse, SampleProcessAccResponse } from '../../client'
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts'

interface Props {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
  resolutionSec?: number
}

function toSeries(data: GetClusterByClusterNodesByNodenameCpuTimeseriesResponse | undefined, nodename: string) {
  if (!data) return [] as Array<{ time: Date; cpu_util: number; cpu_avg: number }>
  // Response: { [key: string]: SampleProcessAccResponse[] }
  const map = data as Record<string, SampleProcessAccResponse[]>
  const arr = map[nodename] ?? map[Object.keys(map)[0]] ?? []
  return arr
    .filter((s) => s && s.time)
    .map((s) => ({ time: s.time, cpu_util: s.cpu_util, cpu_avg: s.cpu_avg }))
    .sort((a, b) => a.time.getTime() - b.time.getTime())
}

export const NodeCpuTimeseries = ({ cluster, nodename, initialCollapsed = false, resolutionSec }: Props) => {
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['cpu'])
  const expanded = value.includes('cpu')

  const queryOpts = getClusterByClusterNodesByNodenameCpuTimeseriesOptions({
    path: { cluster, nodename },
    query: resolutionSec ? { resolution_in_s: resolutionSec } : undefined,
  })

  const { data, isLoading, isError, error } = useQuery({
    ...queryOpts,
    enabled: !!cluster && !!nodename && expanded,
    staleTime: 60_000,
  })

  const series = useMemo(() => toSeries(data, nodename), [data, nodename])

  return (
    <VStack align="start" w="100%" mt={2} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root value={value} onValueChange={(e) => setValue(e.value)} collapsible variant="plain">
          <Accordion.Item value="cpu">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">CPU Usage</Text>
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody px={4}>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading CPU timeseries…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load CPU timeseries.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && series.length > 0 && (
                  <Box w="100%" h="240px" px={2}>
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={series} margin={{ left: 8, right: 8, top: 8, bottom: 8 }}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                          dataKey={(d) => (d.time instanceof Date ? d.time.getTime() : new Date(d.time).getTime())}
                          type="number"
                          domain={['auto', 'auto']}
                          tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                        />
                        <YAxis domain={[0, 100]} tickFormatter={(v) => `${v}%`} />
                        <Tooltip
                          labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()}
                          formatter={(val: number, name) => [ `${val.toFixed(1)}%`, name === 'cpu_util' ? 'CPU Util' : 'CPU Avg' ]}
                        />
                        <Line type="monotone" dataKey="cpu_util" stroke="#3182CE" dot={false} strokeWidth={2} name="cpu_util" />
                        <Line type="monotone" dataKey="cpu_avg" stroke="#A0AEC0" dot={false} strokeWidth={1.5} name="cpu_avg" />
                      </LineChart>
                    </ResponsiveContainer>
                  </Box>
                )}
                {!isLoading && !isError && series.length === 0 && (
                  <Text color="gray.600">No CPU data available.</Text>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
