import { useEffect, useMemo, useState } from 'react'
import {
  Button,
  Card,
  Grid,
  HStack,
  Heading,
  Link,
  Spacer,
  Stack,
  Text,
  VStack,
  useBreakpointValue,
  useDisclosure,
} from '@chakra-ui/react'
import { getCoreRowModel, getSortedRowModel, SortingState, useReactTable } from '@tanstack/react-table'
import { Form, Formik } from 'formik'
import { useSearchParams } from 'react-router'

import {
  EMPTY_ARRAY,
  JOB_QUERY_GPU_OPTIONS,
  JOB_QUERY_VALIDATION_SCHEMA,
  initialFormValues,
  JOB_QUERY_RESULTS_COLUMN,
} from '../Constants.ts'
import { useFetchJobQuery } from '../hooks/useFetchJobQuery.ts'
import { JobQueryFormTextInput, JobQueryFormRadioInput, ShareLinkPopover, PageTitle, } from '../components'
import { JobQueryResultsTable } from '../components/table'
import { getJobQueryResultsTableColumns } from '../util/TableUtils.ts'
import { JobQueryResultsSkeleton } from '../components/skeleton/JobQueryResultsSkeleton.tsx'
import { JobQueryValues } from '../types'
import { prepareShareableJobQueryLink } from '../util/query/QueryUtils.ts'
import JobQueryResultExportModal from '../modals/JobQueryResultExportModal.tsx'
import { LuExternalLink } from 'react-icons/lu'

