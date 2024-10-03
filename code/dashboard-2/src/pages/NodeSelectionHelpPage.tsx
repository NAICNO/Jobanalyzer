import { Heading, VStack } from '@chakra-ui/react'
import { NodeSelectionHelpContent, PageTitle } from '../components'

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
