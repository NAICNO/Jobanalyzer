import { HStack, Box, Text } from '@chakra-ui/react'

interface BenchmarkLegendProps {
  tests: string[]
  colorMap: Record<string, string>
}

export const BenchmarkLegend = ({ tests, colorMap }: BenchmarkLegendProps) => {
  return (
    <HStack gap={4} flexWrap="wrap">
      {tests.map((test) => (
        <HStack key={test} gap={1}>
          <Box w="12px" h="12px" bg={colorMap[test]} borderRadius="2px" flexShrink={0} />
          <Text fontSize="sm">{test}</Text>
        </HStack>
      ))}
    </HStack>
  )
}
