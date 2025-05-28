import { IconButton } from '@chakra-ui/react'
import { FiMoon, FiSun } from 'react-icons/fi'

import { useColorMode } from './ui/color-mode'

export const LightDarkModeButton = () => {

  const {colorMode, toggleColorMode} = useColorMode()

  return (
    <IconButton
      size={{base: 'sm', md: 'sm', lg: 'sm'}}
      aria-label="switch theme"
      onClick={toggleColorMode}
      variant={'outline'}
    >
      {
        colorMode === 'light' ? <FiMoon/> : <FiSun/>
      }
    </IconButton>
  )
}
