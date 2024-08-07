import { Heading, VStack } from '@chakra-ui/react'
import NodeSelectionHelpContent from '../components/NodeSelectionHelpContent.tsx'
import PageTitle from '../components/PageTitle.tsx'

export default function NodeSelectionHelpPage() {
  return (
    <>
      <PageTitle title={'Query Help'}/>
      <VStack maxW={'60%'} alignItems={'start'} mx={4}>
        <Heading my={'20px'}>Query help</Heading>
        <NodeSelectionHelpContent/>
      </VStack>
    </>
  )
}
