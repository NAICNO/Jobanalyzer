import { Box, VStack, HStack, Spinner, Text, Alert, IconButton, NativeSelect, Pagination } from '@chakra-ui/react'
import { useMemo, useState } from 'react'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'
import { LuChevronLeft, LuChevronRight } from 'react-icons/lu'

import type { NodeInfoResponse } from '../../client'
import { createNodeNameRenderer, createCpuRenderer, createMemoryRenderer, createGpuRenderer } from '../../utils/nodeTableRenderers'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useClusterNodesInfoPages, useClusterNodesLastProbeTimestamp } from '../../hooks/v2/useNodeQueries'

interface NodesTableProps {
  clusterName: string
  onNodeClick?: (nodeName: string) => void
}

export const NodesTable = ({ clusterName, onNodeClick }: NodesTableProps) => {
  const client = useClusterClient(clusterName)

  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)

  const { data: nodesPageData, isLoading, isError } = useClusterNodesInfoPages({
    cluster: clusterName,
    client,
    page: currentPage,
    pageSize,
  })
  const { data: lastProbeData } = useClusterNodesLastProbeTimestamp({ cluster: clusterName, client })

  const rowData = useMemo<NodeInfoResponse[]>(() => {
    return nodesPageData?.nodes ?? []
  }, [nodesPageData])

  const totalNodes = nodesPageData?.total ?? 0

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
    <VStack w="100%" gap={0}>
      <Box w="100%" h="calc(100vh - 370px)" borderWidth="1px" borderColor="gray.200" rounded="md" bg="white">
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

      {/* Pagination Controls */}
      <HStack w="100%" justify="space-between" py={2}>
        <HStack gap={2}>
          <Text fontSize="sm" color="gray.600">Rows per page:</Text>
          <NativeSelect.Root size="sm" width="80px">
            <NativeSelect.Field
              value={String(pageSize)}
              onChange={(e) => {
                const newPageSize = Number(e.currentTarget.value)
                setPageSize(newPageSize)
                setCurrentPage(1)
              }}
            >
              <option value="25">25</option>
              <option value="50">50</option>
              <option value="100">100</option>
            </NativeSelect.Field>
            <NativeSelect.Indicator />
          </NativeSelect.Root>
          <Text fontSize="sm" color="gray.600">
            {totalNodes > 0
              ? `${(currentPage - 1) * pageSize + 1}–${Math.min(currentPage * pageSize, totalNodes)} of ${totalNodes}`
              : '0 nodes'}
          </Text>
        </HStack>

        <Pagination.Root
          count={totalNodes}
          pageSize={pageSize}
          page={currentPage}
          onPageChange={(e) => {
            setCurrentPage(e.page)
          }}
        >
          <HStack gap={2}>
            <Pagination.PrevTrigger asChild>
              <IconButton size="sm" variant="ghost">
                <LuChevronLeft />
              </IconButton>
            </Pagination.PrevTrigger>

            <Pagination.Items
              render={(page) => (
                <IconButton
                  key={page.value}
                  size="sm"
                  variant={page.value === currentPage ? 'solid' : 'ghost'}
                >
                  {page.value}
                </IconButton>
              )}
            />

            <Pagination.NextTrigger asChild>
              <IconButton size="sm" variant="ghost">
                <LuChevronRight />
              </IconButton>
            </Pagination.NextTrigger>
          </HStack>
        </Pagination.Root>
      </HStack>
    </VStack>
  )
}
