import { useState, useEffect, useMemo } from 'react'
import { useParams } from 'react-router'
import { VStack, Heading, Alert, Spinner, Box } from '@chakra-ui/react'

import { useClusterClient } from '../../hooks/useClusterClient'
import { useBenchmarks } from '../../hooks/v2/useBenchmarkQueries'
import { BenchmarkFilters } from '../../components/v2/BenchmarkFilters'
import { BenchmarkChart, getTestColorMap } from '../../components/v2/BenchmarkChart'
import { BenchmarkLegend } from '../../components/v2/BenchmarkLegend'
import type { BenchmarkFilterState } from '../../types/benchmark'

const DEFAULT_FILTERS: BenchmarkFilterState = {
  metric: 'throughput',
  selectedPrecisions: ['fp16'],
  selectedGpuCounts: [1],
  selectedTests: [],
  referenceSystem: '',
  comparisonType: 'relative',
  systemFilter: '',
}

export const BenchmarksPage = () => {
  const { clusterName } = useParams()
  const client = useClusterClient()
  const { data: records, isLoading, error } = useBenchmarks({
    cluster: clusterName || '',
    client,
  })

  const [filters, setFilters] = useState<BenchmarkFilterState>(DEFAULT_FILTERS)
  const [initialized, setInitialized] = useState(false)

  // Initialize defaults from data on first load
  useEffect(() => {
    if (!records || records.length === 0 || initialized) return

    const metrics = [...new Set(records.map((r) => r.metric_name))]
    const precisions = [...new Set(records.map((r) => r.precision))].sort()
    const gpuCounts = [...new Set(records.map((r) => r.number_of_gpus))].sort((a, b) => a - b)

    const defaultPrecisions = precisions.includes('fp16') ? ['fp16'] : precisions.length > 0 ? [precisions[0]] : []
    const defaultGpuCounts = gpuCounts.includes(1) ? [1] : gpuCounts.length > 0 ? [gpuCounts[0]] : []

    // Find first reference system matching the default precision + GPU count
    const matchingRecords = records.filter(
      (r) =>
        String(r.exit_code) === 'SUCCESS' &&
        (defaultPrecisions.length === 0 || defaultPrecisions.includes(r.precision)) &&
        (defaultGpuCounts.length === 0 || defaultGpuCounts.includes(r.number_of_gpus)),
    )
    const refSystems = [...new Set(matchingRecords.map((r) => `${r.system} (${r.precision})`))].sort()
    const defaultRef = refSystems.find((s) => s.includes('A100')) || refSystems[0] || ''

    setFilters({
      metric: metrics[0] || 'throughput',
      selectedPrecisions: defaultPrecisions,
      selectedGpuCounts: defaultGpuCounts,
      selectedTests: [],
      referenceSystem: defaultRef,
      comparisonType: 'relative',
      systemFilter: '',
    })
    setInitialized(true)
  }, [records, initialized])

  // Auto-update reference system when precision or GPU count changes
  useEffect(() => {
    if (!records || records.length === 0 || !initialized) return

    const matchingRecords = records.filter(
      (r) =>
        String(r.exit_code) === 'SUCCESS' &&
        r.metric_name === filters.metric &&
        (filters.selectedPrecisions.length === 0 || filters.selectedPrecisions.includes(r.precision)) &&
        (filters.selectedGpuCounts.length === 0 || filters.selectedGpuCounts.includes(r.number_of_gpus)),
    )
    const availableRefSystems = [...new Set(matchingRecords.map((r) => `${r.system} (${r.precision})`))].sort()

    // If current reference system is still valid, keep it
    if (availableRefSystems.includes(filters.referenceSystem)) return

    // Otherwise pick the first matching system
    const newRef = availableRefSystems.find((s) => s.includes('A100')) || availableRefSystems[0] || ''
    if (newRef !== filters.referenceSystem) {
      setFilters((prev) => ({ ...prev, referenceSystem: newRef }))
    }
  }, [records, filters.metric, filters.selectedPrecisions, filters.selectedGpuCounts, filters.referenceSystem, initialized])

  // Compute visible tests and color map
  const allTests = useMemo(() => {
    if (!records) return []
    return [...new Set(records.map((r) => r.task_name))].sort()
  }, [records])

  const visibleTests = useMemo(() => {
    if (filters.selectedTests.length > 0) return filters.selectedTests
    return allTests
  }, [filters.selectedTests, allTests])

  const colorMap = useMemo(() => getTestColorMap(allTests), [allTests])

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

  if (isLoading) {
    return (
      <VStack p={8} align="center" justify="center" minH="400px">
        <Spinner size="xl" />
      </VStack>
    )
  }

  if (error) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>
            Failed to load benchmark data: {error.message}
          </Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  if (!records || records.length === 0) {
    return (
      <VStack p={4} align="start">
        <Heading size="lg">Benchmarks</Heading>
        <Alert.Root status="info">
          <Alert.Indicator />
          <Alert.Description>No benchmark data available for this cluster.</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <VStack align="start" w="100%" p={4} pt={2} gap={4}>
      <Heading size="lg">Benchmarks</Heading>
      <Heading size="md" fontWeight="normal">
        Lambdal (
        {filters.comparisonType === 'relative'
          ? `Relative comparison against reference: ${filters.referenceSystem || 'none'}`
          : 'Absolute values'}
        )
      </Heading>

      <BenchmarkFilters
        records={records}
        filters={filters}
        onFiltersChange={setFilters}
      />

      <BenchmarkLegend tests={visibleTests} colorMap={colorMap} />

      <Box w="100%">
        <BenchmarkChart
          records={records}
          filters={filters}
          colorMap={colorMap}
          visibleTests={visibleTests}
        />
      </Box>
    </VStack>
  )
}
