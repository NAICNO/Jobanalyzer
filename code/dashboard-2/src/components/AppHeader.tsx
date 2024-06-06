import {
  Flex,
  Heading,
  HStack, Image, Show,
  Spacer,
  useColorMode,
} from '@chakra-ui/react'
import { NavLink } from 'react-router-dom'

import { APP_NAME } from '../Constants.ts'
import { LightDarkModeButton } from './LightDarkModeButton.tsx'
import { HamburgerButton } from './HamburgerButton.tsx'

interface AppHeaderProps {
  opOpenSidebarDrawer: () => void
}

export default function AppHeader({opOpenSidebarDrawer}: AppHeaderProps) {

  const {colorMode} = useColorMode()

  const naicLogo = colorMode === 'light' ? '/images/naic/naic_dark.svg' : '/images/naic/naic_light.svg'

  return (
    <Flex
      as="header"
      align="center"
      wrap="wrap"
    >
      <HamburgerButton opOpenSidebarDrawer={opOpenSidebarDrawer}/>
      <NavLink to={''}>
        <Flex alignItems={'center'}>
          <Image
            src={naicLogo}
            alt="NAIC logo"
            w={{base: '60px', md: '80px'}}
            mr="20px"
            mb="5px"
          />
          <Heading as="h1" size={{base: 'md', md: 'xl'}}>
            {APP_NAME}
          </Heading>
        </Flex>
      </NavLink>
      <Spacer/>
      <Show above={'base'}>
        <HStack
          spacing={{base: '10px', md: '10px', lg: '20px'}}
          display={{base: 'none', md: 'flex'}}
          alignItems={'center'}
        >
          <LightDarkModeButton/>
        </HStack>
      </Show>
    </Flex>
  )
}
