import { Table } from '@chakra-ui/react'

interface JobBasicInfo {
  jobId: string | null
  user: string | null
  clusterName: string | null
  hostname: string | null
}

export function JobBasicInfoTable({
  jobId,
  user,
  clusterName,
  hostname,
}: JobBasicInfo) {
  return (
    <Table.ScrollArea borderWidth="1px">
      <Table.Root size={'sm'} showColumnBorder variant="outline">
        <Table.Header>
          <Table.Row>
            <Table.ColumnHeader>Job #</Table.ColumnHeader>
            <Table.ColumnHeader>User</Table.ColumnHeader>
            <Table.ColumnHeader>Cluster</Table.ColumnHeader>
            <Table.ColumnHeader>Node</Table.ColumnHeader>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          <Table.Row>
            <Table.Cell>{jobId}</Table.Cell>
            <Table.Cell>{user}</Table.Cell>
            <Table.Cell>{clusterName}</Table.Cell>
            <Table.Cell>{hostname}</Table.Cell>
          </Table.Row>
        </Table.Body>
      </Table.Root>
    </Table.ScrollArea>
  )
}
