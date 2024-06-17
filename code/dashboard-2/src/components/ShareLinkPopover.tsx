import {
  Box,
  HStack, IconButton,
  Input,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverCloseButton,
  PopoverContent,
  PopoverHeader,
  PopoverTrigger,
  UsePopperProps,
  useToast,
  UseToastOptions
} from '@chakra-ui/react'
import { CopyIcon, LinkIcon } from '@chakra-ui/icons'

interface ShareLinkPopoverProps {
  link: string
  text: string
  placement: UsePopperProps['placement']
  showToast?: boolean
  toastProps?: {
    description?: UseToastOptions['duration']
    duration?: UseToastOptions['duration']
  }
}


const ShareLinkPopover = ({link, text, placement, showToast, toastProps}: ShareLinkPopoverProps) => {

  const toast = useToast()

  return (
    <Popover placement={placement}>
      <PopoverTrigger>
        <IconButton bg={'transparent'} aria-label="locked" icon={<LinkIcon/>}/>
      </PopoverTrigger>
      <PopoverContent>
        <PopoverArrow/>
        <PopoverCloseButton/>
        <PopoverHeader>{text || 'Share this link'}</PopoverHeader>
        <PopoverBody>
          <Box>
            <HStack spacing={2}>
              <Input value={link} isReadOnly/>
              <CopyIcon
                ml="10px"
                onClick={() =>
                  navigator.clipboard.writeText(link)
                    .then(() => {
                      if (showToast) {
                        toast({
                          description: toastProps?.description || 'Link copied to clipboard',
                          status: 'success',
                          duration: toastProps?.duration || 2000,
                          isClosable: true,
                        })
                      }
                    })}
                aria-label={'copy'}
                opacity="0.4"
                _hover={{opacity: 1}}
              />
            </HStack>
          </Box>
        </PopoverBody>
      </PopoverContent>
    </Popover>
  )
}

export default ShareLinkPopover
