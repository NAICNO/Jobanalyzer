import React, { useEffect, useState } from 'react'
import { Button, HStack, Input, Text } from '@chakra-ui/react'

interface NodeSelectionInputProps {
  defaultQuery: string
  onClickSubmit: (query: string) => void
  onClickHelp: () => void
  focusRef: React.RefObject<HTMLInputElement>
}

const NodeSelectionInput = ({defaultQuery, onClickSubmit, onClickHelp, focusRef} : NodeSelectionInputProps) => {

  const [query, setQuery] = useState<string>(defaultQuery)
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => setQuery(event.target.value)

  useEffect(() => {
    setQuery(defaultQuery)
  }, [defaultQuery])

  return (
    <HStack spacing={2} my="20px">
      <Text whiteSpace="nowrap">Node selection:</Text>
      <Input
        placeholder="Type selection"
        width="80%"
        ref={focusRef}
        value={query}
        onChange={handleChange}
      />
      <Button colorScheme="blue" px="30px" onClick={() => onClickSubmit(query)}>Submit</Button>
      <Button variant="outline" onClick={onClickHelp}>Help</Button>
    </HStack>
  )
}

export default NodeSelectionInput
