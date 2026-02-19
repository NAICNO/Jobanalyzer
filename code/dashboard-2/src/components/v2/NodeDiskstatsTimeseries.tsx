import { useMemo, useState } from 'react'
import { Accordion, Alert, Badge, Box, HStack, Select, SimpleGrid, Spinner, Text, VStack, createListCollection } from '@chakra-ui/react'

import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid, Legend, AreaChart, Area } from 'recharts'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeDiskstatsTimeseries } from '../../hooks/v2/useNodeQueries'
import { transformDiskstatsTimeseries } from '../../util/timeseriesTransformers'

interface Props {
  cluster: string
  nodename: string
  initialCollapsed?: boolean
  resolutionSec?: number
}

export const NodeDiskstatsTimeseries = ({ cluster, nodename, initialCollapsed = false, resolutionSec }: Props) => {
  const client = useClusterClient(cluster)
  const [value, setValue] = useState<string[]>(initialCollapsed ? [] : ['diskstats'])
  const [selectedDisk, setSelectedDisk] = useState<string>('')
  const expanded = value.includes('diskstats')

  const { data, isLoading, isError, error } = useNodeDiskstatsTimeseries({ cluster, nodename, client, resolutionSec, enabled: expanded })

  const diskData = useMemo(() => transformDiskstatsTimeseries(data, nodename), [data, nodename])
  const diskNames = Object.keys(diskData).sort()
  const activeDiskName = selectedDisk || diskNames[0] || ''
  const activeDisk = diskData[activeDiskName]
  const series = activeDisk?.data ?? []

  // Create collection for Select
  const diskCollection = useMemo(
    () =>
      createListCollection({
        items: diskNames.map((name) => ({ label: name, value: name })),
      }),
    [diskNames]
  )

  // Auto-select first disk when data loads
  useMemo(() => {
    if (diskNames.length > 0 && !selectedDisk) {
      setSelectedDisk(diskNames[0])
    }
  }, [diskNames, selectedDisk])

  return (
    <VStack align="start" w="100%" mt={2} gap={2}>
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md">
        <Accordion.Root value={value} onValueChange={(e) => setValue(e.value)} collapsible variant="plain">
          <Accordion.Item value="diskstats">
            <Accordion.ItemTrigger px={3} py={2} _hover={{ bg: 'gray.50' }}>
              <HStack flex="1" justify="space-between" align="center">
                <Text fontWeight="semibold">Disk I/O Statistics</Text>
                {diskNames.length > 0 && !isLoading && (
                  <Badge colorPalette="blue" size="sm">
                    {diskNames.length} disk{diskNames.length !== 1 ? 's' : ''}
                  </Badge>
                )}
              </HStack>
              <Accordion.ItemIndicator />
            </Accordion.ItemTrigger>
            <Accordion.ItemContent>
              <Accordion.ItemBody px={4}>
                {isLoading && (
                  <HStack>
                    <Spinner size="sm" />
                    <Text>Loading disk I/O timeseries…</Text>
                  </HStack>
                )}
                {isError && (
                  <Alert.Root status="error">
                    <Alert.Indicator />
                    <Alert.Description>
                      {error instanceof Error ? error.message : 'Failed to load disk I/O timeseries.'}
                    </Alert.Description>
                  </Alert.Root>
                )}
                {!isLoading && !isError && diskNames.length === 0 && (
                  <Text color="gray.600">No disk data available.</Text>
                )}
                {!isLoading && !isError && diskNames.length > 0 && (
                  <VStack align="start" w="100%" gap={4}>
                    {/* Disk Selector */}
                    {diskNames.length > 1 && (
                      <Box w="100%" maxW="300px">
                        <Select.Root
                          collection={diskCollection}
                          value={[activeDiskName]}
                          onValueChange={(details) => setSelectedDisk(details.value[0])}
                          size="sm"
                        >
                          <Select.Label>Select Disk</Select.Label>
                          <Select.Control>
                            <Select.Trigger>
                              <Select.ValueText placeholder="Select a disk" />
                            </Select.Trigger>
                          </Select.Control>
                          <Select.Content>
                            {diskNames.map((name) => (
                              <Select.Item key={name} item={name}>
                                <Select.ItemText>{name}</Select.ItemText>
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select.Root>
                      </Box>
                    )}

                    <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
                      {/* IOPS Chart */}
                      <VStack align="start" gap={1}>
                        <Text fontSize="sm" fontWeight="medium" color="fg.muted">IOPS</Text>
                        <Box w="100%" h="220px">
                          <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={series} margin={{ left: 0, right: 8, top: 8, bottom: 8 }}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis dataKey="timeMs" type="number" domain={['auto', 'auto']} tickFormatter={(ts) => new Date(ts).toLocaleTimeString()} />
                              <YAxis tickFormatter={(v) => v.toFixed(0)} />
                              <Tooltip labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()} />
                              <Legend />
                              <Line type="monotone" dataKey="read_iops" stroke="#3182CE" dot={false} strokeWidth={2} name="Read" />
                              <Line type="monotone" dataKey="write_iops" stroke="#E53E3E" dot={false} strokeWidth={2} name="Write" />
                            </LineChart>
                          </ResponsiveContainer>
                        </Box>
                      </VStack>

                      {/* Throughput Chart */}
                      <VStack align="start" gap={1}>
                        <Text fontSize="sm" fontWeight="medium" color="fg.muted">Throughput (MB/s)</Text>
                        <Box w="100%" h="220px">
                          <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={series} margin={{ left: 0, right: 8, top: 8, bottom: 8 }}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis dataKey="timeMs" type="number" domain={['auto', 'auto']} tickFormatter={(ts) => new Date(ts).toLocaleTimeString()} />
                              <YAxis tickFormatter={(v) => `${v.toFixed(1)}`} />
                              <Tooltip labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()} />
                              <Legend />
                              <Line type="monotone" dataKey="read_throughput_mb" stroke="#38A169" dot={false} strokeWidth={2} name="Read" />
                              <Line type="monotone" dataKey="write_throughput_mb" stroke="#DD6B20" dot={false} strokeWidth={2} name="Write" />
                            </LineChart>
                          </ResponsiveContainer>
                        </Box>
                      </VStack>

                      {/* Latency & Queue Depth Chart */}
                      <VStack align="start" gap={1}>
                        <Text fontSize="sm" fontWeight="medium" color="fg.muted">Latency & Queue Depth</Text>
                        <Box w="100%" h="220px">
                          <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={series} margin={{ left: 0, right: 8, top: 8, bottom: 8 }}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis dataKey="timeMs" type="number" domain={['auto', 'auto']} tickFormatter={(ts) => new Date(ts).toLocaleTimeString()} />
                              <YAxis yAxisId="left" tickFormatter={(v) => `${v.toFixed(1)} ms`} />
                              <YAxis yAxisId="right" orientation="right" tickFormatter={(v) => v.toFixed(0)} />
                              <Tooltip labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()} />
                              <Legend />
                              <Line yAxisId="left" type="monotone" dataKey="read_latency_ms" stroke="#805AD5" dot={false} strokeWidth={2} name="Read (ms)" />
                              <Line yAxisId="left" type="monotone" dataKey="write_latency_ms" stroke="#D69E2E" dot={false} strokeWidth={2} name="Write (ms)" />
                              <Line yAxisId="right" type="monotone" dataKey="ios_in_progress" stroke="#718096" dot={false} strokeWidth={1.5} strokeDasharray="5 5" name="Queue" />
                            </LineChart>
                          </ResponsiveContainer>
                        </Box>
                      </VStack>

                      {/* Utilization Chart */}
                      <VStack align="start" gap={1}>
                        <Text fontSize="sm" fontWeight="medium" color="fg.muted">Utilization</Text>
                        <Box w="100%" h="220px">
                          <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={series} margin={{ left: 0, right: 8, top: 8, bottom: 8 }}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis dataKey="timeMs" type="number" domain={['auto', 'auto']} tickFormatter={(ts) => new Date(ts).toLocaleTimeString()} />
                              <YAxis domain={[0, 100]} tickFormatter={(v) => `${v}%`} />
                              <Tooltip labelFormatter={(ts) => new Date(Number(ts)).toLocaleString()} />
                              <Area type="monotone" dataKey="utilization_pct" stroke="#319795" fill="#319795" fillOpacity={0.3} strokeWidth={2} name="Utilization" />
                            </AreaChart>
                          </ResponsiveContainer>
                        </Box>
                      </VStack>
                    </SimpleGrid>
                  </VStack>
                )}
              </Accordion.ItemBody>
            </Accordion.ItemContent>
          </Accordion.Item>
        </Accordion.Root>
      </Box>
    </VStack>
  )
}
