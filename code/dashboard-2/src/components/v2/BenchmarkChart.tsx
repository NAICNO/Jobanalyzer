import { useMemo } from 'react'
import { Box } from '@chakra-ui/react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import type { BenchmarkRecord, BenchmarkFilterState } from '../../types/benchmark'

const BENCHMARK_COLORS = [
  '#4299E1', '#E53E3E', '#805AD5', '#ED8936', '#718096',
  '#D53F8C', '#38A169', '#A0AEC0', '#D69E2E', '#319795',
  '#2B6CB0',
]

export const getTestColorMap = (tests: string[]): Record<string, string> => {
  return Object.fromEntries(
    tests.map((t, i) => [t, BENCHMARK_COLORS[i % BENCHMARK_COLORS.length]]),
  )
}

interface BenchmarkChartRow {
  system: string
  [taskName: string]: number | string
}

interface BenchmarkChartProps {
  records: BenchmarkRecord[]
  filters: BenchmarkFilterState
  colorMap: Record<string, string>
  visibleTests: string[]
}

export const BenchmarkChart = ({
  records,
  filters,
  colorMap,
  visibleTests,
}: BenchmarkChartProps) => {
  const chartData = useMemo(() => {
    // 1. Filter records
    const filtered = records.filter((r) => {
      if (r.metric_name !== filters.metric) return false
      if (filters.selectedPrecisions.length > 0 && !filters.selectedPrecisions.includes(r.precision)) return false
      if (filters.selectedGpuCounts.length > 0 && !filters.selectedGpuCounts.includes(r.number_of_gpus)) return false
      if (String(r.exit_code) !== 'SUCCESS') return false
      return true
    })

    // 2. Group by system key (system + precision) → task_name, keep latest by end_time
    const grouped = new Map<string, Map<string, BenchmarkRecord>>()
    for (const r of filtered) {
      const systemKey = `${r.system} (${r.precision})`
      if (!grouped.has(systemKey)) {
        grouped.set(systemKey, new Map())
      }
      const taskMap = grouped.get(systemKey)!
      const existing = taskMap.get(r.task_name)
      if (!existing || r.end_time > existing.end_time) {
        taskMap.set(r.task_name, r)
      }
    }

    // 3. Apply system filter
    let systemKeys = [...grouped.keys()]
    if (filters.systemFilter) {
      const filterLower = filters.systemFilter.toLowerCase()
      systemKeys = systemKeys.filter((s) => s.toLowerCase().includes(filterLower))
    }

    // 4. Build reference values for relative mode
    let refValues: Map<string, number> | null = null
    if (filters.comparisonType === 'relative' && filters.referenceSystem) {
      const refMap = grouped.get(filters.referenceSystem)
      if (refMap) {
        refValues = new Map()
        for (const [task, record] of refMap) {
          if (record.metric_value > 0) {
            refValues.set(task, record.metric_value)
          }
        }
      }
    }

    // If relative mode but no valid reference data, return empty
    if (filters.comparisonType === 'relative' && !refValues) {
      return []
    }

    // 5. Build chart rows
    const rows: BenchmarkChartRow[] = systemKeys.map((systemKey) => {
      const taskMap = grouped.get(systemKey)!
      const row: BenchmarkChartRow = { system: systemKey }
      for (const test of visibleTests) {
        const record = taskMap.get(test)
        if (record) {
          if (filters.comparisonType === 'relative' && refValues) {
            const refVal = refValues.get(test)
            if (refVal) {
              row[test] = record.metric_value / refVal
            }
          } else {
            row[test] = record.metric_value
          }
        }
      }
      return row
    })

    // 6. Sort by average performance descending
    rows.sort((a, b) => {
      const avgA =
        visibleTests.reduce((sum, t) => sum + (typeof a[t] === 'number' ? (a[t] as number) : 0), 0) /
        visibleTests.length
      const avgB =
        visibleTests.reduce((sum, t) => sum + (typeof b[t] === 'number' ? (b[t] as number) : 0), 0) /
        visibleTests.length
      return avgB - avgA
    })

    return rows
  }, [records, filters, visibleTests])

  const chartHeight = Math.max(
    400,
    chartData.length * (visibleTests.length * 8 + 40) + 100,
  )

  if (chartData.length === 0) {
    return (
      <Box p={4} textAlign="center" color="fg.muted">
        No benchmark data matches the selected filters.
      </Box>
    )
  }

  return (
    <Box w="100%" overflowX="auto">
      <ResponsiveContainer width="100%" height={chartHeight} minWidth={800}>
        <BarChart layout="vertical" data={chartData} margin={{ left: 0, right: 30, top: 10, bottom: 10 }}>
          <CartesianGrid strokeDasharray="3 3" horizontal={false} />
          <XAxis type="number" />
          <YAxis
            dataKey="system"
            type="category"
            width={380}
            tick={{ fontSize: 11 }}
          />
          <Tooltip
            formatter={(value, name) => {
              const num = typeof value === 'number' ? value : 0
              return [
                filters.comparisonType === 'relative'
                  ? `${num.toFixed(2)}x`
                  : num.toLocaleString(),
                name,
              ]
            }}
          />
          {filters.comparisonType === 'relative' && (
            <ReferenceLine x={1} stroke="#666" strokeDasharray="3 3" />
          )}
          {visibleTests.map((test) => (
            <Bar
              key={test}
              dataKey={test}
              fill={colorMap[test]}
              barSize={10}
            />
          ))}
        </BarChart>
      </ResponsiveContainer>
    </Box>
  )
}
