import { VStack, HStack, Text, SimpleGrid, Box, Stat, Badge } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ICellRendererParams } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import { getClusterByClusterPartitionsOptions } from '../../client/@tanstack/react-query.gen'
import type { PartitionResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'

interface Props {
  cluster: string
}

export const ClusterQueueActivity = ({ cluster }: Props) => {
  const client = useClusterClient(cluster)
  
  const partitionsQ = useQuery({
    ...getClusterByClusterPartitionsOptions({ path: { cluster }, client }),
    enabled: !!cluster,
  })

  const partitions = (partitionsQ.data ?? []) as PartitionResponse[]

  // Aggregate job counts across all partitions
  let totalRunning = 0
  let totalPending = 0
  const jobsByUser: Record<string, { running: number; pending: number }> = {}
  const jobsByPartition: Array<{ partition: string; running: number; pending: number }> = []

  for (const partition of partitions) {
    const runningJobs = partition.jobs_running ?? []
    const pendingJobs = partition.jobs_pending ?? []

    totalRunning += runningJobs.length
    totalPending += pendingJobs.length

    jobsByPartition.push({
      partition: partition.name ?? 'Unknown',
      running: runningJobs.length,
      pending: pendingJobs.length
    })

    // Track by user
    for (const job of runningJobs) {
      const user = job.user_name ?? 'Unknown'
      if (!jobsByUser[user]) {
        jobsByUser[user] = { running: 0, pending: 0 }
      }
      jobsByUser[user].running++
    }

    for (const job of pendingJobs) {
      const user = job.user_name ?? 'Unknown'
      if (!jobsByUser[user]) {
        jobsByUser[user] = { running: 0, pending: 0 }
      }
      jobsByUser[user].pending++
    }
  }

  // Sort users by total jobs
  const topUsers = Object.entries(jobsByUser)
    .map(([user, counts]) => ({
      user,
      running: counts.running,
      pending: counts.pending,
      total: counts.running + counts.pending
    }))
    .sort((a, b) => b.total - a.total)

  // Sort partitions by total jobs
  const sortedPartitions = jobsByPartition
    .map(p => ({ ...p, total: p.running + p.pending }))
    .sort((a, b) => b.total - a.total)

  // Calculate trends (mock for now - would need historical data)
  const runningTrend = totalRunning > 0 ? 'up' : 'neutral'
  const pendingTrend = totalPending > 10 ? 'up' : totalPending > 0 ? 'neutral' : 'down'

  // Cell renderer for running jobs (green badge)
  const runningCellRenderer = (params: ICellRendererParams) => {
    return (
      <span style={{
        display: 'inline-block',
        padding: '0.5px 8px',
        backgroundColor: '#22c55e',
        color: 'white',
        borderRadius: '4px',
        fontSize: '12px',
        fontWeight: '500'
      }}>
        {params.value}
      </span>
    )
  }

  // Cell renderer for pending jobs (orange badge)
  const pendingCellRenderer = (params: ICellRendererParams) => {
    return (
      <span style={{
        display: 'inline-block',
        padding: '0.5px 8px',
        backgroundColor: '#f97316',
        color: 'white',
        borderRadius: '4px',
        fontSize: '12px',
        fontWeight: '500'
      }}>
        {params.value}
      </span>
    )
  }

  // Cell renderer for total (bold)
  const totalCellRenderer = (params: ICellRendererParams) => {
    return <span style={{ fontWeight: '600' }}>{params.value}</span>
  }

  // AG Grid column definitions
  const partitionColumns: ColDef[] = [
    {
      field: 'partition',
      headerName: 'Partition',
      flex: 2,
      sortable: true,
      filter: true,
      cellStyle: { fontWeight: '500' }
    },
    { 
      field: 'running', 
      headerName: 'Running', 
      flex: 1, 
      sortable: true, 
      type: 'numericColumn',
      cellRenderer: runningCellRenderer,
      cellStyle: { textAlign: 'center' }
    },
    { 
      field: 'pending', 
      headerName: 'Pending', 
      flex: 1, 
      sortable: true, 
      type: 'numericColumn',
      cellRenderer: pendingCellRenderer,
      cellStyle: { textAlign: 'center' }
    },
    { 
      field: 'total', 
      headerName: 'Total', 
      flex: 1, 
      sortable: true, 
      type: 'numericColumn',
      cellRenderer: totalCellRenderer,
      cellStyle: { textAlign: 'center' }
    }
  ]

  const userColumns: ColDef[] = [
    { 
      field: 'user', 
      headerName: 'User', 
      flex: 2, 
      sortable: true, 
      filter: true,
      cellStyle: { fontWeight: '500' }
    },
    { 
      field: 'running', 
      headerName: 'Running', 
      flex: 1, 
      sortable: true, 
      type: 'numericColumn',
      cellRenderer: runningCellRenderer,
      cellStyle: { textAlign: 'center' }
    },
    { 
      field: 'pending', 
      headerName: 'Pending', 
      flex: 1, 
      sortable: true, 
      type: 'numericColumn',
      cellRenderer: pendingCellRenderer,
      cellStyle: { textAlign: 'center' }
    },
    { 
      field: 'total', 
      headerName: 'Total', 
      flex: 1, 
      sortable: true, 
      type: 'numericColumn',
      cellRenderer: totalCellRenderer,
      cellStyle: { textAlign: 'center' }
    }
  ]

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Queue Activity</Text>

      {/* Summary Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} gap={3} w="100%">
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Running Jobs</Stat.Label>
            <HStack gap={2}>
              <Stat.ValueText fontSize="2xl" fontWeight="bold" color="green.600">
                {totalRunning}
              </Stat.ValueText>
              {runningTrend === 'up' && (
                <Stat.UpIndicator>
                  <Badge colorPalette="green" size="sm">Active</Badge>
                </Stat.UpIndicator>
              )}
            </HStack>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Pending Jobs</Stat.Label>
            <HStack gap={2}>
              <Stat.ValueText fontSize="2xl" fontWeight="bold" color="orange.600">
                {totalPending}
              </Stat.ValueText>
              {pendingTrend === 'up' && (
                <Stat.UpIndicator>
                  <Badge colorPalette="orange" size="sm">High</Badge>
                </Stat.UpIndicator>
              )}
            </HStack>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Total Jobs</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {totalRunning + totalPending}
            </Stat.ValueText>
          </Stat.Root>
        </Box>

        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <Stat.Root>
            <Stat.Label fontSize="sm" color="gray.600">Active Users</Stat.Label>
            <Stat.ValueText fontSize="2xl" fontWeight="bold">
              {Object.keys(jobsByUser).length}
            </Stat.ValueText>
          </Stat.Root>
        </Box>
      </SimpleGrid>

      {/* Jobs by Partition and User Tables */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
        {/* Jobs by Partition */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Jobs by Partition</Text>
            <Box w="100%" h="300px">
              <AgGridReact
                theme={themeQuartz}
                rowData={sortedPartitions}
                columnDefs={partitionColumns}
                defaultColDef={{
                  resizable: true,
                  sortable: true,
                }}
                domLayout="normal"
                suppressCellFocus
              />
            </Box>
          </VStack>
        </Box>

        {/* Jobs by User */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Top Users (by total jobs)</Text>
            <Box w="100%" h="300px">
              <AgGridReact
                theme={themeQuartz}
                rowData={topUsers}
                columnDefs={userColumns}
                defaultColDef={{
                  resizable: true,
                  sortable: true,
                }}
                domLayout="normal"
                suppressCellFocus
              />
            </Box>
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
