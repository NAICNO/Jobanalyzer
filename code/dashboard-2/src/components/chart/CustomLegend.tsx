import { Box, Text } from '@chakra-ui/react'
import { DataKey } from 'recharts/types/util/types'

interface CustomLegendProps {
  payload?: any[];
  maxLegendItems?: number;
  hiddenSeriesCount?: number;
  onLegendClick: (dataKey?: DataKey<string>) => void;
}

export const CustomLegend = ({
  payload,
  maxLegendItems,
  hiddenSeriesCount = 0,
  onLegendClick
}: CustomLegendProps) => {
  if (!payload) return null

  // If maxLegendItems is undefined, null, or not positive, show all items
  const shouldLimitItems = maxLegendItems !== undefined && maxLegendItems !== null && maxLegendItems > 0
  
  // Display all items or limit based on maxLegendItems
  const displayItems = shouldLimitItems ? payload.slice(0, maxLegendItems) : payload
  
  // Only show hidden count message if we're actually limiting items
  const showHiddenCount = shouldLimitItems && hiddenSeriesCount > 0

  return (
    <Box
      fontSize="xs"
      style={{
        paddingBottom: 16,
        overflowY: 'auto',
        textAlign: 'center'
      }}
    >
      {displayItems.map((entry: any, index: number) => (
        <Box
          key={`legend-item-${index}`}
          style={{
            display: 'inline-block',
            marginRight: 16,
            marginBottom: 2,
            cursor: 'pointer'
          }}
          onClick={() => onLegendClick(entry.dataKey)}
        >
          <Text
            style={{
              display: 'inline-block',
              width: '8px',
              height: '8px',
              backgroundColor: entry.color,
              marginRight: '3px'
            }}
          />
          {entry.value}
        </Box>
      ))}
      {showHiddenCount && (
        <Box as="div" fontStyle="italic" opacity={0.7} fontSize="10px">
          +{hiddenSeriesCount} more
        </Box>
      )}
    </Box>
  )
}
