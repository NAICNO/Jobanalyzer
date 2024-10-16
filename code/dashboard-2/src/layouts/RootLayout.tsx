import { Outlet } from 'react-router-dom'
import { Grid, GridItem, useColorMode, useDisclosure } from '@chakra-ui/react'

import { AppHeader, Sidebar } from '../components'

export default function RootLayout() {

  const {isOpen: isDrawerOpen, onOpen: onOpenDrawer, onClose: onCloseDrawer} = useDisclosure()

  const colorMode = useColorMode()

  const mainGridItemBackgroundColor = colorMode.colorMode === 'light' ? 'white' : 'gray.800'
  const sidebarBackgroundColor = colorMode.colorMode === 'light' ? 'gray.100' : 'gray.700'

  return (
    <Grid
      templateAreas={{
        base: '"header" "main"',
        md: '"header header" "nav main"',
      }}
      gridTemplateRows={{
        base: '60px 1fr',
        md: '80px 1fr',
      }}
      gridTemplateColumns={{
        base: '1fr',
        md: '180px 1fr',
      }}
      gap="1"
      h="100vh"
    >
      <GridItem
        px={{base: '20px', md: '40px'}}
        py={{base: '10px', md: '20px'}}
        area={'header'}
        boxShadow={'md'}
      >
        <AppHeader opOpenSidebarDrawer={onOpenDrawer}/>
      </GridItem>
      <GridItem
        px="20px"
        pt="20px"
        area={'nav'}
        bg={sidebarBackgroundColor}
        display={{base: isDrawerOpen ? 'block' : 'none', md: 'block'}}
      >
        <Sidebar isDrawerOpen={isDrawerOpen} onCloseDrawer={onCloseDrawer}/>
      </GridItem>
      <GridItem
        paddingX="20px"
        paddingY="20px"
        area={'main'}
        bg={mainGridItemBackgroundColor}
      >
        <Outlet/>
      </GridItem>
    </Grid>
  )
}