export default function JobQueryPage() {

  const [formValues, setFormValues] = useState<JobQueryValues>(initialFormValues)
  const [searchParams] = useSearchParams()

  useEffect(() => {
    const clusterName = searchParams.get('cluster')
    const usernames = searchParams.get('user')
    const nodeNames = searchParams.get('host')
    const jobIds = searchParams.get('job')
    const fromDate = searchParams.get('from')
    const toDate = searchParams.get('to')
    const minRuntime = searchParams.get('min-runtime')
    const minPeakCpuCores = searchParams.get('min-cpu-peak')
    const minPeakResidentGb = searchParams.get('min-res-peak')
    const gpuUsage = searchParams.get('some-gpu') ? 'some-gpu'
      : searchParams.get('no-gpu') ? 'no-gpu' : 'either'

    setFormValues({
      clusterName: clusterName || '',
      usernames: usernames || '',
      nodeNames: nodeNames || '',
      jobIds: jobIds || '',
      fromDate: fromDate || '',
      toDate: toDate || '',
      minRuntime: minRuntime || '',
      minPeakCpuCores: parseInt(minPeakCpuCores || '0'),
      minPeakResidentGb: parseInt(minPeakResidentGb || '0'),
      gpuUsage: gpuUsage,
    })
  }, [searchParams])

  const fields = Object.keys(JOB_QUERY_RESULTS_COLUMN)

  const {data, refetch, isLoading} = useFetchJobQuery(formValues, fields)

  const jobQueryResultsTableColumns = useMemo(() => getJobQueryResultsTableColumns(), [])
  const [sorting, setSorting] = useState<SortingState>([])

  const jobQueryResultsTable = useReactTable({
    columns: jobQueryResultsTableColumns,
    data: data || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: sorting,
    },
    autoResetPageIndex: false,
    autoResetExpanded: false,
  })

  const shareableLink = prepareShareableJobQueryLink(formValues, fields)

  const formGridTemplateColumns = useBreakpointValue({base: '1fr', md: 'repeat(2, 1fr)'})

  const {open, onOpen, onClose, } = useDisclosure()

  return (
    <>
      <PageTitle title={'Job Query'}/>
      <VStack alignItems={'start'}>
        <Card.Root maxW={{md: '700px'}} variant={'outline'}>
          <Card.Body>
            <Heading
              size={{base: 'md', md: 'lg'}}
              mb={2}
            >
              Job Query
            </Heading>
            <Formik
              enableReinitialize={true}
              initialValues={formValues}
              validationSchema={JOB_QUERY_VALIDATION_SCHEMA}
              onSubmit={(values) => {
                setFormValues(values)
                setTimeout(() => {
                  refetch()
                }, 0)
              }}>
              {({
                isValid,
                submitForm,
              }) => (
                <Form>
                  <Grid templateColumns={formGridTemplateColumns} gridColumnGap={10} gridRowGap={3}>
                    <JobQueryFormTextInput
                      name="clusterName"
                      label="Cluster name"
                      type="text"
                      placeholder='"ml", "fox", ...'
                    />
                    <JobQueryFormTextInput
                      name="usernames"
                      label="User name(s)"
                      type="text"
                      placeholder="Comma separated. Eg: user1,user2"
                    />
                    <JobQueryFormTextInput
                      name="nodeNames"
                      label="Node names(s)"
                      type="text"
                      placeholder="Enter node names, e.g., host1,host[2-3]"
                    />
                    <JobQueryFormTextInput
                      name="jobIds"
                      label="Job ID(s)"
                      type="text"
                      placeholder="default all"
                    />
                    <JobQueryFormTextInput
                      name="fromDate"
                      label="From date"
                      type="text"
                      placeholder="YYYY-MM-DD or Nw or Nd"
                    />
                    <JobQueryFormTextInput
                      name="toDate"
                      label="To date"
                      type="text"
                      placeholder="YYYY-MM-DD or Nw or Nd"
                    />
                    <JobQueryFormTextInput
                      name="minRuntime"
                      label="Minimum runtime"
                      type="text"
                      placeholder="Eg: 2d12h=two days, 12 hrs"
                    />
                    <JobQueryFormTextInput
                      name="minPeakCpuCores"
                      label="Minimum peak CPU cores"
                      type="text"
                      placeholder="default 0"
                    />
                    <JobQueryFormTextInput
                      name="minPeakResidentGb"
                      label="Minimum peak Resident GB"
                      type="text"
                      placeholder="default 0"
                    />

                    <JobQueryFormRadioInput
                      name="gpuUsage"
                      label="GPU usage"
                      options={JOB_QUERY_GPU_OPTIONS}
                    />
                  </Grid>
                  <VStack alignItems={'start'} mt={1}>
                    <Stack
                      direction={{base: 'column', md: 'row'}}
                      alignItems={{base: 'start', md: 'center'}}
                      gap={4}
                      mt={4}
                    >
                      <Button
                        colorPalette={'blue'}
                        onClick={submitForm}
                        disabled={!isValid}
                        type="submit"
                      >
                        Select Jobs
                      </Button>
                      <Text as="em" fontSize="sm" ml={'10px'} color="gray">
                        Password protected.{' '}
                        <Link
                          href="https://github.com/NAICNO/Jobanalyzer/issues/new?title=Access"
                          target="_blank"
                        >
                          File an issue with the title &quot;Access&quot; <LuExternalLink/>
                        </Link>
                        {' '}if you need access.
                      </Text>
                    </Stack>
                    <Text as="em" fontSize="sm" mt="10px" color="gray">
                      More query terms, data fields, and profiler options are available with the command line interface.
                    </Text>
                  </VStack>
                </Form>
              )}
            </Formik>
          </Card.Body>
        </Card.Root>
        {
          isLoading &&
          <JobQueryResultsSkeleton/>
        }
        {
          data &&
          <Card.Root mt={1} variant={'outline'}>
            <Card.Header>
              <HStack gap={2}>
                <Heading as="h2" size="lg">
                  Selected Jobs
                </Heading>
                <ShareLinkPopover
                  text={'Share link for this query'}
                  link={shareableLink}
                />
              </HStack>
              <Text fontSize="sm">
                Memory values are in GB, cpu/gpu in percent of one core/card.
              </Text>
              <HStack gap={2} my={2}>
                <Text>
                  Click on a job link in the table below to view profiles of the job.
                </Text>
                <Spacer/>
                <Button
                  colorScheme="blue"
                  type="submit"
                  size="sm"
                  onClick={onOpen}
                >
                  Export
                </Button>
              </HStack>
            </Card.Header>
            <Card.Body pt={-4}>
              <JobQueryResultsTable table={jobQueryResultsTable}/>
            </Card.Body>
          </Card.Root>
        }
      </VStack>
      <JobQueryResultExportModal open={open} onClose={onClose} jobQueryFormValues={formValues}/>
    </>
  )
}
