import { useMemo, lazy, Suspense, useState, useEffect } from 'react'
import { Box, Text, VStack, HStack, Spinner, Alert, Badge, SimpleGrid, Separator, Stat, Tabs, Tooltip, Icon } from '@chakra-ui/react'
import { useParams, useLocation } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HiInformationCircle } from 'react-icons/hi2'

import { getClusterByClusterJobsByJobIdOptions, getClusterByClusterJobsByJobIdReportOptions } from '../../client/@tanstack/react-query.gen'
import { useClusterClient } from '../../hooks/useClusterClient'
import { formatDuration, formatMemory, getJobStateColor } from '../../util/formatters'
import { JobStatCard } from '../../components/v2/JobStatCard'
import { NavigateBackButton } from '../../components/NavigateBackButton'

// Lazy load tab components for better initial load performance
const OverviewTab = lazy(() => import('../../components/v2/OverviewTab').then(m => ({ default: m.OverviewTab })))
const PerformanceMetricsTab = lazy(() => import('../../components/v2/PerformanceMetricsTab').then(m => ({ default: m.PerformanceMetricsTab })))
const ResourceTimelineTab = lazy(() => import('../../components/v2/ResourceTimelineTab').then(m => ({ default: m.ResourceTimelineTab })))
const GpuPerformanceTab = lazy(() => import('../../components/v2/GpuPerformanceTab').then(m => ({ default: m.GpuPerformanceTab })))

