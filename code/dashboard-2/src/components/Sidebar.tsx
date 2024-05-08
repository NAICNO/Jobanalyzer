import {
  Box,
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
import { GrNodes, GrServers } from 'react-icons/gr'
import { GiFox } from 'react-icons/gi'

import { LightDarkModeButton } from './LightDarkModeButton.tsx'

const sidebarItems = [
  {
    path: '/dashboard/ml',
    matches: '/ml',
    text: 'ML Nodes',
    icon: GrNodes
  },
  {
    path: '/dashboard/fox',
    matches: '/fox',
    text: 'Fox',
    icon: GiFox
  },
  {
    path: '/dashboard/saga',
    matches: '/saga',
    text: 'Saga',
    icon: GrServers
  }
]

interface SidebarProps {
  onCloseDrawer: () => void
  isDrawerOpen: boolean
}

export default function Sidebar({onCloseDrawer, isDrawerOpen}: SidebarProps) {

  const {colorMode} = useColorMode()
  const isDrawer = useBreakpointValue({base: true, md: false})

  const hoverBgColor = colorMode === 'light' ? 'gray.200' : 'blue.500'
  const activeBgColor = colorMode === 'light' ? 'gray.300' : 'blue.600'

  const SideBarContent = () => (
    <List fontSize={{base: '1em', md: '1.2em'}} spacing="1">
      {sidebarItems.map((item, index) => (
        <SidebarItem
          key={index}
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

  return (
    <>
      {isDrawer ?
        <Drawer isOpen={isDrawerOpen} placement="left" onClose={onCloseDrawer}>
          <DrawerOverlay/>
          <DrawerContent>
            <DrawerCloseButton/>
            <DrawerHeader>Menu</DrawerHeader>
            <DrawerBody>
              <SideBarContent/>
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
        <SideBarContent/>
      }
    </>
  )
}

const SidebarItem = ({path, matches, text, icon, hoverBgColor, activeBgColor, onCloseDrawer}: {
  path: string,
  matches: string,
  text: string,
  icon: any,
  hoverBgColor: string,
  activeBgColor: string,
  onCloseDrawer: () => void
}) => {

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
          px={{base: '12px', md: '16px'}}
          py="8px"
          borderRadius="md"
        >
          <ListIcon as={icon}/>
          {text}
        </ListItem>
      </NavLink>
    </Box>
  )
}
