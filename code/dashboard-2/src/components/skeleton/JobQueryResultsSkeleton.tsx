import {
  Card,
  Skeleton,
  Stack,
  VStack
} from '@chakra-ui/react'

const SKELETON_COUNT = 10

export const JobQueryResultsSkeleton = () => {
  return (
    <VStack alignItems={'start'} w="80%">
      <Card.Root mt="10px" w="100%" variant={'outline'}>
        <Card.Body>
          <Stack gap={4}>
            <Skeleton height="30px" width="200px"/>
            <Skeleton height="20px" width="400px"/>
            <Skeleton height="20px"/>
            <Stack gap={2}>
              {
                Array.from({length: SKELETON_COUNT}).map((_, index) => (
                  <Skeleton key={index} height="10px"/>
                ))
              }
            </Stack>
          </Stack>
        </Card.Body>
      </Card.Root>
    </VStack>
  )
}
