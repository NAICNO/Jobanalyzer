import {
  Flex,
  Heading,
  HStack,
  Image,
  Show,
  Spacer,
  Text,
} from '@chakra-ui/react'
import { NavLink, useParams } from 'react-router'
import { LuUser } from 'react-icons/lu'

import { APP_NAME } from '../Constants.ts'
import { LightDarkModeButton } from './LightDarkModeButton.tsx'
import { useColorMode } from './ui/color-mode.tsx'
import { useAuth } from '../hooks/useAuth'
import { decodeJwt } from '../utils/cryptoUtils'
import { getClusterConfig } from '../config/clusters'

export const AppHeader = () => {
  const {colorMode} = useColorMode()
  const { clusterName } = useParams<{ clusterName: string }>()
  const { authState } = useAuth()

  const naicLogo = colorMode === 'light' ? '/images/naic/naic_dark.svg' : '/images/naic/naic_light.svg'

  // Extract user ID from ID token only if:
  // 1. We're on a cluster route (clusterName exists in URL params)
  // 2. The cluster requires authentication
  // 3. User is authenticated
  let userId: string | null = null
  if (clusterName) {
    const clusterConfig = getClusterConfig(clusterName)
    if (clusterConfig?.requiresAuth && authState[clusterName]?.user?.id_token) {
      const decoded = decodeJwt(authState[clusterName].user.id_token)
      userId = decoded?.user || decoded?.preferred_username || decoded?.email || null
    }
  }

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
          {userId && (
            <HStack gap={2} px={3} py={1} bg="gray.100" _dark={{ bg: 'gray.700' }} rounded="md">
              <LuUser size={16} />
              <Text fontSize="sm" fontWeight="medium">{userId}</Text>
            </HStack>
          )}
          <LightDarkModeButton/>
        </HStack>
      </Show>
    </Flex>
  )
}
