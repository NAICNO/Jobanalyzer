import { Outlet } from 'react-router'
import { Grid, GridItem } from '@chakra-ui/react'

import { AppHeader, Sidebar } from '../components'
import { useColorMode } from '../components/ui/color-mode.tsx'
import { useCluster } from '../hooks/useCluster'

export default function RootLayout() {
  const colorMode = useColorMode()
  const { hasSelectedClusters } = useCluster()

  const mainGridItemBackgroundColor = colorMode.colorMode === 'light' ? 'white' : 'gray.800'
  const sidebarBackgroundColor = colorMode.colorMode === 'light' ? 'gray.100' : 'gray.700'

  return (
    <Grid
      templateAreas={{
        base: '"header" "main"',
        md: hasSelectedClusters ? '"header header" "nav main"' : '"header" "main"',
      }}
      gridTemplateRows={{
        base: '60px 1fr',
        md: '60px 1fr',
      }}
      gridTemplateColumns={{
        base: '1fr',
        md: hasSelectedClusters ? '195px 1fr' : '1fr',
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
        <AppHeader />
      </GridItem>
      {hasSelectedClusters && (
        <GridItem
          px="10px"
          pt="10px"
          area={'nav'}
          bg={sidebarBackgroundColor}
          display="flex"
          flexDirection="column"
          h="100%"
        >
          <Sidebar />
        </GridItem>
      )}
      <GridItem
        paddingX="20px"
        paddingY="10px"
        area={'main'}
        bg={mainGridItemBackgroundColor}
        overflowY="auto"
        overflowX="hidden"
      >
        <Outlet/>
      </GridItem>
    </Grid>
  )
}
