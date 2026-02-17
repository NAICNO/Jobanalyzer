import { useState, useEffect, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router'
import { useFormik } from 'formik'
import { AgGridReact } from 'ag-grid-react'
import type { ColDef, ICellRendererParams } from 'ag-grid-community'
import { themeQuartz } from 'ag-grid-community'

import { useClusterClient } from '../../hooks/useClusterClient'
import {
  VStack,
  HStack,
  Text,
  Box,
  SimpleGrid,
  Input,
  Button,
  Badge,
  Spinner,
  Alert,
  Field,
  Select,
  createListCollection,
  Portal,
  Collapsible,
  Icon,
  Pagination,
  IconButton,
  NativeSelect,
} from '@chakra-ui/react'

import type { JobResponse, GetClusterByClusterQueryJobsPagesData } from '../../client'
import { useJobQueryPages } from '../../hooks/v2/useJobQueries'
import { JobState } from '../../types/jobStates'
import { getJobStateColor, formatDurationBetweenDates, formatDateTimeToLocaleString } from '../../util/formatters'
import { LuChevronDown, LuChevronRight, LuChevronLeft } from 'react-icons/lu'

interface JobQueryFormValues {
  user: string
  userId: string
  jobId: string
  states: string
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

interface QueriesPageProps {
  filter?: 'running'
}

// Column definitions factory
const createColumnDefs = (
  clusterName: string,
  navigate: (path: string) => void
): ColDef<JobResponse>[] => [
  {
    field: 'job_id',
    headerName: 'Job ID',
    width: 120,
    filter: 'agNumberColumnFilter',
    pinned: 'left',
    cellRenderer: (params: ICellRendererParams<JobResponse>) => {
      const jobStep = params.data?.job_step
      const display = `${params.value}${jobStep ? `.${jobStep}` : ''}`
      return (
        <span
          style={{ color: '#3182ce', cursor: 'pointer', fontWeight: 500 }}
          onClick={() => {
            if (params.data?.job_id) {
              navigate(`/v2/${clusterName}/jobs/${params.data.job_id}`)
            }
          }}
        >
          {display}
        </span>
      )
    },
  },
  {
    field: 'user_name',
    headerName: 'User',
    width: 120,
    filter: 'agTextColumnFilter',
  },
  {
    field: 'job_name',
    headerName: 'Job Name',
    flex: 1,
    minWidth: 200,
    filter: 'agTextColumnFilter',
    valueFormatter: (params) => params.value ?? 'N/A',
  },
  {
    field: 'job_state',
    headerName: 'State',
    width: 130,
    filter: 'agTextColumnFilter',
    cellRenderer: (params: ICellRendererParams<JobResponse>) => {
      const state = params.value
      const colorPalette = getJobStateColor(state)
      return <Badge colorPalette={colorPalette} size="sm">{state}</Badge>
    },
  },
  {
    field: 'partition',
    headerName: 'Partition',
    width: 130,
    filter: 'agTextColumnFilter',
    valueFormatter: (params) => params.value ?? 'N/A',
  },
  {
    field: 'nodes',
    headerName: 'Nodes',
    width: 200,
    filter: 'agTextColumnFilter',
    valueFormatter: (params) => {
      const nodes = params.value
      if (!nodes || nodes.length === 0) return 'N/A'
      return nodes.join(', ')
    },
  },
  {
    field: 'submit_time',
    headerName: 'Submit Time',
    width: 180,
    filter: 'agDateColumnFilter',
    valueFormatter: (params) => formatDateTimeToLocaleString(params.value),
  },
  {
    field: 'start_time',
    headerName: 'Start Time',
    width: 180,
    filter: 'agDateColumnFilter',
    valueFormatter: (params) => formatDateTimeToLocaleString(params.value),
  },
  {
    field: 'end_time',
    headerName: 'End Time',
    width: 180,
    filter: 'agDateColumnFilter',
    valueFormatter: (params) => formatDateTimeToLocaleString(params.value),
  },
  {
    headerName: 'Duration',
    width: 130,
    valueGetter: (params) => {
      return formatDurationBetweenDates(params.data?.start_time, params.data?.end_time)
    },
  },
  {
    field: 'account',
    headerName: 'Account',
    width: 130,
    filter: 'agTextColumnFilter',
    valueFormatter: (params) => params.value ?? 'N/A',
  },
]

export const QueriesPage = ({ filter }: QueriesPageProps = {}) => {
  const {clusterName} = useParams<{ clusterName: string }>()
  const navigate = useNavigate()

  const client = useClusterClient(clusterName)
  if (!client) {
    return <Spinner />
  }

  // Get today at 00:00 in local datetime format for input
  const getTodayMidnight = () => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const year = today.getFullYear()
    const month = String(today.getMonth() + 1).padStart(2, '0')
    const day = String(today.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}T00:00`
  }

  const [hasSearched, setHasSearched] = useState(false)
  const [queryParams, setQueryParams] = useState<GetClusterByClusterQueryJobsPagesData['query']>({})
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(25)
  const [isFormOpen, setIsFormOpen] = useState(true)

  const formik = useFormik<JobQueryFormValues>({
    initialValues: {
      user: '',
      userId: '',
      jobId: '',
      states: filter === 'running' ? JobState.RUNNING : '',
      startAfter: '',
      startBefore: '',
      endAfter: '',
      endBefore: '',
      submitAfter: getTodayMidnight(),
      submitBefore: '',
      minDuration: '',
      maxDuration: '',
    },
    onSubmit: (values) => {
      const params: GetClusterByClusterQueryJobsPagesData['query'] = {}

      if (values.user) params.user = values.user
      if (values.userId) params.user_id = parseInt(values.userId)
      if (values.jobId) params.job_id = parseInt(values.jobId)
      if (values.states) params.states = values.states

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
      setCurrentPage(1)
    },
  })

  // Auto-submit the form when filter prop is provided
  useEffect(() => {
    if (filter === 'running' && !hasSearched) {
      formik.handleSubmit()
    }
  }, [filter, hasSearched])

  const jobsQuery = useJobQueryPages({
    cluster: clusterName ?? '',
    client,
    queryParams: queryParams ?? {},
    page: currentPage,
    pageSize,
    enabled: hasSearched,
  })

  const jobs = (jobsQuery.data?.jobs as JobResponse[]) ?? []
  const totalJobs = jobsQuery.data?.total ?? 0

  // Auto-collapse form after successful query with results
  useEffect(() => {
    if (hasSearched && !jobsQuery.isLoading && jobs.length > 0) {
      const timer = setTimeout(() => {
        setIsFormOpen(false)
      }, 400)
      return () => clearTimeout(timer)
    }
  }, [hasSearched, jobsQuery.isLoading, jobs.length])

  // Generate summary of active filters
  const filterSummary = useMemo(() => {
    const filters: string[] = []
    if (formik.values.user) filters.push(`User: ${formik.values.user}`)
    if (formik.values.userId) filters.push(`User ID: ${formik.values.userId}`)
    if (formik.values.jobId) filters.push(`Job ID: ${formik.values.jobId}`)
    if (formik.values.states) {
      const stateLabels = formik.values.states.split(',').join(', ')
      filters.push(`States: ${stateLabels}`)
    }
    if (formik.values.minDuration) filters.push(`Min Duration: ${formik.values.minDuration}s`)
    if (formik.values.maxDuration) filters.push(`Max Duration: ${formik.values.maxDuration}s`)
    if (formik.values.startAfter) filters.push(`Start After: ${new Date(formik.values.startAfter).toLocaleDateString()}`)
    if (formik.values.startBefore) filters.push(`Start Before: ${new Date(formik.values.startBefore).toLocaleDateString()}`)
    if (formik.values.endAfter) filters.push(`End After: ${new Date(formik.values.endAfter).toLocaleDateString()}`)
    if (formik.values.endBefore) filters.push(`End Before: ${new Date(formik.values.endBefore).toLocaleDateString()}`)
    if (formik.values.submitAfter) filters.push(`Submit After: ${new Date(formik.values.submitAfter).toLocaleDateString()}`)
    if (formik.values.submitBefore) filters.push(`Submit Before: ${new Date(formik.values.submitBefore).toLocaleDateString()}`)
    
    return filters.length > 0 ? filters.join(' • ') : 'No filters applied'
  }, [formik.values])

  // AG Grid column definitions
  const columnDefs = useMemo<ColDef<JobResponse>[]>(
    () => createColumnDefs(clusterName ?? '', navigate),
    [clusterName, navigate]
  )

  const defaultColDef = useMemo<ColDef>(
    () => ({
      sortable: true,
      resizable: true,
    }),
    []
  )

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
          <Collapsible.Root open={isFormOpen} onOpenChange={(e) => setIsFormOpen(e.open)}>
            <Collapsible.Trigger
              width="100%"
              display="flex"
              alignItems="center"
              gap={2}
              py={2}
              px={1}
              _hover={{ bg: 'gray.50' }}
              rounded="md"
              transition="background 0.2s"
            >
              <Icon fontSize="xl">
                <Collapsible.Context>
                  {(api) => (api.open ? <LuChevronDown/> : <LuChevronRight />)}
                </Collapsible.Context>
              </Icon>
              <VStack align="start" gap={0} flex="1">
                <Text fontSize="lg" fontWeight="semibold">Search Criteria</Text>
                {!isFormOpen && (
                  <Text fontSize="xs" color="gray.500" truncate maxW="100%">
                    {filterSummary}
                  </Text>
                )}
              </VStack>
            </Collapsible.Trigger>

            <Collapsible.Content>
              <VStack align="start" gap={2} w="100%" mt={2}>
                {/* Basic Filters */}
                <SimpleGrid columns={{base: 2, md: 3, lg: 6}} gap={2} w="100%">
                  <Field.Root>
                    <Field.Label fontSize="sm">User</Field.Label>
                    <Input
                      name="user"
                      value={formik.values.user}
                      onChange={formik.handleChange}
                      placeholder="e.g., username"
                      size="sm"
                      borderLeftWidth={formik.values.user ? '3px' : undefined}
                      borderLeftColor={formik.values.user ? 'blue.500' : undefined}
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
                      borderLeftWidth={formik.values.userId ? '3px' : undefined}
                      borderLeftColor={formik.values.userId ? 'blue.500' : undefined}
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
                      borderLeftWidth={formik.values.jobId ? '3px' : undefined}
                      borderLeftColor={formik.values.jobId ? 'blue.500' : undefined}
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
                      <Select.Control
                        borderLeftWidth={formik.values.states ? '3px' : undefined}
                        borderLeftColor={formik.values.states ? 'blue.500' : undefined}
                      >
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

                  <Field.Root>
                    <Field.Label fontSize="sm">Min Duration (s)</Field.Label>
                    <Input
                      name="minDuration"
                      type="number"
                      value={formik.values.minDuration}
                      onChange={formik.handleChange}
                      placeholder="e.g., 3600"
                      size="sm"
                      borderLeftWidth={formik.values.minDuration ? '3px' : undefined}
                      borderLeftColor={formik.values.minDuration ? 'blue.500' : undefined}
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
                      borderLeftWidth={formik.values.maxDuration ? '3px' : undefined}
                      borderLeftColor={formik.values.maxDuration ? 'blue.500' : undefined}
                    />
                  </Field.Root>
                </SimpleGrid>

                {/* Time Range Filters in Grouped Containers */}
                <HStack gap={2} w="100%" flexWrap="wrap" align="start">
                  <Box borderWidth="1px" borderColor="gray.300" rounded="md" p={2} flex="1" minW="200px">
                    <Text fontSize="xs" fontWeight="semibold" mb={1} color="gray.600">Start Times</Text>
                    <VStack gap={2} align="stretch">
                      <Field.Root>
                        <Field.Label fontSize="sm">After</Field.Label>
                        <Input
                          name="startAfter"
                          type="datetime-local"
                          value={formik.values.startAfter}
                          onChange={formik.handleChange}
                          size="sm"
                          borderLeftWidth={formik.values.startAfter ? '3px' : undefined}
                          borderLeftColor={formik.values.startAfter ? 'blue.500' : undefined}
                        />
                      </Field.Root>
                      <Field.Root>
                        <Field.Label fontSize="sm">Before</Field.Label>
                        <Input
                          name="startBefore"
                          type="datetime-local"
                          value={formik.values.startBefore}
                          onChange={formik.handleChange}
                          size="sm"
                          borderLeftWidth={formik.values.startBefore ? '3px' : undefined}
                          borderLeftColor={formik.values.startBefore ? 'blue.500' : undefined}
                        />
                      </Field.Root>
                    </VStack>
                  </Box>

                  <Box borderWidth="1px" borderColor="gray.300" rounded="md" p={2} flex="1" minW="200px">
                    <Text fontSize="xs" fontWeight="semibold" mb={1} color="gray.600">End Times</Text>
                    <VStack gap={2} align="stretch">
                      <Field.Root>
                        <Field.Label fontSize="sm">After</Field.Label>
                        <Input
                          name="endAfter"
                          type="datetime-local"
                          value={formik.values.endAfter}
                          onChange={formik.handleChange}
                          size="sm"
                          borderLeftWidth={formik.values.endAfter ? '3px' : undefined}
                          borderLeftColor={formik.values.endAfter ? 'blue.500' : undefined}
                        />
                      </Field.Root>
                      <Field.Root>
                        <Field.Label fontSize="sm">Before</Field.Label>
                        <Input
                          name="endBefore"
                          type="datetime-local"
                          value={formik.values.endBefore}
                          onChange={formik.handleChange}
                          size="sm"
                          borderLeftWidth={formik.values.endBefore ? '3px' : undefined}
                          borderLeftColor={formik.values.endBefore ? 'blue.500' : undefined}
                        />
                      </Field.Root>
                    </VStack>
                  </Box>

                  <Box borderWidth="1px" borderColor="gray.300" rounded="md" p={2} flex="1" minW="200px">
                    <Text fontSize="xs" fontWeight="semibold" mb={1} color="gray.600">Submit Times</Text>
                    <VStack gap={2} align="stretch">
                      <Field.Root>
                        <Field.Label fontSize="sm">After</Field.Label>
                        <Input
                          name="submitAfter"
                          type="datetime-local"
                          value={formik.values.submitAfter}
                          onChange={formik.handleChange}
                          size="sm"
                          borderLeftWidth={formik.values.submitAfter ? '3px' : undefined}
                          borderLeftColor={formik.values.submitAfter ? 'blue.500' : undefined}
                        />
                      </Field.Root>
                      <Field.Root>
                        <Field.Label fontSize="sm">Before</Field.Label>
                        <Input
                          name="submitBefore"
                          type="datetime-local"
                          value={formik.values.submitBefore}
                          onChange={formik.handleChange}
                          size="sm"
                          borderLeftWidth={formik.values.submitBefore ? '3px' : undefined}
                          borderLeftColor={formik.values.submitBefore ? 'blue.500' : undefined}
                        />
                      </Field.Root>
                    </VStack>
                  </Box>
                </HStack>

                {/* Action Buttons */}
                <HStack gap={2}>
                  <Button 
                    type="submit" 
                    colorPalette="blue" 
                    size="sm"
                    loading={jobsQuery.isLoading}
                  >
                    Search Jobs
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      formik.resetForm()
                      setHasSearched(false)
                      setQueryParams({})
                      setIsFormOpen(true)
                    }}
                  >
                    Reset
                  </Button>
                </HStack>
              </VStack>
            </Collapsible.Content>
          </Collapsible.Root>
        </form>
      </Box>

      {/* Results */}
      {hasSearched && (
        <Box w="100%" borderWidth="1px" borderColor="gray.200" rounded="md" p={4} bg="white">
          <VStack align="start" gap={3} w="100%">
            <HStack gap={2}>
              <Text textStyle="lg" fontWeight="semibold">
                Search Results
              </Text>
              {!jobsQuery.isLoading && (
                <Text textStyle="sm" color="gray.600">
                  ({totalJobs} {totalJobs === 1 ? 'result' : 'results'})
                </Text>
              )}
            </HStack>

            {jobsQuery.isLoading ? (
              <Box w="100%" h="400px" display="flex" alignItems="center" justifyContent="center">
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
              <VStack w="100%" h="calc(100vh - 175px)" gap={3}>
                <Box w="100%" flex="1">
                  <AgGridReact<JobResponse>
                    theme={themeQuartz}
                    rowData={jobs}
                    columnDefs={columnDefs}
                    defaultColDef={defaultColDef}
                    getRowId={(params) => `${params.data.job_id}-${params.data.job_step ?? ''}`}
                    pagination={false}
                    animateRows={true}
                    enableCellTextSelection={true}
                  />
                </Box>
                
                {/* Custom Pagination Controls */}
                <HStack w="100%" justify="space-between" py={2}>
                  <HStack gap={2}>
                    <Text fontSize="sm" color="gray.600">Rows per page:</Text>
                    <NativeSelect.Root size="sm" width="80px">
                      <NativeSelect.Field
                        value={String(pageSize)}
                        onChange={(e) => {
                          setPageSize(Number(e.currentTarget.value))
                          setCurrentPage(1)
                        }}
                      >
                        <option value="10">10</option>
                        <option value="25">25</option>
                        <option value="50">50</option>
                        <option value="100">100</option>
                      </NativeSelect.Field>
                      <NativeSelect.Indicator />
                    </NativeSelect.Root>
                  </HStack>
                  
                  <Pagination.Root
                    count={totalJobs}
                    pageSize={pageSize}
                    page={currentPage}
                    onPageChange={(e) => setCurrentPage(e.page)}
                  >
                    <HStack gap={2}>
                      <Pagination.PrevTrigger asChild>
                        <IconButton size="sm" variant="ghost">
                          <LuChevronLeft />
                        </IconButton>
                      </Pagination.PrevTrigger>

                      <Pagination.Items
                        render={(page) => (
                          <IconButton
                            key={page.value}
                            size="sm"
                            variant={page.value === currentPage ? 'solid' : 'ghost'}
                          >
                            {page.value}
                          </IconButton>
                        )}
                      />

                      <Pagination.NextTrigger asChild>
                        <IconButton size="sm" variant="ghost">
                          <LuChevronRight />
                        </IconButton>
                      </Pagination.NextTrigger>
                    </HStack>
                  </Pagination.Root>
                </HStack>
              </VStack>
            )}
          </VStack>
        </Box>
      )}
    </VStack>
  )
}
