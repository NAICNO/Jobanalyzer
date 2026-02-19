import { useMemo } from 'react'
import { VStack, Text, HStack, Spinner, Alert, Tag } from '@chakra-ui/react'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ICellRendererParams } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import type { NodeStateResponse } from '../../client'
import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeStates } from '../../hooks/v2/useNodeQueries'

interface Props {
  cluster: string
  nodename: string
}

export const NodeStates = ({ cluster, nodename }: Props) => {
  const client = useClusterClient(cluster)

  const { data, isLoading, isError, error } = useNodeStates({ cluster, nodename, client })

  const items = useMemo<NodeStateResponse[]>(() => (Array.isArray(data) ? (data as NodeStateResponse[]) : []), [data])

  const columnDefs = useMemo<ColDef<NodeStateResponse>[]>(() => [
    {
      headerName: 'States',
      field: 'states',
      width: 300,
      cellRenderer: (params: ICellRendererParams<NodeStateResponse>) => {
        const states = params.value as string[] | undefined
        if (!states?.length) return null
        return (
          <HStack wrap="wrap" gap={1} h="100%" align="center">
            {states.map((s, i) => (
              <Tag.Root key={i} colorPalette="blue" variant="surface" size="sm">
                <Tag.Label>{s}</Tag.Label>
              </Tag.Root>
            ))}
          </HStack>
        )
      },
    },
    {
      headerName: 'Time',
      field: 'time',
      flex: 1,
      minWidth: 180,
      valueFormatter: (params) => params.value?.toLocaleString() ?? '',
    },
  ], [])

  const defaultColDef = useMemo<ColDef>(() => ({
    sortable: true,
    resizable: true,
  }), [])

  if (isLoading) {
    return (
      <HStack>
        <Spinner size="sm" />
        <Text>Loading states…</Text>
      </HStack>
    )
  }

  if (isError) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Description>
          {error instanceof Error ? error.message : 'Failed to load states.'}
        </Alert.Description>
      </Alert.Root>
    )
  }

  if (items.length === 0) {
    return <Text color="gray.600">No states available.</Text>
  }

  return (
    <VStack align="start" w="100%" gap={2}>
      <Text fontWeight="semibold">States</Text>
      <div style={{ width: '100%', height: `${Math.min(items.length * 42 + 49, 300)}px` }}>
        <AgGridReact<NodeStateResponse>
          theme={themeQuartz}
          rowData={items}
          columnDefs={columnDefs}
          defaultColDef={defaultColDef}
          domLayout="normal"
          suppressCellFocus
        />
      </div>
    </VStack>
  )
}
