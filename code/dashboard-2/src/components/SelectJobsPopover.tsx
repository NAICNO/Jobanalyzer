import {
  IconButton,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverCloseButton,
  PopoverContent,
  PopoverHeader,
  PopoverTrigger,
  Text,
  Link as ChakraLink,
} from '@chakra-ui/react'
import { ExternalLinkIcon, Icon } from '@chakra-ui/icons'
import { FaLock } from 'react-icons/fa'

const SelectJobsPopover = () => {
  return (
    <Popover placement="right">
      <PopoverTrigger>
        <IconButton bg={'transparent'} aria-label="locked" icon={<Icon as={FaLock}/>}/>
      </PopoverTrigger>
      <PopoverContent>
        <PopoverArrow/>
        <PopoverCloseButton/>
        <PopoverHeader>Password protected!</PopoverHeader>
        <PopoverBody>
          <Text>
            <ChakraLink
              color="teal.500"
              href="https://github.com/NAICNO/Jobanalyzer/issues/new?title=Access"
              isExternal
            >
              File an issue with the title "Access"
              <ExternalLinkIcon mx="4px" mb="4px"/>
            </ChakraLink>
            {' '}if you need access.
          </Text>
        </PopoverBody>
      </PopoverContent>
    </Popover>

  )
}

export default SelectJobsPopover
