import { Center, Container, Heading, VStack } from '@chakra-ui/react'

export default function FoxHomePage() {

  return (
    <Container maxW="xl" height="100vh" centerContent>
      <Center height="100%">
        <VStack spacing={6} width="100%" maxW="md" padding="4">
          <Heading>Fox Page</Heading>
        </VStack>
      </Center>
    </Container>
  )
}
