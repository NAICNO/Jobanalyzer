import {
  Drawer,
  DrawerBody,
  DrawerCloseButton,
  DrawerContent,
  DrawerHeader,
  DrawerOverlay, Heading, HStack,
  Link as ChakraLink,
} from '@chakra-ui/react'
import { Link as ReactRouterLink } from 'react-router-dom'

import NodeSelectionHelpContent from './NodeSelectionHelpContent.tsx'
import { ExternalLinkIcon } from '@chakra-ui/icons'

interface NodeSelectionHelpDrawerProps {
  isOpen: boolean
  onClose: () => void
  finalFocusRef: any
}

const NodeSelectionHelpDrawer = ({isOpen, onClose, finalFocusRef}: NodeSelectionHelpDrawerProps) => {
  return (
    <Drawer
      isOpen={isOpen}
      placement="right"
      onClose={onClose}
      finalFocusRef={finalFocusRef}
      size="xl"
    >
      <DrawerOverlay/>
      <DrawerContent>
        <DrawerCloseButton/>
        <DrawerHeader>
          <HStack>
            <Heading>
              Query help
            </Heading>
            <ChakraLink as={ReactRouterLink} to="/dashboard/help/node-selection" isExternal onClick={onClose}>
              <ExternalLinkIcon mx="10px"/>
            </ChakraLink>
          </HStack>
        </DrawerHeader>

        <DrawerBody>
          <NodeSelectionHelpContent/>
        </DrawerBody>
      </DrawerContent>
    </Drawer>
  )
}

export default NodeSelectionHelpDrawer
