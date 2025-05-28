import React from 'react'
import {
  Box,
  Drawer,
  HStack,
  List,
  Separator,
  Spacer,
  VStack,
  useBreakpointValue, DrawerOpenChangeDetails,
} from '@chakra-ui/react'
import { NavLink, useLocation } from 'react-router'

import { LightDarkModeButton } from './LightDarkModeButton.tsx'
import { SIDEBAR_ITEMS } from '../Constants.ts'
import { useColorMode } from './ui/color-mode.tsx'

interface SidebarProps {
  onCloseDrawer: () => void
  isDrawerOpen: boolean
}

export const Sidebar = ({onCloseDrawer, isDrawerOpen}: SidebarProps) => {

  const isDrawer = useBreakpointValue({base: true, md: false})

  const handleOnOpenChange = (details: DrawerOpenChangeDetails) => {
    if (!details.open) {
      onCloseDrawer()
    }
  }

  return (
    <>
      {isDrawer ?
        <Drawer.Root open={isDrawerOpen} placement="end" onOpenChange={handleOnOpenChange}>
          <Drawer.Backdrop/>
          <Drawer.Content>
            <Drawer.CloseTrigger/>
            <Drawer.Header>Menu</Drawer.Header>
            <Drawer.Body>
              <SideBarContent onCloseDrawer={onCloseDrawer}/>
              <VStack
                pt="50px"
                gap={{base: '10px', md: '10px', lg: '20px'}}
                alignItems={'start'}
              >
                <HStack w={'100%'} pt="10px">
                  <Spacer/>
                  <LightDarkModeButton/>
                </HStack>
              </VStack>
            </Drawer.Body>
          </Drawer.Content>
        </Drawer.Root>
        :
        <SideBarContent onCloseDrawer={onCloseDrawer}/>
      }
    </>
  )
}

const SideBarContent = ({onCloseDrawer}: { onCloseDrawer: () => void }) => {

  const {colorMode} = useColorMode()
  const hoverBgColor = colorMode === 'light' ? 'gray.200' : 'blue.500'
  const activeBgColor = colorMode === 'light' ? 'gray.300' : 'blue.600'

  return (
    <List.Root variant={'plain'} fontSize={{base: '1em', md: '1.2em'}} gap="1">
      {SIDEBAR_ITEMS.map((item, index) => {
        if (item.type === 'separator') {
          return (
            <List.Item key={index}>
              <Separator variant={'solid'}/>
            </List.Item>
          )
        }

        const location = useLocation()
        const {pathname} = location

        const isActive = (path: string) => {
          return pathname.includes(path)
        }
        return (
          <List.Item
            key={index}
          >
            <NavLink to={item.path} onClick={onCloseDrawer} style={{width: '100%', display: 'block'}}>
              <Box
                _hover={{bg: hoverBgColor}}
                bg={isActive(item.matches) ? activeBgColor : 'transparent'}
                px={{base: '12px', md: '10px'}}
                py="6px"
                borderRadius="md"
              >
                <List.Indicator asChild>
                  {React.createElement(item.icon)}
                </List.Indicator>
                {item.text}
              </Box>
            </NavLink>
          </List.Item>
        )
      })}
    </List.Root>
  )
}
