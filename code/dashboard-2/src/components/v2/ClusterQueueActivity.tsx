import { VStack, Text, SimpleGrid, Box, Skeleton, Tag } from '@chakra-ui/react'
import { useNavigate } from 'react-router'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ICellRendererParams, RowClickedEvent } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import type { PartitionResponse } from '../../client'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'

// Cell renderer for running jobs (green badge)
const runningCellRenderer = (params: ICellRendererParams) => {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
      <Tag.Root size="lg" colorPalette="green" variant="solid">
        <Tag.Label>{params.value}</Tag.Label>
      </Tag.Root>
    </div>
  )
}

// Cell renderer for pending jobs (orange badge)
const pendingCellRenderer = (params: ICellRendererParams) => {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
      <Tag.Root size="lg" colorPalette="orange" variant="solid">
        <Tag.Label>{params.value}</Tag.Label>
      </Tag.Root>
    </div>
  )
}

// Cell renderer for total (bold)
const totalCellRenderer = (params: ICellRendererParams) => {
  return <span style={{ fontWeight: '600' }}>{params.value}</span>
}

// Cell renderer for GPU count
const gpuCellRenderer = (params: ICellRendererParams) => {
  return (
    <span style={{
      fontWeight: '600',
      color: params.value > 0 ? '#3182CE' : undefined,
    }}>
      {params.value}
    </span>
  )
}

// AG Grid column definitions for partitions
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

// AG Grid column definitions for merged users
const mergedUserColumns: ColDef[] = [
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
    headerName: 'Run',
    flex: 1,
    sortable: true,
    type: 'numericColumn',
    cellRenderer: runningCellRenderer,
    cellStyle: { textAlign: 'center' }
  },
  {
    field: 'pending',
    headerName: 'Pend',
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
  },
  {
    field: 'cpusInUse',
    headerName: 'CPUs',
    flex: 1,
    sortable: true,
    type: 'numericColumn',
    cellStyle: { textAlign: 'center' }
  },
  {
    field: 'gpusInUse',
    headerName: 'GPUs',
    flex: 1,
    sortable: true,
    type: 'numericColumn',
    cellRenderer: gpuCellRenderer,
    cellStyle: { textAlign: 'center' }
  }
]

export const ClusterQueueActivity = () => {
  const navigate = useNavigate()
  const { cluster, partitionsQuery: partitionsQ } = useClusterOverviewContext()

  // Check loading state
  const isLoading = partitionsQ.isLoading

  // Early return for loading state
  if (isLoading) {
    return (
      <VStack w="100%" align="start" gap={4}>
        <Text fontSize="lg" fontWeight="semibold">Queue Activity</Text>
        <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
          <Skeleton height="330px" rounded="md" />
          <Skeleton height="330px" rounded="md" />
        </SimpleGrid>
      </VStack>
    )
  }

  const partitionsMap = (partitionsQ.data ?? {}) as Record<string, PartitionResponse>
  const partitions = Object.values(partitionsMap)

  // Aggregate job counts across all partitions
  const jobsByUser: Record<string, { running: number; pending: number; cpusInUse: number; gpusInUse: number }> = {}
  const jobsByPartition: Array<{ partition: string; running: number; pending: number }> = []

  for (const partition of partitions) {
    const runningJobs = partition.jobs_running ?? []
    const pendingJobs = partition.jobs_pending ?? []

    jobsByPartition.push({
      partition: partition.name ?? 'Unknown',
      running: runningJobs.length,
      pending: pendingJobs.length
    })

    // Track by user
    for (const job of runningJobs) {
      const user = job.user_name ?? 'Unknown'
      if (!jobsByUser[user]) {
        jobsByUser[user] = { running: 0, pending: 0, cpusInUse: 0, gpusInUse: 0 }
      }
      jobsByUser[user].running++
      jobsByUser[user].cpusInUse += job.requested_cpus ?? 0
      jobsByUser[user].gpusInUse += job.requested_gpus ?? 0
    }

    for (const job of pendingJobs) {
      const user = job.user_name ?? 'Unknown'
      if (!jobsByUser[user]) {
        jobsByUser[user] = { running: 0, pending: 0, cpusInUse: 0, gpusInUse: 0 }
      }
      jobsByUser[user].pending++
    }
  }

  // Merged users table with all columns
  const mergedUsers = Object.entries(jobsByUser)
    .map(([user, counts]) => ({
      user,
      running: counts.running,
      pending: counts.pending,
      total: counts.running + counts.pending,
      cpusInUse: counts.cpusInUse,
      gpusInUse: counts.gpusInUse,
    }))
    .sort((a, b) => b.total - a.total)

  // Sort partitions by total jobs
  const sortedPartitions = jobsByPartition
    .map(p => ({ ...p, total: p.running + p.pending }))
    .sort((a, b) => b.total - a.total)

  return (
    <VStack w="100%" align="start" gap={4}>
      <Text fontSize="lg" fontWeight="semibold">Queue Activity</Text>

      <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
        {/* Users Table - Merged with all columns */}
        <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="sm" fontWeight="semibold" color="gray.700">Users</Text>
            <Box w="100%" h="300px">
              <AgGridReact
                theme={themeQuartz}
                rowData={mergedUsers}
                columnDefs={mergedUserColumns}
                defaultColDef={{ resizable: true, sortable: true }}
                domLayout="normal"
                suppressCellFocus
                onRowClicked={(e: RowClickedEvent) => {
                  if (e.data?.user) navigate(`/v2/${cluster}/jobs/query`)
                }}
              />
            </Box>
          </VStack>
        </Box>

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
                onRowClicked={(e: RowClickedEvent) => {
                  if (e.data?.partition) navigate(`/v2/${cluster}/partitions/${e.data.partition}`)
                }}
              />
            </Box>
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
