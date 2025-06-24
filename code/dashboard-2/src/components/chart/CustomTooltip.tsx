import { Box, Text } from '@chakra-ui/react'

import { useColorModeValue } from '../ui/color-mode.tsx'
import { dateTimeFormatter } from '../../util'

interface CustomTooltipProps {
  active?: boolean;
  payload?: any[];
  label?: any;
  maxItems?: number;
}

export const CustomTooltip = ({ active, payload, label, maxItems }: CustomTooltipProps) => {
  if (!active || !payload || payload.length === 0) {
    return null
  }

  // Sort payload items by value (descending) to show the most significant ones
  const sortedPayload = [...payload].sort((a, b) => b.value - a.value)

  // If maxItems is undefined or null, show all items
  const shouldLimitItems = maxItems !== undefined && maxItems !== null && maxItems > 0

  // Limit the number of items displayed only if maxItems is specified
  const displayItems = shouldLimitItems ? sortedPayload.slice(0, maxItems) : sortedPayload
  const hiddenCount = shouldLimitItems ? payload.length - maxItems : 0

  // Theme-aware colors
  const bg = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.600')
  const textColor = useColorModeValue('gray.800', 'gray.100')

  return (
    <Box
      bg={bg}
      p={2}
      border="1px solid"
      borderColor={borderColor}
      borderRadius="md"
      boxShadow="md"
      color={textColor}
      fontSize="sm"
    >
      <Text  >{dateTimeFormatter(label)}</Text>
      {displayItems.map((item, index) => (
        <Text key={index} color={item.color}>
          {item.name}: {item.value}
        </Text>
      ))}
      {hiddenCount > 0 && (
        <Text fontStyle="italic" mt={1} opacity={0.8}>
          +{hiddenCount} more series not shown
        </Text>
      )}
    </Box>
  )
}
