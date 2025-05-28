import { Outlet } from 'react-router'
import { Grid, GridItem, useDisclosure } from '@chakra-ui/react'

import { AppHeader, Sidebar } from '../components'
import { useColorMode } from '../components/ui/color-mode.tsx'

export default function RootLayout() {

  const {open: isDrawerOpen, onOpen: onOpenDrawer, onClose: onCloseDrawer} = useDisclosure()

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
        md: '60px 1fr',
      }}
      gridTemplateColumns={{
        base: '1fr',
        md: '160px 1fr',
      }}
      gap="1"
      h="100vh"
    >
      <GridItem
        px={{base: '20px', md: '20px'}}
        py={{base: '10px', md: '10px'}}
        area={'header'}
        boxShadow={'md'}
      >
        <AppHeader opOpenSidebarDrawer={onOpenDrawer}/>
      </GridItem>
      <GridItem
        px="10px"
        pt="10px"
        area={'nav'}
        bg={sidebarBackgroundColor}
        display={{base: isDrawerOpen ? 'block' : 'none', md: 'block'}}
      >
        <Sidebar isDrawerOpen={isDrawerOpen} onCloseDrawer={onCloseDrawer}/>
      </GridItem>
      <GridItem
        paddingX="20px"
        paddingY="10px"
        area={'main'}
        bg={mainGridItemBackgroundColor}
      >
        <Outlet/>
      </GridItem>
    </Grid>
  )
}
