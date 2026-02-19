import { CloseButton, Drawer, DrawerOpenChangeDetails, Portal, SimpleGrid, VStack } from '@chakra-ui/react'

import { NodeInfoSummary } from './NodeInfoSummary'
import { NodeGpuTable } from './NodeGpuTable'
import { NodeErrorMessages } from './NodeErrorMessages'
import { NodeStates } from './NodeStates'
import { NodeCpuMemoryTimeseries } from './NodeCpuMemoryTimeseries'
import { NodeGpuTimeseries } from './NodeGpuTimeseries'
import { NodeDiskstatsTimeseries } from './NodeDiskstatsTimeseries'
import { NodeTopology } from './NodeTopology'

interface NodeDetailDrawerProps {
  cluster: string
  nodename: string | undefined
  onClose: () => void
}

export const NodeDetailDrawer = ({ cluster, nodename, onClose }: NodeDetailDrawerProps) => {
  const handleOpenChange = (details: DrawerOpenChangeDetails) => {
    if (!details.open) {
      onClose()
    }
  }

  return (
    <Drawer.Root
      open={!!nodename}
      onOpenChange={handleOpenChange}
      placement="end"
      size="full"
      lazyMount
      unmountOnExit
    >
      <Portal>
        <Drawer.Backdrop />
        <Drawer.Positioner>
          <Drawer.Content maxW="75vw">
            <Drawer.CloseTrigger />
            <Drawer.Header>
              <Drawer.Title>{nodename} Details</Drawer.Title>
            </Drawer.Header>
            <Drawer.Body>
              {nodename && (
                <VStack gap={4} align="start">
                  <SimpleGrid columns={{ base: 1, lg: 2 }} gap={4} w="100%">
                    <NodeInfoSummary cluster={cluster} nodename={nodename} />
                    <NodeStates cluster={cluster} nodename={nodename} />
                  </SimpleGrid>
                  <NodeGpuTable cluster={cluster} nodename={nodename} />
                  <NodeErrorMessages cluster={cluster} nodename={nodename} />
                  <NodeCpuMemoryTimeseries cluster={cluster} nodename={nodename} />
                  <NodeGpuTimeseries cluster={cluster} nodename={nodename} />
                  <NodeDiskstatsTimeseries cluster={cluster} nodename={nodename} />
                  <NodeTopology cluster={cluster} nodename={nodename} />
                </VStack>
              )}
            </Drawer.Body>
            <Drawer.CloseTrigger asChild>
              <CloseButton size="sm" />
            </Drawer.CloseTrigger>
          </Drawer.Content>
        </Drawer.Positioner>
      </Portal>
    </Drawer.Root>
  )
}
