import {
  Box,
  Clipboard,
  HStack,
  IconButton,
  Input,
  InputGroup,
  Popover,
} from '@chakra-ui/react'
import { FaLink } from 'react-icons/fa'

interface ShareLinkPopoverProps {
  link: string
  text: string
}

export const ShareLinkPopover = ({link, text}: ShareLinkPopoverProps) => {

  return (
    <Popover.Root positioning={{placement: 'bottom-start'}}>
      <Popover.Trigger asChild>
        <IconButton bg={'transparent'} aria-label="locked">
          <FaLink/>
        </IconButton>
      </Popover.Trigger>
      <Popover.Content>
        <Popover.Arrow/>
        <Popover.CloseTrigger/>
        <Popover.Header>{text || 'Share this link'}</Popover.Header>
        <Popover.Body>
          <Box>
            <HStack gap={2}>
              <Clipboard.Root maxW="300px" value={link}>
                <InputGroup endElement={<ClipboardIconButton/>}>
                  <Clipboard.Input asChild>
                    <Input/>
                  </Clipboard.Input>
                </InputGroup>
              </Clipboard.Root>
            </HStack>
          </Box>
        </Popover.Body>
      </Popover.Content>
    </Popover.Root>
  )
}

const ClipboardIconButton = () => {
  return (
    <Clipboard.Trigger asChild>
      <IconButton variant="surface" size="xs" me="-2">
        <Clipboard.Indicator/>
      </IconButton>
    </Clipboard.Trigger>
  )
}
