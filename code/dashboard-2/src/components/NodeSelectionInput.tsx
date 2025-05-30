import React, { useEffect, useState } from 'react'
import { Button, HStack, Input, Text } from '@chakra-ui/react'

interface NodeSelectionInputProps {
  defaultQuery: string
  onClickSubmit: (query: string) => void
  onClickHelp: () => void
  focusRef: React.RefObject<HTMLInputElement>
}

export const NodeSelectionInput = ({defaultQuery, onClickSubmit, onClickHelp, focusRef}: NodeSelectionInputProps) => {

  const [query, setQuery] = useState<string>(defaultQuery)
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => setQuery(event.target.value)

  useEffect(() => {
    setQuery(defaultQuery)
  }, [defaultQuery])

  const submitQuery = () => {
    onClickSubmit(query)
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      submitQuery()
    }
  }

  return (
    <HStack gap={2} mt={2} mb={6}>
      <Text whiteSpace="nowrap">Node selection:</Text>
      <Input
        placeholder="Type selection"
        ref={focusRef}
        value={query}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        variant="subtle"
      />
      <Button onClick={submitQuery} colorPalette={'blue'}>Submit</Button>
      <Button onClick={onClickHelp} variant={'outline'}>Help</Button>
    </HStack>
  )
}
