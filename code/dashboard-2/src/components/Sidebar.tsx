import {
  Box,
  Divider,
  Drawer,
  DrawerBody,
  DrawerCloseButton,
  DrawerContent,
  DrawerHeader,
  DrawerOverlay, HStack,
  List,
  ListIcon,
  ListItem,
  Spacer,
  useBreakpointValue,
  useColorMode, VStack,
} from '@chakra-ui/react'
import { NavLink, useLocation } from 'react-router-dom'

import { LightDarkModeButton } from './LightDarkModeButton.tsx'
import { SIDEBAR_ITEMS } from '../Constants.ts'

interface SidebarProps {
  onCloseDrawer: () => void
  isDrawerOpen: boolean
}

export default function Sidebar({onCloseDrawer, isDrawerOpen}: SidebarProps) {

  const isDrawer = useBreakpointValue({base: true, md: false})

  return (
    <>
      {isDrawer ?
        <Drawer isOpen={isDrawerOpen} placement="left" onClose={onCloseDrawer}>
          <DrawerOverlay/>
          <DrawerContent>
            <DrawerCloseButton/>
            <DrawerHeader>Menu</DrawerHeader>
            <DrawerBody>
              <SideBarContent onCloseDrawer={onCloseDrawer}/>
              <VStack
                pt="50px"
                spacing={{base: '10px', md: '10px', lg: '20px'}}
                alignItems={'start'}
              >
                <HStack w={'100%'} pt="10px">
                  <Spacer/>
                  <LightDarkModeButton/>
                </HStack>
              </VStack>
            </DrawerBody>
          </DrawerContent>
        </Drawer>
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
    <List fontSize={{base: '1em', md: '1.2em'}} spacing="1">
      {SIDEBAR_ITEMS.map((item, index) => (
        <SidebarItem
          key={index}
          type={item.type}
          path={item.path}
          matches={item.matches}
          text={item.text}
          icon={item.icon}
          hoverBgColor={hoverBgColor}
          activeBgColor={activeBgColor}
          onCloseDrawer={onCloseDrawer}
        />
      ))}
    </List>
  )
}

const SidebarItem = ({type, path = '', matches = '', text = '', icon, hoverBgColor, activeBgColor, onCloseDrawer}: {
  type: string,
  path: string | undefined,
  matches: string | undefined,
  text: string | undefined,
  icon: any,
  hoverBgColor: string,
  activeBgColor: string,
  onCloseDrawer: () => void
}) => {

  if (type === 'separator') {
    return <Divider my="20px"/>
  } else {

    const location = useLocation()
    const {pathname} = location

    const isActive = (path: string) => {
      return pathname.includes(path)
    }

    return (
      <Box>
        <NavLink to={path} onClick={onCloseDrawer}>
          <ListItem
            _hover={{bg: hoverBgColor}}
            bg={isActive(matches) ? activeBgColor : 'transparent'}
            px={{base: '12px', md: '10px'}}
            py="6px"
            borderRadius="md"
          >
            <ListIcon as={icon}/>
            {text}
          </ListItem>
        </NavLink>
      </Box>
    )
  }
}
