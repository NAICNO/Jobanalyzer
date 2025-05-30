import { IconButton } from '@chakra-ui/react'
import { MdMenu } from 'react-icons/md'

interface HamburgerButtonProps {
  opOpenSidebarDrawer: () => void
}

export const HamburgerButton = ({opOpenSidebarDrawer}: HamburgerButtonProps) => {
  return (
    <IconButton
      variant={'subtle'}
      display={{base: 'inline-flex', md: 'none'}}
      mr="20px"
      onClick={opOpenSidebarDrawer}
      aria-label="Open menu"
    >
      <MdMenu/>
    </IconButton>
  )
}
