/**
 * Cell renderers for the nodes table
 */

import type { ICellRendererParams } from 'ag-grid-community'
import { TbBroadcast, TbBroadcastOff } from 'react-icons/tb'

import type { NodeInfoResponse } from '../client'
import { calculateLiveness } from './nodeLiveness'
import { getUtilizationColor, calculateUtilization } from './utilizationColors'

/**
 * Create a cell renderer for node names with liveness indicators
 */
export const createNodeNameRenderer = (lastProbeData: Record<string, Date | null> | undefined) => {
  const NodeNameCell = (params: ICellRendererParams<NodeInfoResponse>) => {
    const data = params.data
    if (!data) return <span>N/A</span>
    
    const lastProbeMap = (lastProbeData ?? {}) as Record<string, Date | null>
    const lastProbeTimestamp = lastProbeMap[data.node]
    const liveness = calculateLiveness(lastProbeTimestamp)
    
    // Choose icon based on liveness status
    const IconComponent = liveness.icon === 'offline' || liveness.icon === 'unknown' 
      ? TbBroadcastOff 
      : TbBroadcast
    
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <IconComponent 
          title={liveness.ageText}
          style={{ 
            color: liveness.color, 
            fontSize: '1.2em',
            flexShrink: 0
          }}
        />
        <span style={{ fontWeight: 500 }}>{data.node}</span>
      </div>
    )
  }
  NodeNameCell.displayName = 'NodeNameCell'
  return NodeNameCell
}

/**
 * Create a cell renderer for CPU allocation (available / reserved)
 */
export const createCpuRenderer = () => {
  const CpuCell = (params: ICellRendererParams<NodeInfoResponse>) => {
    const data = params.data
    if (!data) return <span>N/A</span>
    
    const totalCpus = data.sockets * data.cores_per_socket * data.threads_per_core
    const reservedCpus = data.alloc_tres?.cpu || 0
    const availableCpus = totalCpus - reservedCpus
    const utilizationPct = calculateUtilization(reservedCpus, totalCpus)
    
    const color = getUtilizationColor(utilizationPct)
    
    return (
      <span style={{ color, fontWeight: 500 }}>
        {availableCpus} / {reservedCpus}
      </span>
    )
  }
  CpuCell.displayName = 'CpuCell'
  return CpuCell
}

/**
 * Create a cell renderer for memory allocation (available / reserved)
 */
export const createMemoryRenderer = () => {
  const MemoryCell = (params: ICellRendererParams<NodeInfoResponse>) => {
    const data = params.data
    if (!data) return <span>N/A</span>
    
    const totalMemoryGiB = Math.round(data.memory / 1024 / 1024)
    const reservedMemoryGiB = data.alloc_tres?.memory ? Math.round(data.alloc_tres.memory / 1024 / 1024) : 0
    const availableMemoryGiB = totalMemoryGiB - reservedMemoryGiB
    const utilizationPct = calculateUtilization(reservedMemoryGiB, totalMemoryGiB)
    
    const color = getUtilizationColor(utilizationPct)
    
    return (
      <span style={{ color, fontWeight: 500 }}>
        {availableMemoryGiB} / {reservedMemoryGiB}
      </span>
    )
  }
  MemoryCell.displayName = 'MemoryCell'
  return MemoryCell
}

/**
 * Create a cell renderer for GPU allocation (available / reserved)
 */
export const createGpuRenderer = () => {
  const GpuCell = (params: ICellRendererParams<NodeInfoResponse>) => {
    const data = params.data
    if (!data) return <span>N/A</span>
    
    const totalGpus = data.cards?.length || 0
    const reservedGpus = data.alloc_tres?.gpu || 0
    const availableGpus = totalGpus - reservedGpus
    
    if (totalGpus === 0) return <span>0 / 0</span>
    
    const utilizationPct = calculateUtilization(reservedGpus, totalGpus)
    const color = getUtilizationColor(utilizationPct)
    
    return (
      <span style={{ color, fontWeight: 500 }}>
        {availableGpus} / {reservedGpus}
      </span>
    )
  }
  GpuCell.displayName = 'GpuCell'
  return GpuCell
}
