import { useMemo, useState } from 'react'
import {
  Heading,
  HStack,
  SlideFade,
  Text,
  VStack
} from '@chakra-ui/react'
import { Navigate, useParams } from 'react-router-dom'
import {
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable
} from '@tanstack/react-table'

import { EMPTY_ARRAY } from '../Constants.ts'
import { findCluster } from '../util'
import { getDeadWeightTableColumns } from '../util/TableUtils.ts'
import { NavigateBackButton } from '../components/NavigateBackButton.tsx'
import DeadWeightTable from '../components/table/DeadWeightTable.tsx'
import { useFetchDeadWeight } from '../hooks/useFetchDeadWeight.ts'
import PageTitle from '../components/PageTitle.tsx'

export default function DeadWeightPage() {
  const {clusterName} = useParams<string>()

  const cluster = findCluster(clusterName)

  if(!cluster) {
    return (
      <Navigate to="/"/>
    )
  }

  const {data, isFetched} = useFetchDeadWeight(clusterName!)

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

  return (
    <>
      <PageTitle title={`${cluster.name} Deadweight`}/>
      <VStack alignItems={'start'}>
        <HStack mb="20px">
          <NavigateBackButton/>
          <Heading ml="20px">{cluster.name} dead weight</Heading>
        </HStack>
        <Text mb="20px">
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
