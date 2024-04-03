import { MoonIcon, SunIcon } from '@chakra-ui/icons'
import { IconButton, useColorMode } from '@chakra-ui/react'

export const LightDarkModeButton = () => {

  const {colorMode, toggleColorMode} = useColorMode()

  return (
    <IconButton
      size={{base: 'sm', md: 'sm', lg: 'md'}}
      aria-label="switch theme"
      icon={colorMode === 'light' ? <MoonIcon/> : <SunIcon/>}
      onClick={toggleColorMode}
    />
  )
}
