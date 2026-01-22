import { useState, useEffect } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useFormik } from 'formik'

import { useClusterClient } from '../../hooks/useClusterClient'
import {
  VStack,
  HStack,
  Text,
  Box,
  SimpleGrid,
  Input,
  Button,
  Table,
  Tag,
  Spinner,
  Alert,
  Field,
  Select,
  createListCollection,
  Portal,
} from '@chakra-ui/react'

import { getClusterByClusterJobsOptions } from '../../client/@tanstack/react-query.gen'
import type { JobResponse, GetClusterByClusterJobsData } from '../../client'
import { JobState } from '../../types/jobStates'
import { getJobStateColor } from '../../util/formatters'

interface JobQueryFormValues {
  user: string
  userId: string
  jobId: string
  states: string
  limit: string
  startAfter: string
  startBefore: string
  endAfter: string
  endBefore: string
  submitAfter: string
  submitBefore: string
  minDuration: string
  maxDuration: string
}

const statesCollection = createListCollection({
  items: [
    { label: 'COMPLETED', value: JobState.COMPLETED },
    { label: 'FAILED', value: JobState.FAILED },
    { label: 'TIMEOUT', value: JobState.TIMEOUT },
    { label: 'CANCELLED', value: JobState.CANCELLED },
    { label: 'RUNNING', value: JobState.RUNNING },
    { label: 'PENDING', value: JobState.PENDING },
    { label: 'SUSPENDED', value: JobState.SUSPENDED },
    { label: 'PREEMPTED', value: JobState.PREEMPTED },
    { label: 'OUT_OF_MEMORY', value: JobState.OUT_OF_MEMORY },
    { label: 'NODE_FAIL', value: JobState.NODE_FAIL },
    { label: 'BOOT_FAIL', value: JobState.BOOT_FAIL },
    { label: 'DEADLINE', value: JobState.DEADLINE },
  ],
})

const limitCollection = createListCollection({
  items: [
    { label: 'All', value: 'all' },
    { label: '10', value: '10' },
    { label: '25', value: '25' },
    { label: '50', value: '50' },
    { label: '100', value: '100' },
    { label: '500', value: '500' },
    { label: '1000', value: '1000' },
  ],
})

interface QueriesPageProps {
  filter?: 'running'
}

