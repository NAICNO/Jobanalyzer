import { useMemo, useState } from 'react'
import {
  Box,
  Card,
  CardBody,
  Heading,
  HStack,
  List,
  ListIcon,
  ListItem,
  SlideFade,
  Text,
  UnorderedList,
  VStack
} from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'
import { WarningTwoIcon } from '@chakra-ui/icons'
import moment from 'moment-timezone'

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

  const {data: allJobsOfUser, isFetched} = useFetchViolator(cluster, violator)

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
        <Card variant={'outline'}>
          <CardBody>
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
              <List ml="20px">
                {
                  violatedPolicies.map(policy => {
                    if (!policy) {
                      return null
                    }
                    return (
                      <ListItem key={policy.name}>
                        <ListIcon as={WarningTwoIcon} color="red.500"/>
                        {policy.name}
                        <Box ml="20px" mt="5px">
                          <UnorderedList>
                            <ListItem>Trigger: {policy.trigger}</ListItem>
                            <ListItem>Problem: {policy.problem}</ListItem>
                            <ListItem>Remedy: {policy.remedy}</ListItem>
                          </UnorderedList>
                        </Box>
                      </ListItem>
                    )
                  })
                }
              </List>
              <Text marginY="20px">(Times below are UTC, job numbers are derived from session leader if not running
                under
                Slurm)</Text>
              <SlideFade in={isFetched}>
                <ViolatingJobTable table={violatingJobTable}/>
              </SlideFade>
            </VStack>
          </CardBody>
        </Card>
      </VStack>
    </>
  )
}
