import { useMemo } from 'react'
import { VStack, Text, HStack, Spinner, Alert } from '@chakra-ui/react'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, SizeColumnsToFitGridStrategy, ICellRendererParams } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeInfo, useNodeProcessGpuUtil } from '../../hooks/v2/useNodeQueries'
import type { NodeInfoResponse, GpuCardResponse, SampleProcessGpuAccResponse } from '../../client'

interface Props {
  cluster: string
  nodename: string
}

interface GpuRow extends GpuCardResponse {
  compute_util?: number
  memory_util_live?: number
}

function UtilCell({ compute, memory }: { compute?: number; memory?: number }) {
  if (compute == null && memory == null) return <span style={{ fontSize: 11, color: '#a0aec0' }}>—</span>
  return (
    <div style={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
      <UtilRow label="CPU" value={compute} />
      <UtilRow label="Mem" value={memory} />
    </div>
  )
}

function UtilRow({ label, value }: { label: string; value?: number }) {
  if (value == null) return null
  const color = value > 80 ? '#e53e3e' : value > 50 ? '#d69e2e' : '#38a169'
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 4, width: '100%', height: 16 }}>
      <span style={{ fontSize: 10, color: '#718096', width: 24, flexShrink: 0 }}>{label}</span>
      <div style={{ flex: 1, height: 5, background: '#e2e8f0', borderRadius: 3, overflow: 'hidden' }}>
        <div style={{ width: `${Math.min(value, 100)}%`, height: '100%', background: color, borderRadius: 3 }} />
      </div>
      <span style={{ fontSize: 10, width: 28, textAlign: 'right', flexShrink: 0 }}>{value.toFixed(0)}%</span>
    </div>
  )
}

export const NodeGpuTable = ({ cluster, nodename }: Props) => {
  const client = useClusterClient(cluster)

  const { data: nodeInfoData, isLoading, isError, error } = useNodeInfo({ cluster, nodename, client })
  const { data: gpuUtilData } = useNodeProcessGpuUtil({ cluster, nodename, client })

  const info: NodeInfoResponse | undefined = useMemo(() => {
    if (!nodeInfoData) return undefined
    const map = nodeInfoData as Record<string, NodeInfoResponse>
    return map[nodename] ?? Object.values(map)[0]
  }, [nodeInfoData, nodename])

  // Build a map of UUID → utilization from process GPU util data
  const utilByUuid = useMemo<Record<string, SampleProcessGpuAccResponse>>(() => {
    if (!gpuUtilData) return {}
    const nodeMap = gpuUtilData as Record<string, Record<string, SampleProcessGpuAccResponse>>
    const gpuMap = nodeMap[nodename] ?? nodeMap[Object.keys(nodeMap)[0]]
    return gpuMap ?? {}
  }, [gpuUtilData, nodename])

  // Merge card info with live utilization
  const rows = useMemo<GpuRow[]>(() => {
    const cards = info?.cards
    if (!Array.isArray(cards)) return []
    return cards.map((card) => {
      const util = card.uuid ? utilByUuid[card.uuid] : undefined
      return {
        ...card,
        compute_util: util?.gpu_util,
        memory_util_live: util?.gpu_memory_util,
      }
    })
  }, [info?.cards, utilByUuid])

  const columnDefs = useMemo<ColDef<GpuRow>[]>(() => [
    { field: 'index', headerName: '#', width: 60 },
    {
      headerName: 'Model',
      valueGetter: (params) => `${params.data?.manufacturer} ${params.data?.model}`,
      minWidth: 300,
      flex: 1,
    },
    {
      headerName: 'Utilization',
      width: 200,
      cellStyle: { display: 'flex', alignItems: 'center' },
      cellRenderer: (params: ICellRendererParams<GpuRow>) => (
        <UtilCell compute={params.data?.compute_util} memory={params.data?.memory_util_live} />
      ),
    },
    { field: 'architecture', headerName: 'Arch', width: 100 },
    { field: 'address', headerName: 'Address', width: 150 },
    { field: 'driver', headerName: 'Driver', width: 110 },
    { field: 'firmware', headerName: 'Firmware', width: 120 },
    {
      headerName: 'Power (W)',
      valueGetter: (params) => `${params.data?.min_power_limit}–${params.data?.max_power_limit}`,
      width: 120,
    },
    {
      headerName: 'Max Clocks',
      valueGetter: (params) => `CE ${params.data?.max_ce_clock} · Mem ${params.data?.max_memory_clock} MHz`,
      width: 220,
    },
    { field: 'uuid', headerName: 'UUID', width: 320 },
  ], [])

  const defaultColDef = useMemo<ColDef>(() => ({
    sortable: true,
    resizable: true,
  }), [])

  const autoSizeStrategy = useMemo<SizeColumnsToFitGridStrategy>(() => ({
    type: 'fitGridWidth',
    defaultMinWidth: 80,
  }), [])

  if (isLoading) {
    return (
      <HStack>
        <Spinner size="sm" />
        <Text>Loading GPU info…</Text>
      </HStack>
    )
  }

  if (isError) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Description>
          {error instanceof Error ? error.message : 'Failed to load GPU info.'}
        </Alert.Description>
      </Alert.Root>
    )
  }

  if (rows.length === 0) {
    return null
  }

  return (
    <VStack align="start" w="100%" gap={2}>
      <Text fontWeight="semibold">GPUs ({rows.length})</Text>
      <div style={{ width: '100%', height: `${Math.min(rows.length * 42 + 49, 400)}px` }}>
        <AgGridReact<GpuRow>
          theme={themeQuartz}
          rowData={rows}
          columnDefs={columnDefs}
          defaultColDef={defaultColDef}
          autoSizeStrategy={autoSizeStrategy}
          rowHeight={42}
          domLayout="normal"
          suppressCellFocus
          getRowId={(params) => params.data.uuid ?? String(params.data.index)}
        />
      </div>
    </VStack>
  )
}
