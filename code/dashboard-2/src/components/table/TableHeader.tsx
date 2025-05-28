import { chakra, Table } from '@chakra-ui/react'
import { flexRender } from '@tanstack/react-table'
import { Table as TableType } from '@tanstack/react-table'
import { GoTriangleDown, GoTriangleUp } from 'react-icons/go'

interface TableHeaderProps<T> {
  table: TableType<T>
}

export const TableHeader = ({table}: TableHeaderProps<any>) => {
  return (
    <Table.Header>
      {table.getHeaderGroups().map((headerGroup) => (
        <Table.Row
          key={headerGroup.id}
          bg="bg.subtle"
          position="sticky"
          top={headerGroup.depth === 0 ? 0 : '2rem'}
          zIndex={headerGroup.depth === 0 ? 2 : 1}
        >
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
              <Table.ColumnHeader
                padding={2}
                key={header.id}
                onClick={header.column.getToggleSortingHandler()}
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
                        <GoTriangleDown aria-label="sorted descending"/>
                        :
                        <GoTriangleUp aria-label="sorted ascending"/>

                    ) : null
                  }
                </chakra.span>
              </Table.ColumnHeader>
            )
          })}
        </Table.Row>
      ))}
    </Table.Header>
  )
}
