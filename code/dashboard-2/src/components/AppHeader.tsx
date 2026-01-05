import {
  Flex,
  Heading,
  HStack, Image, Show,
  Spacer,
} from '@chakra-ui/react'
import { NavLink } from 'react-router'

import { APP_NAME } from '../Constants.ts'
import { LightDarkModeButton } from './LightDarkModeButton.tsx'
import { useColorMode } from './ui/color-mode.tsx'

export const AppHeader = () => {
  const {colorMode} = useColorMode()

  const naicLogo = colorMode === 'light' ? '/images/naic/naic_dark.svg' : '/images/naic/naic_light.svg'

  return (
    <Flex
      as="header"
      align="center"
      wrap="wrap"
    >
      <NavLink to={'/'}>
        <Flex alignItems={'center'}>
          <Image
            src={naicLogo}
            alt="NAIC logo"
            w={{base: '40px', md: '60px'}}
            mr="20px"
            mb="5px"
          />
          <Heading as="h1" size={{base: 'lg', md: '2xl'}}>
            {APP_NAME}
          </Heading>
        </Flex>
      </NavLink>
      <Spacer/>
      <Show when={'base'}>
        <HStack
          gap={{base: '10px', md: '10px', lg: '20px'}}
          display={{base: 'none', md: 'flex'}}
          alignItems={'center'}
        >
          <LightDarkModeButton/>
        </HStack>
      </Show>
    </Flex>
  )
}
