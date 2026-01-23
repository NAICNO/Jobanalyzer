import { useMemo } from 'react'
import { Box, Text, VStack, HStack, Spinner, Alert, Button, Badge } from '@chakra-ui/react'
import { useParams, useNavigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ValueFormatterParams, ICellRendererParams } from 'ag-grid-community'
import { AllCommunityModule, ModuleRegistry, themeQuartz } from 'ag-grid-community'

import { getClusterByClusterJobsOptions } from '../../client/@tanstack/react-query.gen'
import { useClusterClient } from '../../hooks/useClusterClient'
import type { JobResponse } from '../../client'
import { formatDuration, formatMemory, getJobStateColor } from '../../util/formatters'
import { JobState } from '../../types/jobStates'

ModuleRegistry.registerModules([AllCommunityModule])

// Helper to parse resource strings like "cpu=2,mem=1024,node=1,gres/gpu=1"
const parseResourceValue = (resourceStr: string | null | undefined, key: string): number => {
  if (!resourceStr) return 0
  const regex = new RegExp(`${key}=(\\d+)`)
  const match = resourceStr.match(regex)
  return match ? parseInt(match[1], 10) : 0
}

export const JobsPage = () => {
  const { clusterName } = useParams<{ clusterName: string }>()
  const navigate = useNavigate()

  const client = useClusterClient(clusterName)

  if (!client) {
    return (
      <VStack p={4} align="start">
        <Spinner />
      </VStack>
    )
  }

  // Fetch jobs for the cluster
  const { data, isLoading, isError, error, refetch } = useQuery({
    ...getClusterByClusterJobsOptions({
      path: { cluster: clusterName ?? '' },
      client,
    }),
    enabled: !!clusterName,
  })

  // Deduplicate jobs: Each SLURM job returns 3 entries (main job + batch + step).
  // Keep only the main job entry (job_step="") which has complete metadata.
  const jobs = useMemo(() => {
    const allJobs = (data?.jobs ?? []) as JobResponse[]
    return allJobs.filter(job => job.job_step === '')
  }, [data?.jobs])

  // Column definitions for AG Grid
  const columnDefs = useMemo<ColDef<JobResponse>[]>(
    () => [
      {
        field: 'job_id',
        headerName: 'Job ID',
        width: 115,
        sort: 'desc',
        filter: 'agNumberColumnFilter',
        pinned: 'left',
        cellRenderer: (params: ICellRendererParams<JobResponse>) => {
          return params.value ? params.value : ''
        }
      },
      {
        field: 'job_state',
        headerName: 'State',
        width: 120,
        filter: 'agSetColumnFilter',
        pinned: 'left',
        cellRenderer: (params: ICellRendererParams<JobResponse>) => {
          const state = params.value
          const colorPalette = getJobStateColor(state)
          return <Badge colorPalette={colorPalette} size="sm">{state}</Badge>
        }
      },
      {
        field: 'job_name',
        headerName: 'Job Name',
        flex: 1,
        minWidth: 200,
        filter: 'agTextColumnFilter',
      },
      {
        field: 'user_name',
        headerName: 'User',
        width: 120,
        filter: 'agTextColumnFilter',
      },
      {
        field: 'account',
        headerName: 'Account',
        width: 130,
        filter: 'agTextColumnFilter',
      },
      {
        field: 'partition',
        headerName: 'Partition',
        width: 130,
        filter: 'agTextColumnFilter',
      },
      {
        field: 'requested_node_count',
        headerName: 'Nodes',
        width: 100,
        filter: 'agNumberColumnFilter',
      },
      {
        headerName: 'CPUs\n(Req/Alloc)',
        width: 120,
        filter: 'agTextColumnFilter',
        wrapHeaderText: true,
        autoHeaderHeight: true,
        valueGetter: (params) => {
          const requested = params.data?.requested_cpus || 0
          const allocated = parseResourceValue(params.data?.allocated_resources, 'cpu')
          return `${requested} / ${allocated}`
        }
      },
      {
        headerName: 'GPUs\n(Req/Alloc)',
        width: 120,
        filter: 'agTextColumnFilter',
        wrapHeaderText: true,
        autoHeaderHeight: true,
        valueGetter: (params) => {
          const requested = parseResourceValue(params.data?.requested_resources, 'gres/gpu')
          const allocated = parseResourceValue(params.data?.allocated_resources, 'gres/gpu')
          if (requested === 0 && allocated === 0) return '-'
          return `${requested} / ${allocated}`
        },
        cellRenderer: (params: ICellRendererParams<JobResponse>) => {
          if (params.value === '-') return '-'
          return <Text color="purple.fg" fontWeight="medium">{params.value}</Text>
        }
      },
      {
        headerName: 'Memory\n(Req/Alloc)',
        width: 150,
        filter: 'agTextColumnFilter',
        wrapHeaderText: true,
        autoHeaderHeight: true,
        valueGetter: (params) => {
          const requested = params.data?.requested_memory_per_node || 0
          const allocated = parseResourceValue(params.data?.allocated_resources, 'mem')
          return { requested, allocated }
        },
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          const { requested, allocated } = params.value || { requested: 0, allocated: 0 }
          return `${formatMemory(requested)} / ${formatMemory(allocated)}`
        }
      },
      {
        headerName: 'Elapsed',
        width: 120,
        filter: 'agNumberColumnFilter',
        valueGetter: (params) => {
          const startTime = params.data?.start_time
          const endTime = params.data?.end_time
          if (!startTime) return 0

          const start = new Date(startTime).getTime()
          const end = endTime ? new Date(endTime).getTime() : Date.now()
          return Math.floor((end - start) / 1000)
        },
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          return formatDuration(params.value)
        }
      },
      {
        field: 'time_limit',
        headerName: 'Time Limit',
        width: 120,
        filter: 'agNumberColumnFilter',
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          return formatDuration(params.value)
        }
      },
      {
        field: 'submit_time',
        headerName: 'Submit Time',
        width: 160,
        filter: 'agDateColumnFilter',
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          if (!params.value) return '-'
          return new Date(params.value).toLocaleString(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
          })
        }
      },
      {
        field: 'start_time',
        headerName: 'Start Time',
        width: 160,
        filter: 'agDateColumnFilter',
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          if (!params.value) return '-'
          return new Date(params.value).toLocaleString(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
          })
        }
      },
      {
        field: 'end_time',
        headerName: 'End Time',
        width: 160,
        filter: 'agDateColumnFilter',
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          if (!params.value) return '-'
          return new Date(params.value).toLocaleString(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
          })
        }
      },
      {
        field: 'nodes',
        headerName: 'Node List',
        width: 100,
        filter: 'agTextColumnFilter',
        valueFormatter: (params: ValueFormatterParams<JobResponse>) => {
          if (!params.value) return '-'
          // Handle both string and array formats
          if (typeof params.value === 'string') return params.value
          if (Array.isArray(params.value)) return params.value.join(', ')
          return '-'
        }
      },
      {
        field: 'priority',
        headerName: 'Priority',
        width: 100,
        filter: 'agNumberColumnFilter',
      },
    ],
    []
  )

  const defaultColDef = useMemo<ColDef>(() => ({
    sortable: true,
    resizable: true,
    filter: true,
    enableCellTextSelection: true,
  }), [])

  if (!clusterName) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Missing cluster name in route.</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  if (isError) {
    return (
      <VStack p={4} align="start" gap={4}>
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>
            Failed to load jobs: {error?.message || 'Unknown error'}
          </Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <VStack w="100%" align="start" gap={4} p={4}>
      <HStack w="100%" justify="space-between" align="center">
        <VStack align="start" gap={1}>
          <Text fontSize="2xl" fontWeight="bold">
            Jobs - {clusterName}
          </Text>
          <HStack gap={2}>
            <Text fontSize="sm" color="fg.muted">
              {isLoading ? 'Loading...' : `${jobs.length} jobs`}
            </Text>
            {!isLoading && jobs.length > 0 && (
              <>
                <Text fontSize="sm" color="fg.muted">•</Text>
                <Text fontSize="sm" color="fg.muted">
                  {jobs.filter(j => j.job_state === JobState.RUNNING).length} running
                </Text>
                <Text fontSize="sm" color="fg.muted">•</Text>
                <Text fontSize="sm" color="fg.muted">
                  {jobs.filter(j => j.job_state === JobState.PENDING).length} pending
                </Text>
              </>
            )}
          </HStack>
        </VStack>
        <Button onClick={() => refetch()} size="sm" variant="outline">
          Refresh
        </Button>
      </HStack>

      {isLoading ? (
        <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
          <Spinner size="xl" />
        </Box>
      ) : (
        <Box w="100%" h="calc(100vh - 220px)">
          <AgGridReact<JobResponse>
            theme={themeQuartz}
            rowData={jobs}
            columnDefs={columnDefs}
            defaultColDef={defaultColDef}
            getRowId={(params) => params.data.job_id?.toString() ?? ''}
            pagination={true}
            paginationPageSize={50}
            paginationPageSizeSelector={[25, 50, 100, 200]}
            rowSelection="single"
            animateRows={true}
            enableCellTextSelection={true}
            onRowDoubleClicked={(event) => {
              if (event.data?.job_id) {
                navigate(`/v2/${clusterName}/jobs/${event.data.job_id}`)
              }
            }}
          />
        </Box>
      )}
    </VStack>
  )
}
