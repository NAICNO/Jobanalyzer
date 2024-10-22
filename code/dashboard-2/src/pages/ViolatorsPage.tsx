import { useMemo, useState } from 'react'
import { Heading, HStack, SlideFade, Text, VStack } from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router-dom'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'

import { EMPTY_ARRAY } from '../Constants.ts'
import { findCluster } from '../util'
import { useFetchViolations } from '../hooks/useFetchViolations.ts'
import { getViolatingJobTableColumns, getViolatingUserTableColumns } from '../util/TableUtils.ts'
import { ViolatingUserTable, ViolatingJobTable } from '../components/table'
import { NavigateBackButton, PageTitle } from '../components'

export default function ViolatorsPage() {
  const {clusterName} = useParams<string>()

  const cluster = findCluster(clusterName)

  if (!cluster) {
    return (
      <Navigate to="/"/>
    )
  }

  const {data, isFetched} = useFetchViolations(cluster)

  const violatingUserTableColumns = useMemo(() => getViolatingUserTableColumns(), [cluster])
  const [violatingUserTableSorting, setViolatingUserTableSorting] = useState<SortingState>([])

  const violatingUserTable = useReactTable({
    columns: violatingUserTableColumns,
    data: data?.byUser || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setViolatingUserTableSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: violatingUserTableSorting,
    }
  })

  const violatingJobTableColumns = useMemo(() => getViolatingJobTableColumns(), [cluster])
  const [violatingJobTableSorting, setViolatingJobTableSorting] = useState<SortingState>([])

  const violatingJobTable = useReactTable({
    columns: violatingJobTableColumns,
    data: data?.byJob || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setViolatingJobTableSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: violatingJobTableSorting,
    }
  })

  return (
    <>
      <PageTitle title={`${cluster.name} Policy Violators`}/>
      <VStack alignItems={'start'}>
        <HStack mb="10px">
          <NavigateBackButton/>
          <Heading ml={2} size={{base: 'md', md: 'lg'}}>{cluster.name} policy violators</Heading>
        </HStack>
        <Text>The following users and jobs have been running significantly outside of policy and are probably
          not appropriate to run on this cluster. The list is recomputed at noon and midnight
          and goes back four weeks.</Text>
        <Heading as="h4" size={{base: 'md', md: 'lg'}} mt={4}>
          By user
        </Heading>
        <SlideFade in={isFetched}>
          <ViolatingUserTable table={violatingUserTable}/>
        </SlideFade>
        <Heading as="h4" size="lg" mt="20px">
          By job and time
        </Heading>
        <SlideFade in={isFetched}>
          <ViolatingJobTable table={violatingJobTable}/>
        </SlideFade>
      </VStack>
    </>
  )
}
