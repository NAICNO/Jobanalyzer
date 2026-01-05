import { Box, VStack, Spinner, Text, Alert } from '@chakra-ui/react'
import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import { getClusterByClusterNodesInfoOptions, getClusterByClusterNodesLastProbeTimestampOptions } from '../../client/@tanstack/react-query.gen'
import type { NodeInfoResponse } from '../../client'
import { createNodeNameRenderer, createCpuRenderer, createMemoryRenderer, createGpuRenderer } from '../../utils/nodeTableRenderers'

interface NodesTableProps {
  clusterName: string
  onNodeClick?: (nodeName: string) => void
}

export const NodesTable = ({ clusterName, onNodeClick }: NodesTableProps) => {
  // Fetch node info data
  const { data: nodesInfoData, isLoading, isError } = useQuery({
    ...getClusterByClusterNodesInfoOptions({
      path: { cluster: clusterName },
    }),
    enabled: !!clusterName,
  })

  // Fetch last probe timestamps for liveness
  const { data: lastProbeData } = useQuery({
    ...getClusterByClusterNodesLastProbeTimestampOptions({
      path: { cluster: clusterName },
    }),
    enabled: !!clusterName,
  })

  // Transform data to array format
  const rowData = useMemo<NodeInfoResponse[]>(() => {
    if (!nodesInfoData) return []
    
    const nodesInfo = nodesInfoData as Record<string, NodeInfoResponse>
    return Object.values(nodesInfo)
  }, [nodesInfoData])

  // Cell renderers using utility functions
  const nodeNameCellRenderer = useMemo(
    () => createNodeNameRenderer(lastProbeData as Record<string, Date | null> | undefined),
    [lastProbeData]
  )

  const cpuCellRenderer = useMemo(() => createCpuRenderer(), [])
  const memoryCellRenderer = useMemo(() => createMemoryRenderer(), [])
  const gpuCellRenderer = useMemo(() => createGpuRenderer(), [])

  // Column definitions
  const columnDefs = useMemo<ColDef<NodeInfoResponse>[]>(
    () => [
      { 
        field: 'node', 
        headerName: 'Node', 
        width: 180,
        sortable: true,
        filter: true,
        pinned: 'left',
        cellRenderer: nodeNameCellRenderer
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'architecture', 
        headerName: 'Arch', 
        width: 100,
        sortable: true,
        filter: true,
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'cpu_model', 
        headerName: 'CPU Model', 
        flex: 1,
        minWidth: 200,
        sortable: true,
        filter: true,
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'sockets', 
        headerName: 'Sockets', 
        width: 90,
        sortable: true,
        type: 'numericColumn',
        cellStyle: { textAlign: 'center' }
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'cores_per_socket', 
        headerName: 'Cores/Socket', 
        width: 130,
        sortable: true,
        type: 'numericColumn',
        cellStyle: { textAlign: 'center' }
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'threads_per_core', 
        headerName: 'Threads/Core', 
        width: 130,
        sortable: true,
        type: 'numericColumn',
        cellStyle: { textAlign: 'center' }
      } as ColDef<NodeInfoResponse>,
      { 
        headerName: 'CPU\n(avail/res)', 
        width: 120,
        sortable: true,
        type: 'numericColumn',
        cellRenderer: cpuCellRenderer,
        valueGetter: (params) => {
          const data = params.data
          if (!data) return 0
          return data.sockets * data.cores_per_socket * data.threads_per_core
        },
        cellStyle: { textAlign: 'center' }
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'memory', 
        headerName: 'Memory GiB\n(avail/res)', 
        width: 140,
        sortable: true,
        type: 'numericColumn',
        cellRenderer: memoryCellRenderer,
        cellStyle: { textAlign: 'center' }
      } as ColDef<NodeInfoResponse>,
      { 
        headerName: 'GPU\n(avail/res)', 
        width: 120,
        sortable: true,
        type: 'numericColumn',
        cellRenderer: gpuCellRenderer,
        valueGetter: (params) => params.data?.cards?.length || 0,
        cellStyle: { textAlign: 'center' }
      } as ColDef<NodeInfoResponse>,
      { 
        headerName: 'GPU Model', 
        width: 180,
        sortable: true,
        filter: true,
        valueGetter: (params) => {
          const cards = params.data?.cards
          if (!cards || cards.length === 0) return 'N/A'
          // Get unique GPU models
          const models = [...new Set(cards.map(card => card.model))]
          return models.join(', ')
        },
      } as ColDef<NodeInfoResponse>,
      { 
        headerName: 'Partitions', 
        width: 200,
        sortable: true,
        filter: true,
        valueGetter: (params) => params.data?.partitions?.join(', ') || 'N/A',
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'os_name', 
        headerName: 'OS', 
        width: 100,
        sortable: true,
        filter: true,
      } as ColDef<NodeInfoResponse>,
      { 
        field: 'os_release', 
        headerName: 'OS Release', 
        width: 150,
        sortable: true,
        filter: true,
      } as ColDef<NodeInfoResponse>,
    ],
    [nodeNameCellRenderer, cpuCellRenderer, memoryCellRenderer, gpuCellRenderer]
  )

  // Default column configuration
  const defaultColDef = useMemo<ColDef<NodeInfoResponse>>(
    () => ({
      resizable: true,
      sortable: true,
      wrapHeaderText: true,
      autoHeaderHeight: true,
    }),
    []
  )

  // Row styling for hover effect to indicate clickability
  const getRowStyle = useMemo(
    () => () => ({ cursor: 'pointer' }),
    []
  )

  if (isLoading) {
    return (
      <Box w="100%" h="calc(100vh - 320px)" borderWidth="1px" borderColor="gray.200" rounded="md" bg="white" display="flex" alignItems="center" justifyContent="center">
        <VStack gap={2}>
          <Spinner size="lg" />
          <Text>Loading node information...</Text>
        </VStack>
      </Box>
    )
  }

  if (isError) {
    return (
      <Box w="100%" h="calc(100vh - 320px)" borderWidth="1px" borderColor="gray.200" rounded="md" bg="white" p={4}>
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Failed to load node information.</Alert.Description>
        </Alert.Root>
      </Box>
    )
  }

  return (
    <Box w="100%" h="calc(100vh - 320px)" borderWidth="1px" borderColor="gray.200" rounded="md" bg="white">
      <AgGridReact<NodeInfoResponse>
        theme={themeQuartz}
        rowData={rowData}
        columnDefs={columnDefs}
        defaultColDef={defaultColDef}
        getRowStyle={getRowStyle}
        domLayout="normal"
        suppressCellFocus
        loading={isLoading}
        onRowClicked={(event) => {
          const nodeName = event.data?.node
          if (nodeName && onNodeClick) {
            onNodeClick(nodeName)
          }
        }}
      />
    </Box>
  )
}
