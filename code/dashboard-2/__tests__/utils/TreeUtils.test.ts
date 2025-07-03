import {
  parseCSV,
  buildPidToLabel,
  generateEdges,
  formatEdges
} from '../../src/util/TreeUtils.ts'
import { Edge } from '@xyflow/react'

describe('TreeUtils', () => {
  const sampleCSV = `
    1,0,init
    2,1,systemd
    3,1,systemd
    10,2,sshd
    11,2,nginx
    12,2,nginx
    20,3,python
    21,3,python
    22,3,node
    30,22,child_proc
    31,22,child_proc
    31,22,child_proc
    40,0,orphan
    41,40,child
    42,41,subchild
  `
  it('parseCSV returns ProcessEntry array', () => {
    const result = parseCSV(sampleCSV)
    expect(result).toEqual([
      { pid: 1, ppid: 0, command: 'init' },
      { pid: 2, ppid: 1, command: 'systemd' },
      { pid: 3, ppid: 1, command: 'systemd' },
      { pid: 10, ppid: 2, command: 'sshd' },
      { pid: 11, ppid: 2, command: 'nginx' },
      { pid: 12, ppid: 2, command: 'nginx' },
      { pid: 20, ppid: 3, command: 'python' },
      { pid: 21, ppid: 3, command: 'python' },
      { pid: 22, ppid: 3, command: 'node' },
      { pid: 30, ppid: 22, command: 'child_proc' },
      { pid: 31, ppid: 22, command: 'child_proc' },
      { pid: 40, ppid: 0, command: 'orphan' },
      { pid: 41, ppid: 40, command: 'child' },
      { pid: 42, ppid: 41, command: 'subchild' }
    ])
  })

  it('buildPidToLabel maps pid and ppid with labels', () => {
    const entries = parseCSV(sampleCSV)
    const labelMap = buildPidToLabel(entries)
    expect(labelMap.get(1)).toBe('1:init')
    expect(labelMap.get(2)).toBe('2:systemd')
    expect(labelMap.get(3)).toBe('3:systemd')
    expect(labelMap.get(10)).toBe('10:sshd')
    expect(labelMap.get(11)).toBe('11:nginx')
    expect(labelMap.get(22)).toBe('22:node')
    expect(labelMap.get(31)).toBe('31:child_proc')
    expect(labelMap.get(0)).toBe('0:unknown') // implied parent
  })

  it('generateEdges returns [parent, child] label pairs', () => {
    const entries = parseCSV(sampleCSV)
    const labelMap = buildPidToLabel(entries)
    const edges = generateEdges(entries, labelMap)
    expect(edges).toContainEqual(['1:init', '2:systemd'])
    expect(edges).toContainEqual(['2:systemd', '10:sshd'])
    expect(edges).toContainEqual(['3:systemd', '20:python'])
    expect(edges).toContainEqual(['22:node', '30:child_proc'])
    expect(edges).toContainEqual(['40:orphan', '41:child'])
    expect(edges).toContainEqual(['41:child', '42:subchild'])
  })

  it('formatEdges returns correctly structured Edge array', () => {
    const rawEdges: [string, string][] = [
      ['22:node', '30:child_proc']
    ]
    const edges: Edge[] = formatEdges(rawEdges)
    expect(edges[0].id).toBe('22:node-30:child_proc')
    expect(edges[0].source).toBe('22:node')
    expect(edges[0].target).toBe('30:child_proc')

  })
})




