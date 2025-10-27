import { useState } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useFormik } from 'formik'
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
} from '@chakra-ui/react'

import { getClusterByClusterJobsQueryOptions } from '../../client/@tanstack/react-query.gen'
import type { JobResponse, GetClusterByClusterJobsQueryData } from '../../client'

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

export const JobsPage = () => {
  const {clusterName} = useParams<{ clusterName: string }>()

  const [hasSearched, setHasSearched] = useState(false)
  const [queryParams, setQueryParams] = useState<GetClusterByClusterJobsQueryData['query']>({})

  const formik = useFormik<JobQueryFormValues>({
    initialValues: {
      user: '',
      userId: '',
      jobId: '',
      states: '',
      limit: '100',
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
      const params: GetClusterByClusterJobsQueryData['query'] = {}

      if (values.user) params.user = values.user
      if (values.userId) params.user_id = parseInt(values.userId)
      if (values.jobId) params.job_id = parseInt(values.jobId)
      if (values.states) params.states = values.states
      if (values.limit) params.limit = parseInt(values.limit)

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

  const jobsQuery = useQuery({
    ...getClusterByClusterJobsQueryOptions({
      path: {cluster: clusterName ?? ''},
      query: queryParams,
    }),
    enabled: !!clusterName && hasSearched,
  })

  const jobs = (jobsQuery.data as JobResponse[]) ?? []

  const getStateColor = (state: string) => {
    if (state === 'COMPLETED') return 'green'
    if (state === 'FAILED') return 'red'
    if (state === 'TIMEOUT') return 'orange'
    if (state === 'CANCELLED') return 'gray'
    if (state === 'RUNNING') return 'blue'
    if (state === 'PENDING') return 'yellow'
    return 'gray'
  }

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
      <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md" p={4} bg="white">
        <form onSubmit={formik.handleSubmit}>
          <VStack align="start" gap={4} w="100%">
            <Text fontSize="lg" fontWeight="semibold">Search Criteria</Text>

            <SimpleGrid columns={{base: 1, md: 2, lg: 3}} gap={4} w="100%">
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
                <Field.Label fontSize="sm">States (comma-separated)</Field.Label>
                <Input
                  name="states"
                  value={formik.values.states}
                  onChange={formik.handleChange}
                  placeholder="e.g., COMPLETED,FAILED"
                  size="sm"
                />
              </Field.Root>

              <Field.Root>
                <Field.Label fontSize="sm">Result Limit</Field.Label>
                <Input
                  name="limit"
                  type="number"
                  value={formik.values.limit}
                  onChange={formik.handleChange}
                  placeholder="Default: 100"
                  size="sm"
                />
              </Field.Root>
            </SimpleGrid>

            {/* Date Filters */}
            <Text fontSize="md" fontWeight="semibold" mt={2}>Start Time Range</Text>
            <SimpleGrid columns={{base: 1, md: 2}} gap={4} w="100%">
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
            </SimpleGrid>

            <Text fontSize="md" fontWeight="semibold" mt={2}>End Time Range</Text>
            <SimpleGrid columns={{base: 1, md: 2}} gap={4} w="100%">
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
            </SimpleGrid>

            <Text fontSize="md" fontWeight="semibold" mt={2}>Submit Time Range</Text>
            <SimpleGrid columns={{base: 1, md: 2}} gap={4} w="100%">
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
            <Text fontSize="md" fontWeight="semibold" mt={2}>Duration (seconds)</Text>
            <SimpleGrid columns={{base: 1, md: 2}} gap={4} w="100%">
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
            <HStack gap={3} mt={4}>
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
                          <Tag.Root size="sm" colorPalette={getStateColor(job.job_state ?? '')}>
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
