import {
  Card,
  CardBody,
  Skeleton,
  Stack,
  VStack
} from '@chakra-ui/react'

const SKELETON_COUNT = 10

export const JobQueryResultsSkeleton = () => {
  return (
    <VStack alignItems={'start'} w="80%">
      <Card mt="10px" w="100%">
        <CardBody>
          <Stack spacing={4}>
            <Skeleton height="30px" width="200px"/>
            <Skeleton height="20px" width="400px"/>
            <Skeleton height="20px"/>
            <Stack spacing={2}>
              {
                Array.from({length: SKELETON_COUNT}).map((_, index) => (
                  <Skeleton key={index} height="10px"/>
                ))
              }
            </Stack>
          </Stack>
        </CardBody>
      </Card>
    </VStack>
  )
}