export const JobDetailsPage = () => {
  const { clusterName, jobId } = useParams<{ clusterName: string; jobId: string }>()
  const client = useClusterClient(clusterName)
  const location = useLocation()
  const [activeTab, setActiveTab] = useState('overview')

  const jobIdNum = useMemo(() => {
    const parsed = parseInt(jobId ?? '', 10)
    return isNaN(parsed) ? undefined : parsed
  }, [jobId])

  const { data: job, isLoading, isError, error, refetch } = useQuery({
    ...getClusterByClusterJobsByJobIdOptions({
      path: {
        cluster: clusterName ?? '',
        job_id: jobIdNum ?? 0
      },
      client: client ?? undefined,
    }),
    enabled: !!clusterName && !!jobIdNum && !!client,
    staleTime: 5 * 60 * 1000, // 5 minutes - job details rarely change
    gcTime: 10 * 60 * 1000, // 10 minutes cache
  })

  const { data: report, refetch: refetchReport } = useQuery({
    ...getClusterByClusterJobsByJobIdReportOptions({
      path: {
        cluster: clusterName ?? '',
        job_id: jobIdNum ?? 0
      },
      client: client ?? undefined,
    }),
    enabled: !!clusterName && !!jobIdNum && !!client,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes cache
  })

  const elapsed = useMemo(() => {
    if (!job?.start_time) return 0
    const start = new Date(job.start_time).getTime()
    const end = job.end_time ? new Date(job.end_time).getTime() : Date.now()
    return Math.floor((end - start) / 1000)
  }, [job?.start_time, job?.end_time])

  // Extract GPU count from resource strings
  const gpuInfo = useMemo(() => {
    const extractGpuCount = (resources: string | null | undefined): number => {
      if (!resources) return 0
      const match = resources.match(/gres\/gpu[=:](\d+)/)
      return match ? parseInt(match[1], 10) : 0
    }

    return {
      requested: extractGpuCount(job?.requested_resources),
      allocated: extractGpuCount(job?.allocated_resources),
      uuids: job?.used_gpu_uuids || [],
      gresDetail: job?.gres_detail
    }
  }, [job?.requested_resources, job?.allocated_resources, job?.used_gpu_uuids, job?.gres_detail])

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if user is typing in input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return
      }

      switch (e.key) {
      case '1':
        setActiveTab('overview')
        break
      case '2':
        setActiveTab('metrics')
        break
      case '3':
        setActiveTab('timeline')
        break
      case '4':
        // Only switch to GPU tab if GPUs are available
        if (gpuInfo.uuids.length > 0) {
          setActiveTab('gpu')
        }
        break
      case 'r':
      case 'R':
        // Refresh data
        void refetch()
        void refetchReport()
        break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [gpuInfo.uuids.length, refetch, refetchReport])

  // Render loading/error states after all hooks
  if (!client) {
    return (
      <VStack p={4} align="start">
        <Spinner />
      </VStack>
    )
  }

  if (!clusterName || !jobIdNum) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Invalid job ID or cluster name</Alert.Description>
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
            Failed to load job details: {error?.message || 'Unknown error'}
          </Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  if (isLoading) {
    return (
      <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
        <Spinner size="xl" />
      </Box>
    )
  }

  if (!job) {
    return (
      <VStack p={4} align="start">
        <Alert.Root status="warning">
          <Alert.Indicator />
          <Alert.Description>Job not found</Alert.Description>
        </Alert.Root>
      </VStack>
    )
  }

  return (
    <VStack w="100%" align="start" gap={0} p={4}>
      {/* Header */}
      <VStack align="start" w="100%">
        <HStack gap={3} justify="space-between" w="100%">
          <HStack gap={3}>
            {location.key !== 'default' && <NavigateBackButton />}
            <Text fontSize="3xl" fontWeight="bold">
              Job {job.job_id}
            </Text>
            <Badge colorPalette={getJobStateColor(job.job_state)} size="lg">
              {job.job_state}
            </Badge>
          </HStack>
          <Tooltip.Root openDelay={200}>
            <Tooltip.Trigger asChild>
              <Icon size="md" color="gray.500" cursor="help">
                <HiInformationCircle />
              </Icon>
            </Tooltip.Trigger>
            <Tooltip.Positioner>
              <Tooltip.Content maxW="300px">
                <VStack align="start" gap={2}>
                  <Text fontSize="xs" fontWeight="semibold">Keyboard Shortcuts</Text>
                  <Text fontSize="xs">• Press <Badge size="xs">1</Badge> for Overview</Text>
                  <Text fontSize="xs">• Press <Badge size="xs">2</Badge> for Performance Metrics</Text>
                  <Text fontSize="xs">• Press <Badge size="xs">3</Badge> for Resource Timeline</Text>
                  {gpuInfo.uuids.length > 0 && (
                    <Text fontSize="xs">• Press <Badge size="xs">4</Badge> for GPU Performance</Text>
                  )}
                  <Text fontSize="xs">• Press <Badge size="xs">R</Badge> to refresh data</Text>
                </VStack>
              </Tooltip.Content>
            </Tooltip.Positioner>
          </Tooltip.Root>
        </HStack>
        <Text fontSize="md" color="fg.muted">
          Cluster: {clusterName}
        </Text>
      </VStack>

      {/* Quick Stats - Sticky Header */}
      <Box
        position="sticky"
        top={0}
        zIndex={10}
        bg="bg"
        w="100%"
        borderY="1px"
        borderColor="border"
        py={4}
        mx={-6}
        px={6}
      >
        <SimpleGrid columns={{ base: 2, md: 3, lg: 5 }} gap={3}>
          <Box borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
            <Stat.Root>
              <Stat.Label fontSize="sm" color="gray.600">Status</Stat.Label>
              <Stat.ValueText>
                <Badge colorPalette={getJobStateColor(job.job_state)}>
                  {job.job_state}
                </Badge>
              </Stat.ValueText>
            </Stat.Root>
          </Box>

          <JobStatCard
            label="Elapsed Time"
            value={formatDuration(elapsed)}
            tooltip="Total time from job start to completion or current time for running jobs"
          />

          <JobStatCard
            label="CPU Hours"
            value={((elapsed / 3600) * (job.requested_cpus || 0)).toFixed(2)}
            tooltip="Total CPU time: Elapsed time × Requested CPUs. Measures total compute capacity allocated."
          />

          <JobStatCard
            label="Peak Memory"
            value={job.sacct?.MaxRSS ? formatMemory(job.sacct.MaxRSS) : 'N/A'}
            tooltip="Maximum resident memory used (MaxRSS from SLURM accounting). Shows actual memory footprint."
          />

          {gpuInfo.allocated > 0 && (
            <JobStatCard
              label="GPU Hours"
              value={((elapsed / 3600) * gpuInfo.allocated).toFixed(2)}
              tooltip="Total GPU time: Elapsed time × Allocated GPUs. Measures total GPU compute capacity used."
            />
          )}
        </SimpleGrid>
      </Box>

      <Separator />

      {/* Tabbed Content */}
      <Tabs.Root value={activeTab} onValueChange={(e) => setActiveTab(e.value)} w="100%">
        <Tabs.List>
          <Tabs.Trigger value="overview">
            <Text>Overview</Text>
            <Badge size="xs" ml={1} colorPalette="gray">1</Badge>
          </Tabs.Trigger>
          <Tabs.Trigger value="metrics">
            <Text>Performance Metrics</Text>
            <Badge size="xs" ml={1} colorPalette="gray">2</Badge>
          </Tabs.Trigger>
          <Tabs.Trigger value="timeline">
            <Text>Resource Timeline</Text>
            <Badge size="xs" ml={1} colorPalette="gray">3</Badge>
          </Tabs.Trigger>
          {gpuInfo.uuids.length > 0 && (
            <Tabs.Trigger value="gpu">
              <Text>GPU Performance</Text>
              <Badge size="xs" ml={1} colorPalette="gray">4</Badge>
            </Tabs.Trigger>
          )}
        </Tabs.List>

        <Box mt={6}>
          <Tabs.Content value="overview">
            <Suspense fallback={
              <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            }>
              <OverviewTab job={job} elapsed={elapsed} gpuInfo={gpuInfo} />
            </Suspense>
          </Tabs.Content>

          <Tabs.Content value="metrics">
            <Suspense fallback={
              <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            }>
              <PerformanceMetricsTab job={job} report={report} elapsed={elapsed} />
            </Suspense>
          </Tabs.Content>

          <Tabs.Content value="timeline">
            <Suspense fallback={
              <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg" />
              </Box>
            }>
              <ResourceTimelineTab 
                cluster={clusterName} 
                jobId={jobIdNum} 
                client={client} 
              />
            </Suspense>
          </Tabs.Content>

          {gpuInfo.uuids.length > 0 && (
            <Tabs.Content value="gpu">
              <Suspense fallback={
                <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
                  <Spinner size="lg" />
                </Box>
              }>
                <GpuPerformanceTab 
                  cluster={clusterName} 
                  jobId={jobIdNum} 
                  client={client}
                  gpuUuids={gpuInfo.uuids}
                />
              </Suspense>
            </Tabs.Content>
          )}
        </Box>
      </Tabs.Root>
    </VStack>
  )
}
