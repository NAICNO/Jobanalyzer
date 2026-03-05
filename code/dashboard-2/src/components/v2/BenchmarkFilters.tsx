import { useMemo } from 'react'
import {
  HStack,
  VStack,
  Text,
  Input,
  Field,
  Select,
  createListCollection,
  Portal,
} from '@chakra-ui/react'
import type { BenchmarkRecord, BenchmarkFilterState } from '../../types/benchmark'

interface BenchmarkFiltersProps {
  records: BenchmarkRecord[]
  filters: BenchmarkFilterState
  onFiltersChange: (filters: BenchmarkFilterState) => void
}

export const BenchmarkFilters = ({ records, filters, onFiltersChange }: BenchmarkFiltersProps) => {
  const metricOptions = useMemo(() => {
    const metrics = [...new Set(records.map((r) => r.metric_name))].sort()
    return createListCollection({
      items: metrics.map((m) => ({ label: m, value: m })),
    })
  }, [records])

  const precisionOptions = useMemo(() => {
    const precisions = [...new Set(records.map((r) => r.precision))].sort()
    return createListCollection({
      items: precisions.map((p) => ({ label: p, value: p })),
    })
  }, [records])

  const gpuCountOptions = useMemo(() => {
    const counts = [...new Set(records.map((r) => r.number_of_gpus))].sort((a, b) => a - b)
    return createListCollection({
      items: counts.map((c) => ({ label: String(c), value: String(c) })),
    })
  }, [records])

  const testOptions = useMemo(() => {
    const tests = [...new Set(records.map((r) => r.task_name))].sort()
    return createListCollection({
      items: tests.map((t) => ({ label: t, value: t })),
    })
  }, [records])

  const referenceSystemOptions = useMemo(() => {
    const filtered = records.filter((r) => {
      if (r.metric_name !== filters.metric) return false
      if (filters.selectedPrecisions.length > 0 && !filters.selectedPrecisions.includes(r.precision)) return false
      if (filters.selectedGpuCounts.length > 0 && !filters.selectedGpuCounts.includes(r.number_of_gpus)) return false
      return String(r.exit_code) === 'SUCCESS'
    })
    const systems = [...new Set(filtered.map((r) => `${r.system} (${r.precision})`))]
      .sort()
    return createListCollection({
      items: systems.map((s) => ({ label: s, value: s })),
    })
  }, [records, filters.metric, filters.selectedPrecisions, filters.selectedGpuCounts])

  const comparisonTypeOptions = useMemo(
    () =>
      createListCollection({
        items: [
          { label: 'relative', value: 'relative' },
          { label: 'absolute', value: 'absolute' },
        ],
      }),
    [],
  )

  const update = (partial: Partial<BenchmarkFilterState>) => {
    onFiltersChange({ ...filters, ...partial })
  }

  return (
    <VStack align="start" gap={3} w="100%">
      <HStack gap={4} flexWrap="wrap" w="100%">
        {/* Metric */}
        <Field.Root width="auto" minW="140px">
          <Field.Label fontSize="sm">Metric:</Field.Label>
          <Select.Root
            collection={metricOptions}
            value={[filters.metric]}
            onValueChange={(details) => update({ metric: details.value[0] })}
            size="sm"
          >
            <Select.HiddenSelect />
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="Select metric" />
              </Select.Trigger>
              <Select.IndicatorGroup>
                <Select.Indicator />
              </Select.IndicatorGroup>
            </Select.Control>
            <Portal>
              <Select.Positioner>
                <Select.Content>
                  {metricOptions.items.map((item) => (
                    <Select.Item item={item} key={item.value}>
                      {item.label}
                      <Select.ItemIndicator />
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select.Positioner>
            </Portal>
          </Select.Root>
        </Field.Root>

        {/* Precision (multi-select) */}
        <Field.Root width="auto" minW="140px" maxW="200px">
          <Field.Label fontSize="sm">Precision:</Field.Label>
          <Select.Root
            multiple
            collection={precisionOptions}
            value={filters.selectedPrecisions}
            onValueChange={(details) =>
              update({ selectedPrecisions: details.value })
            }
            size="sm"
          >
            <Select.HiddenSelect />
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="Select precision" />
              </Select.Trigger>
              <Select.IndicatorGroup>
                {filters.selectedPrecisions.length > 0 && (
                  <Select.ClearTrigger />
                )}
                <Select.Indicator />
              </Select.IndicatorGroup>
            </Select.Control>
            <Portal>
              <Select.Positioner>
                <Select.Content>
                  {precisionOptions.items.map((item) => (
                    <Select.Item item={item} key={item.value}>
                      {item.label}
                      <Select.ItemIndicator />
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select.Positioner>
            </Portal>
          </Select.Root>
        </Field.Root>

        {/* GPU Count (multi-select) */}
        <Field.Root width="auto" minW="120px" maxW="180px">
          <Field.Label fontSize="sm">GPU Count:</Field.Label>
          <Select.Root
            multiple
            collection={gpuCountOptions}
            value={filters.selectedGpuCounts.map(String)}
            onValueChange={(details) =>
              update({ selectedGpuCounts: details.value.map(Number) })
            }
            size="sm"
          >
            <Select.HiddenSelect />
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="Select GPU count" />
              </Select.Trigger>
              <Select.IndicatorGroup>
                {filters.selectedGpuCounts.length > 0 && (
                  <Select.ClearTrigger />
                )}
                <Select.Indicator />
              </Select.IndicatorGroup>
            </Select.Control>
            <Portal>
              <Select.Positioner>
                <Select.Content>
                  {gpuCountOptions.items.map((item) => (
                    <Select.Item item={item} key={item.value}>
                      {item.label}
                      <Select.ItemIndicator />
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select.Positioner>
            </Portal>
          </Select.Root>
        </Field.Root>

        {/* Tests (multi-select) */}
        <Field.Root width="auto" minW="180px" maxW="220px">
          <Field.Label fontSize="sm">Tests:</Field.Label>
          <Select.Root
            multiple
            collection={testOptions}
            value={filters.selectedTests}
            onValueChange={(details) => update({ selectedTests: details.value })}
            size="sm"
          >
            <Select.HiddenSelect />
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="Select test" />
              </Select.Trigger>
              <Select.IndicatorGroup>
                {filters.selectedTests.length > 0 && (
                  <Select.ClearTrigger />
                )}
                <Select.Indicator />
              </Select.IndicatorGroup>
            </Select.Control>
            <Portal>
              <Select.Positioner>
                <Select.Content>
                  {testOptions.items.map((item) => (
                    <Select.Item item={item} key={item.value}>
                      {item.label}
                      <Select.ItemIndicator />
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select.Positioner>
            </Portal>
          </Select.Root>
        </Field.Root>

        {/* Reference System */}
        <Field.Root width="auto" minW="240px" maxW="320px">
          <Field.Label fontSize="sm">Reference system:</Field.Label>
          <Select.Root
            collection={referenceSystemOptions}
            value={filters.referenceSystem ? [filters.referenceSystem] : []}
            onValueChange={(details) =>
              update({ referenceSystem: details.value[0] || '' })
            }
            size="sm"
          >
            <Select.HiddenSelect />
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="Select reference" />
              </Select.Trigger>
              <Select.IndicatorGroup>
                <Select.Indicator />
              </Select.IndicatorGroup>
            </Select.Control>
            <Portal>
              <Select.Positioner>
                <Select.Content>
                  {referenceSystemOptions.items.map((item) => (
                    <Select.Item item={item} key={item.value}>
                      {item.label}
                      <Select.ItemIndicator />
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select.Positioner>
            </Portal>
          </Select.Root>
        </Field.Root>

        {/* Comparison Type */}
        <Field.Root width="auto" minW="180px">
          <Field.Label fontSize="sm">Comparison Type:</Field.Label>
          <Select.Root
            collection={comparisonTypeOptions}
            value={[filters.comparisonType]}
            onValueChange={(details) =>
              update({
                comparisonType: (details.value[0] as 'relative' | 'absolute') || 'relative',
              })
            }
            size="sm"
          >
            <Select.HiddenSelect />
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="Select comparison type" />
              </Select.Trigger>
              <Select.IndicatorGroup>
                <Select.Indicator />
              </Select.IndicatorGroup>
            </Select.Control>
            <Portal>
              <Select.Positioner>
                <Select.Content>
                  {comparisonTypeOptions.items.map((item) => (
                    <Select.Item item={item} key={item.value}>
                      {item.label}
                      <Select.ItemIndicator />
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select.Positioner>
            </Portal>
          </Select.Root>
        </Field.Root>
      </HStack>

      {/* System Filter */}
      <Field.Root maxW="360px">
        <Field.Label fontSize="sm">System Filter</Field.Label>
        <Input
          size="sm"
          placeholder="Set filter pattern"
          value={filters.systemFilter}
          onChange={(e) => update({ systemFilter: e.target.value })}
        />
        <Text fontSize="xs" color="fg.muted">
          Use this text filter to narrow number of systems
        </Text>
      </Field.Root>
    </VStack>
  )
}
