import { useMemo, useState } from 'react'
import {
  Heading,
  HStack,
  SlideFade,
  Text,
  VStack
} from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'

import { EMPTY_ARRAY } from '../Constants.ts'
import { findCluster } from '../util'
import { getDeadWeightTableColumns } from '../util/TableUtils.ts'
import { NavigateBackButton, PageTitle } from '../components'
import { DeadWeightTable } from '../components/table'
import { useFetchDeadWeight } from '../hooks/useFetchDeadWeight.ts'


export default function DeadWeightPage() {
  const {clusterName, hostname} = useParams<string>()

  const cluster = findCluster(clusterName)

  if (!cluster) {
    return (
      <Navigate to="/"/>
    )
  }

  const filter = {
    afterDate: null,
    hostname: hostname || null
  }

  const {data, isFetched} = useFetchDeadWeight(cluster, filter)

  const deadWeightJobTableColumns = useMemo(() => getDeadWeightTableColumns(), [cluster])
  const [sorting, setSorting] = useState<SortingState>([])

  const deadWeightTable = useReactTable({
    columns: deadWeightJobTableColumns,
    data: data || EMPTY_ARRAY,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: {
      sorting: sorting,
    }
  })

  const pageTitle = `${cluster.name} Deadweight${hostname ? ` - ${hostname}` : ''}`

  return (
    <>
      <PageTitle title={pageTitle}/>
      <VStack alignItems={'start'}>
        <HStack mb={3}>
          <NavigateBackButton/>
          <Heading ml={2} size={{base: 'md', md: 'lg'}}>{pageTitle}</Heading>
        </HStack>
        <Text mb={3}>
          The following processes and jobs are zombies or defuncts or
          otherwise dead and may be bogging down the system. The list is
          recomputed at noon and midnight and goes back four weeks.
        </Text>
        <SlideFade in={isFetched}>
          <DeadWeightTable table={deadWeightTable}/>
        </SlideFade>
      </VStack>
    </>
  )
}
