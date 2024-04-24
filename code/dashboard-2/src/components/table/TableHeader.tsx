import { chakra, Th, Thead, Tr } from '@chakra-ui/react'
import { flexRender } from '@tanstack/react-table'
import { TriangleDownIcon, TriangleUpIcon } from '@chakra-ui/icons'
import { Table as TableType } from '@tanstack/react-table'

interface TableHeaderProps<T> {
  table: TableType<T>
}

const TableHeader = ({table}: TableHeaderProps<any>) => {
  return (
    <Thead borderBottom="1px solid" borderColor="gray.200">
      {table.getHeaderGroups().map((headerGroup) => (
        <Tr key={headerGroup.id}>
          {headerGroup.headers.map((header,) => {
            const meta: any = header.column.columnDef.meta
            return (
              <Th
                key={header.id}
                borderRight={header.index !== headerGroup.headers.length - 1 ? '1px solid' : 'none'}
                borderColor="gray.200"
                onClick={header.column.getToggleSortingHandler()}
                isNumeric={meta?.isNumeric}
                colSpan={header.colSpan}
                style={{textTransform: 'none'}}
                title={meta?.helpText}
              >
                {
                  flexRender(header.column.columnDef.header, header.getContext())
                }
                <chakra.span>
                  {
                    header.column.getIsSorted() ? (
                      header.column.getIsSorted() === 'desc' ?
                        <TriangleDownIcon aria-label="sorted descending"/>
                        :
                        <TriangleUpIcon aria-label="sorted ascending"/>

                    ) : null
                  }
                </chakra.span>
              </Th>
            )
          })}
        </Tr>
      ))}
    </Thead>
  )
}

export default TableHeader
