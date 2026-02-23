import { VStack, Text, SimpleGrid, Box } from '@chakra-ui/react'
import { useNavigate } from 'react-router'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ICellRendererParams, RowClickedEvent } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import type { PartitionResponse } from '../../client'
import { useClusterOverviewContext } from '../../contexts/ClusterOverviewContext'

export const ClusterQueueActivity = () => {
  const navigate = useNavigate()
  const { cluster, partitionsQuery: partitionsQ } = useClusterOverviewContext()

  const partitionsMap = (partitionsQ.data ?? {}) as Record<string, PartitionResponse>
  const partitions = Object.values(partitionsMap)

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
                onRowClicked={(e: RowClickedEvent) => {
                  if (e.data?.partition) navigate(`/v2/${cluster}/partitions/${e.data.partition}`)
                }}
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
                onRowClicked={(e: RowClickedEvent) => {
                  if (e.data?.user) navigate(`/v2/${cluster}/jobs/query`)
                }}
              />
            </Box>
          </VStack>
        </Box>
      </SimpleGrid>
    </VStack>
  )
}
