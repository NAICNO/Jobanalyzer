import {
  Box,
  Button, HStack,
  Link as ChakraLink,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverCloseButton,
  PopoverContent,
  PopoverHeader,
  PopoverTrigger,
} from '@chakra-ui/react'
import { Link as ReactRouterLink } from 'react-router-dom'

import { CellWrapper } from './CellWrapper.tsx'
import { APP_URL, PROFILE_NAMES } from '../../../Constants.ts'
import { JobQueryJobId } from '../../../types'

interface JobQueryJobIdCellProps {
  value: JobQueryJobId
}

export const JobQueryJobIdCell = ({value}: JobQueryJobIdCellProps) => {
  return (
    <CellWrapper styleProps={{paddingLeft: 2}}>
      <Popover
        placement="right"
        closeOnBlur={false}
        isLazy
        trigger={'hover'}
        openDelay={250}
      >
        <PopoverTrigger>
          <ChakraLink>
            {value.jobId}
          </ChakraLink>
        </PopoverTrigger>
        <PopoverContent boxShadow="2xl">
          <PopoverArrow/>
          <PopoverCloseButton/>
          <PopoverHeader>Show profile for</PopoverHeader>
          <PopoverBody>
            <Box>
              <HStack alignItems="stretch" spacing={2}>
                {
                  PROFILE_NAMES.map((profile) => (
                    <LinkButton
                      key={profile.key}
                      cluster={value.clusterName}
                      jobId={value.jobId}
                      host={value.hostName}
                      from={value.from}
                      to={value.to}
                      profile={profile}
                    />
                  ))
                }
              </HStack>
            </Box>
          </PopoverBody>
        </PopoverContent>
      </Popover>
    </CellWrapper>
  )
}

interface LinkButtonProps {
  cluster: string
  jobId: string
  host: string
  from?: string
  to?: string
  profile: {
    key: string
    text: string
  }
}

const LinkButton = ({cluster, jobId, host, from, to, profile}: LinkButtonProps) => {
  const buttonWidth = '68px'
  const fmt = `html,${profile.key}`
  let link = `${APP_URL}/profile?cluster=${cluster}&job=${jobId}&host=${host}&fmt=${fmt}`
  if (from) {
    link += `&from=${from}`
  }
  if (to) {
    link += `&to=${to}`
  }
  const encodedLink = encodeURI(link)
  return (
    <ChakraLink as={ReactRouterLink} to={encodedLink} width="100%" isExternal>
      <Button size="xs" colorScheme="teal" width={buttonWidth}>{profile.text}</Button>
    </ChakraLink>
  )
}
