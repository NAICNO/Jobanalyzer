import {
  Drawer,
  DrawerBody,
  DrawerCloseButton,
  DrawerContent,
  DrawerHeader,
  DrawerOverlay,
  HStack,
  Heading,
  Link as ChakraLink,
} from '@chakra-ui/react'
import { Link as ReactRouterLink } from 'react-router-dom'
import { ExternalLinkIcon } from '@chakra-ui/icons'

import { NodeSelectionHelpContent } from './NodeSelectionHelpContent.tsx'

interface NodeSelectionHelpDrawerProps {
  isOpen: boolean
  onClose: () => void
  finalFocusRef: any
}

export const NodeSelectionHelpDrawer = ({isOpen, onClose, finalFocusRef}: NodeSelectionHelpDrawerProps) => {
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
