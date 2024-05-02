import { Heading, VStack } from '@chakra-ui/react'
import NodeSelectionHelpContent from '../components/NodeSelectionHelpContent.tsx'

export default function NodeSelectionHelpPage() {
  return (
    <VStack maxW='60%' alignItems={'start'} mx={4}>
      <Heading my='20px'>Query help</Heading>
      <NodeSelectionHelpContent/>
    </VStack>
  )
}
