import { HamburgerIcon } from '@chakra-ui/icons'
import { IconButton } from '@chakra-ui/react'

interface HamburgerButtonProps {
  opOpenSidebarDrawer: () => void
}

export const HamburgerButton = ({opOpenSidebarDrawer}: HamburgerButtonProps) => {
  return (
    <IconButton
      variant={'subtle'}
      display={{base: 'inline-flex', md: 'none'}}
      mr="20px"
      icon={<HamburgerIcon/>}
      onClick={opOpenSidebarDrawer}
      aria-label="Open menu"
    />
  )
}
