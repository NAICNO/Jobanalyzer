import {
  CloseButton,
  Drawer,
  DrawerOpenChangeDetails,
  HStack,
  Heading,
  Link as ChakraLink,
  Portal,
} from '@chakra-ui/react'
import { Link as ReactRouterLink } from 'react-router'
import { LuExternalLink } from 'react-icons/lu'

import { NodeSelectionHelpContent } from './NodeSelectionHelpContent.tsx'

interface NodeSelectionHelpDrawerProps {
  isOpen: boolean
  onClose: () => void
  finalFocusRef: any
}

export const NodeSelectionHelpDrawer = ({isOpen, onClose, finalFocusRef}: NodeSelectionHelpDrawerProps) => {

  const handleOpenChange = (details: DrawerOpenChangeDetails) => {
    if (!details.open) {
      onClose()
    }
  }

  return (
    <Drawer.Root
      open={isOpen}
      onOpenChange={handleOpenChange}
      placement="end"
      finalFocusEl={finalFocusRef}
      size="xl"
    >
      <Portal>
        <Drawer.Backdrop/>
        <Drawer.Positioner>
          <Drawer.Content>
            <Drawer.CloseTrigger/>
            <Drawer.Header>
              <HStack>
                <Heading>
                  Query help
                </Heading>
                <ChakraLink asChild onClick={onClose} target="_blank">
                  <ReactRouterLink to={'/dashboard/help/node-selection'}>
                    <LuExternalLink/>
                  </ReactRouterLink>
                </ChakraLink>
              </HStack>
            </Drawer.Header>
            <Drawer.Body>
              <NodeSelectionHelpContent/>
            </Drawer.Body>
            <Drawer.CloseTrigger asChild>
              <CloseButton size={'sm'}/>
            </Drawer.CloseTrigger>
          </Drawer.Content>
        </Drawer.Positioner>
      </Portal>
    </Drawer.Root>
  )
}
