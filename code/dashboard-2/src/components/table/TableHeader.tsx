import { chakra, Th, Thead, Tr } from '@chakra-ui/react'
import { flexRender } from '@tanstack/react-table'
import { TriangleDownIcon, TriangleUpIcon } from '@chakra-ui/icons'
import { Table as TableType } from '@tanstack/react-table'

interface TableHeaderProps<T> {
  table: TableType<T>
}

export const TableHeader = ({table}: TableHeaderProps<any>) => {
  return (
    <Thead>
      {table.getHeaderGroups().map((headerGroup) => (
        <Tr key={headerGroup.id}>
          {headerGroup.headers.map((header,) => {
            const meta: any = header.column.columnDef.meta
            const columnRelativeDepth = header.depth - header.column.depth
            if (
              !header.isPlaceholder &&
              columnRelativeDepth > 1 &&
              header.id === header.column.id
            ) {
              return null
            }

            let rowSpan = 1
            if(header.isPlaceholder) {
              const leafs = header.getLeafHeaders()
              rowSpan = leafs[leafs.length - 1].depth - header.depth
            }

            return (
              <Th
                padding={2}
                key={header.id}
                onClick={header.column.getToggleSortingHandler()}
                isNumeric={meta?.isNumeric}
                colSpan={header.colSpan}
                {...(header.depth > 1 ? { 'data-is-grouped-column-sub-header':true } : {})}
                {...(header.colSpan > 1 ? { 'data-is-grouped-column-header':true } : {})}
                rowSpan={rowSpan}
                style={{
                  minWidth: header.column.columnDef.minSize,
                }}
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