export const QueriesPage = ({ filter }: QueriesPageProps = {}) => {
  const {clusterName} = useParams<{ clusterName: string }>()

  const client = useClusterClient(clusterName)
  if (!client) {
    return <Spinner />
  }

  const [hasSearched, setHasSearched] = useState(false)
  const [queryParams, setQueryParams] = useState<GetClusterByClusterJobsData['query']>({})

  const formik = useFormik<JobQueryFormValues>({
    initialValues: {
      user: '',
      userId: '',
      jobId: '',
      states: filter === 'running' ? JobState.RUNNING : '',
      limit: 'all',
      startAfter: '',
      startBefore: '',
      endAfter: '',
      endBefore: '',
      submitAfter: '',
      submitBefore: '',
      minDuration: '',
      maxDuration: '',
    },
    onSubmit: (values) => {
      const params: GetClusterByClusterJobsData['query'] = {}

      if (values.user) params.user = values.user
      if (values.userId) params.user_id = parseInt(values.userId)
      if (values.jobId) params.job_id = parseInt(values.jobId)
      if (values.states) params.states = values.states
      if (values.limit && values.limit !== 'all') params.limit = parseInt(values.limit)

      // Convert date strings to timestamps (seconds)
      if (values.startAfter) {
        const date = new Date(values.startAfter)
        params.start_after_in_s = Math.floor(date.getTime() / 1000)
      }
      if (values.startBefore) {
        const date = new Date(values.startBefore)
        params.start_before_in_s = Math.floor(date.getTime() / 1000)
      }
      if (values.endAfter) {
        const date = new Date(values.endAfter)
        params.end_after_in_s = Math.floor(date.getTime() / 1000)
      }
      if (values.endBefore) {
        const date = new Date(values.endBefore)
        params.end_before_in_s = Math.floor(date.getTime() / 1000)
      }
      if (values.submitAfter) {
        const date = new Date(values.submitAfter)
        params.submit_after_in_s = Math.floor(date.getTime() / 1000)
      }
      if (values.submitBefore) {
        const date = new Date(values.submitBefore)
        params.submit_before_in_s = Math.floor(date.getTime() / 1000)
      }

      // Duration in seconds
      if (values.minDuration) params.min_duration_in_s = parseInt(values.minDuration)
      if (values.maxDuration) params.max_duration_in_s = parseInt(values.maxDuration)

      setQueryParams(params)
      setHasSearched(true)
    },
  })

  // Auto-submit the form when filter prop is provided
  useEffect(() => {
    if (filter === 'running' && !hasSearched) {
      formik.handleSubmit()
    }
  }, [filter, hasSearched])

  const jobsQuery = useQuery({
    ...getClusterByClusterJobsOptions({
      path: {cluster: clusterName ?? ''},
      query: queryParams,
      client,
    }),
    enabled: !!clusterName && hasSearched,
  })

  const jobs = (jobsQuery.data as JobResponse[]) ?? []


  const formatDuration = (startTime?: Date | null, endTime?: Date | null) => {
    if (!startTime || !endTime) return 'N/A'
    const start = new Date(startTime).getTime()
    const end = new Date(endTime).getTime()
    const durationMs = end - start
    if (durationMs <= 0) return 'N/A'

    const seconds = Math.floor(durationMs / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (days > 0) return `${days}d ${hours % 24}h`
    if (hours > 0) return `${hours}h ${minutes % 60}m`
    if (minutes > 0) return `${minutes}m ${seconds % 60}s`
    return `${seconds}s`
  }

  const formatDateTime = (date?: Date | null) => {
    if (!date) return 'N/A'
    return new Date(date).toLocaleString()
  }

  if (!clusterName) {
    return <Text>No cluster selected</Text>
  }

  return (
    <VStack w="100%" align="start" gap={6} p={4}>
      <Text fontSize="2xl" fontWeight="bold">
        Job Query - {clusterName}
      </Text>

      {/* Search Form */}
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md" p={3} bg="white">
        <form onSubmit={formik.handleSubmit}>
          <VStack align="start" gap={2} w="100%">
            <Text fontSize="lg" fontWeight="semibold">Search Criteria</Text>

            <SimpleGrid columns={{base: 1, md: 3, lg: 4}} gap={2} w="100%">
              {/* Basic Filters */}
              <Field.Root>
                <Field.Label fontSize="sm">User</Field.Label>
                <Input
                  name="user"
                  value={formik.values.user}
                  onChange={formik.handleChange}
                  placeholder="e.g., username"
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">User ID</Field.Label>
                <Input
                  name="userId"
                  type="number"
                  value={formik.values.userId}
                  onChange={formik.handleChange}
                  placeholder="e.g., 1000"
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">Job ID</Field.Label>
                <Input
                  name="jobId"
                  type="number"
                  value={formik.values.jobId}
                  onChange={formik.handleChange}
                  placeholder="e.g., 12345"
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">States</Field.Label>
                <Select.Root
                  multiple
                  collection={statesCollection}
                  value={formik.values.states ? formik.values.states.split(',') : []}
                  onValueChange={(details) => {
                    formik.setFieldValue('states', details.value.join(','))
                  }}
                  size="sm"
                >
                  <Select.HiddenSelect />
                  <Select.Control>
                    <Select.Trigger>
                      <Select.ValueText placeholder="Select job states" />
                    </Select.Trigger>
                    <Select.IndicatorGroup>
                      <Select.Indicator />
                    </Select.IndicatorGroup>
                  </Select.Control>
                  <Portal>
                    <Select.Positioner>
                      <Select.Content>
                        {statesCollection.items.map((state) => (
                          <Select.Item item={state} key={state.value}>
                            {state.label}
                            <Select.ItemIndicator />
                          </Select.Item>
                        ))}
                      </Select.Content>
                    </Select.Positioner>
                  </Portal>
                </Select.Root>
              </Field.Root>
            </SimpleGrid>

            {/* Time Range Filters */}
            <Text fontSize="md" fontWeight="semibold" mt={1}>Time Ranges</Text>
            <SimpleGrid columns={{base: 1, md: 3, lg: 6}} gap={2} w="100%">
              <Field.Root>
                <Field.Label fontSize="sm">Start After</Field.Label>
                <Input
                  name="startAfter"
                  type="datetime-local"
                  value={formik.values.startAfter}
                  onChange={formik.handleChange}
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">Start Before</Field.Label>
                <Input
                  name="startBefore"
                  type="datetime-local"
                  value={formik.values.startBefore}
                  onChange={formik.handleChange}
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">End After</Field.Label>
                <Input
                  name="endAfter"
                  type="datetime-local"
                  value={formik.values.endAfter}
                  onChange={formik.handleChange}
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">End Before</Field.Label>
                <Input
                  name="endBefore"
                  type="datetime-local"
                  value={formik.values.endBefore}
                  onChange={formik.handleChange}
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">Submit After</Field.Label>
                <Input
                  name="submitAfter"
                  type="datetime-local"
                  value={formik.values.submitAfter}
                  onChange={formik.handleChange}
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">Submit Before</Field.Label>
                <Input
                  name="submitBefore"
                  type="datetime-local"
                  value={formik.values.submitBefore}
                  onChange={formik.handleChange}
                  size="sm"
                />
              </Field.Root>
            </SimpleGrid>

            {/* Duration Filters */}
            <Text fontSize="md" fontWeight="semibold" mt={1}>Duration (seconds)</Text>
            <SimpleGrid columns={{base: 1, md: 2}} gap={2} w="100%">
              <Field.Root>
                <Field.Label fontSize="sm">Min Duration (s)</Field.Label>
                <Input
                  name="minDuration"
                  type="number"
                  value={formik.values.minDuration}
                  onChange={formik.handleChange}
                  placeholder="e.g., 3600"
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">Max Duration (s)</Field.Label>
                <Input
                  name="maxDuration"
                  type="number"
                  value={formik.values.maxDuration}
                  onChange={formik.handleChange}
                  placeholder="e.g., 86400"
                  size="sm"
                />
              </Field.Root>
            </SimpleGrid>

            {/* Action Buttons */}
            <HStack gap={3} mt={4} w="100%" justify="space-between">
              <HStack gap={3}>
                <Button type="submit" colorPalette="blue">
                  Search Jobs
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    formik.resetForm()
                    setHasSearched(false)
                    setQueryParams({})
                  }}
                >
                  Reset
                </Button>
              </HStack>
              <Field.Root maxW="120px">
                <Field.Label fontSize="sm">Result Limit</Field.Label>
                <Select.Root
                  collection={limitCollection}
                  value={[formik.values.limit]}
                  onValueChange={(details) => {
                    formik.setFieldValue('limit', details.value[0])
                  }}
                  size="sm"
                >
                  <Select.HiddenSelect />
                  <Select.Control>
                    <Select.Trigger>
                      <Select.ValueText />
                    </Select.Trigger>
                    <Select.IndicatorGroup>
                      <Select.Indicator />
                    </Select.IndicatorGroup>
                  </Select.Control>
                  <Portal>
                    <Select.Positioner>
                      <Select.Content>
                        {limitCollection.items.map((item) => (
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
          </VStack>
        </form>
      </Box>

      {/* Results */}
      {hasSearched && (
        <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md" p={4} bg="white">
          <VStack align="start" gap={3} w="100%">
            <HStack justify="space-between" w="100%">
              <Text fontSize="lg" fontWeight="semibold">
                Search Results
              </Text>
              {!jobsQuery.isLoading && (
                <Text fontSize="sm" color="gray.600">
                  Found {jobs.length} job{jobs.length !== 1 ? 's' : ''}
                </Text>
              )}
            </HStack>

            {jobsQuery.isLoading ? (
              <Box w="100%" h="200px" display="flex" alignItems="center" justifyContent="center">
                <Spinner size="lg"/>
              </Box>
            ) : jobsQuery.isError ? (
              <Alert.Root status="error">
                <Alert.Indicator/>
                <Alert.Title>Error loading jobs</Alert.Title>
                <Alert.Description>
                  {jobsQuery.error.message}
                </Alert.Description>
              </Alert.Root>
            ) : jobs.length === 0 ? (
              <Box w="100%" textAlign="center" py={8}>
                <Text color="gray.500">No jobs found matching your criteria</Text>
              </Box>
            ) : (
              <Box w="100%" overflowX="auto">
                <Table.Root size="sm" variant="outline">
                  <Table.Header>
                    <Table.Row bg="gray.50">
                      <Table.ColumnHeader fontSize="xs">Job ID</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">User</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">Job Name</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">State</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">Submit Time</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">Start Time</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">End Time</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">Duration</Table.ColumnHeader>
                      <Table.ColumnHeader fontSize="xs">Account</Table.ColumnHeader>
                    </Table.Row>
                  </Table.Header>
                  <Table.Body>
                    {jobs.map((job) => (
                      <Table.Row key={`${job.job_id}-${job.job_step}`}>
                        <Table.Cell fontSize="xs" fontWeight="medium">
                          {job.job_id}{job.job_step ? `.${job.job_step}` : ''}
                        </Table.Cell>
                        <Table.Cell fontSize="xs">{job.user_name}</Table.Cell>
                        <Table.Cell fontSize="xs" maxW="200px" truncate>
                          {job.job_name ?? 'N/A'}
                        </Table.Cell>
                        <Table.Cell fontSize="xs">
                          <Tag.Root size="sm" colorPalette={getJobStateColor(job.job_state ?? '')}>
                            <Tag.Label>{job.job_state}</Tag.Label>
                          </Tag.Root>
                        </Table.Cell>
                        <Table.Cell fontSize="xs">{formatDateTime(job.submit_time)}</Table.Cell>
                        <Table.Cell fontSize="xs">{formatDateTime(job.start_time)}</Table.Cell>
                        <Table.Cell fontSize="xs">{formatDateTime(job.end_time)}</Table.Cell>
                        <Table.Cell fontSize="xs">{formatDuration(job.start_time, job.end_time)}</Table.Cell>
                        <Table.Cell fontSize="xs">{job.account ?? 'N/A'}</Table.Cell>
                      </Table.Row>
                    ))}
                  </Table.Body>
                </Table.Root>
              </Box>
            )}
          </VStack>
        </Box>
      )}
    </VStack>
  )
}
