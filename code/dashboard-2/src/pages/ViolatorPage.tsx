import { useMemo, useState } from 'react'
import {
  Box,
  Card,
  HStack,
  Heading,
  Icon,
  List,
  Text,
  VStack,
} from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'
import moment from 'moment-timezone'
import { IoWarning } from 'react-icons/io5'

import { EMPTY_ARRAY, POLICIES } from '../Constants.ts'
import { findCluster } from '../util'
import { getUserViolatingJobTableColumns } from '../util/TableUtils.ts'
import { useFetchViolator } from '../hooks/useFetchViolator.ts'
import { NavigateBackButton, PageTitle } from '../components'
import { ViolatingJobTable } from '../components/table'

export default function ViolatorPage() {
  const {clusterName, violator} = useParams<string>()

  const cluster = findCluster(clusterName)

  if (!cluster || !violator) {
    return (
      <Navigate to="/"/>
    )
  }

  const {data: allJobsOfUser} = useFetchViolator(cluster, violator)

  const violatingJobTableColumns = useMemo(() => getUserViolatingJobTableColumns(), [cluster])
  const [violatingJobTableSorting, setViolatingJobTableSorting] = useState<SortingState>([])

  const violatingJobTable = useReactTable({
    columns: violatingJobTableColumns,
    data: allJobsOfUser || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setViolatingJobTableSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: violatingJobTableSorting,
    }
  })

  const allPolicyNamesOfAllJobs = allJobsOfUser?.map(job => job.policyName) || EMPTY_ARRAY

  const violatedPolicyNames = Array.from(new Set<string>(allPolicyNamesOfAllJobs))

  const violatedPolicies = violatedPolicyNames.map(policyName => {
    return POLICIES[clusterName!].find(policy => policy.name === policyName)
  })

  const timestamp = moment.utc()
  timestamp.tz('Europe/Oslo')
  const formattedTimestamp = timestamp.format('ddd MMM DD YYYY HH:mm:ss [GMT]ZZ (z)')

  return (
    <>
      <PageTitle title={`${cluster.name} Individual Policy Violator Report`}/>
      <VStack alignItems={'start'}>
        <HStack mb="20px">
          <NavigateBackButton/>
          <Heading ml={3} size={{base: 'md', md: 'lg'}}>{cluster.name} individual policy violator report</Heading>
        </HStack>
        <Card.Root variant={'outline'}>
          <Card.Body>
            <VStack alignItems="start">
              <Text>Hi,</Text>
              <Text>This is a message from your friendly UiO systems administrator.</Text>
              <Text>To ensure that computing resources are used in the best possible way,
                we monitor how jobs are using the systems and ask users to move when
                they are using a particular system in a way that is contrary to the
                intended use of that system.</Text>
              <Text>
                You are receiving this message because you have been running jobs
                in just such a manner, as detailed below. Please apply the suggested
                remedies (usually this means moving your work to another system).
              </Text>
              <Text mt="20px">&quot;{cluster.name}&quot; individual policy violator report</Text>
              <Text mt="10px">Report generated on {formattedTimestamp}</Text>
              <Text mt="10px">User: {violator}</Text>
              <Text mt="10px">Policies violated:</Text>
              <List.Root ml="20px">
                {
                  violatedPolicies.map(policy => {
                    if (!policy) {
                      return null
                    }
                    return (
                      <List.Item key={policy.name}>
                        <List.Indicator>
                          <Icon color={'red.500'}>
                            <IoWarning/>
                          </Icon>
                        </List.Indicator>
                        {policy.name}
                        <Box ml="20px" mt="5px">
                          <List.Root>
                            <List.Item>Trigger: {policy.trigger}</List.Item>
                            <List.Item>Problem: {policy.problem}</List.Item>
                            <List.Item>Remedy: {policy.remedy}</List.Item>
                          </List.Root>
                        </Box>
                      </List.Item>
                    )
                  })
                }
              </List.Root>
              <Text marginY="20px">(Times below are UTC, job numbers are derived from session leader if not running
                under
                Slurm)</Text>
              <ViolatingJobTable table={violatingJobTable}/>
            </VStack>
          </Card.Body>
        </Card.Root>
      </VStack>
    </>
  )
}
