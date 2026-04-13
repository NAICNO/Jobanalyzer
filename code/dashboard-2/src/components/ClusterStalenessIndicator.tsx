import { useEffect, useState } from 'react'
import { HStack, Text, IconButton, Circle } from '@chakra-ui/react'
import { LuRefreshCw } from 'react-icons/lu'

interface Props {
  oldestDataUpdatedAt: number
  isFetching: boolean
  onRefresh: () => void
}

export const ClusterStalenessIndicator = ({ oldestDataUpdatedAt, isFetching, onRefresh }: Props) => {
  const [now, setNow] = useState(Date.now())

  // Tick every 15 seconds to update the "ago" label
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 15_000)
    return () => clearInterval(id)
  }, [])

  if (oldestDataUpdatedAt === 0) {
    return null
  }

  const ageMs = now - oldestDataUpdatedAt
  const ageSec = Math.floor(ageMs / 1000)

  let ageLabel: string
  if (ageSec < 60) {
    ageLabel = 'just now'
  } else if (ageSec < 3600) {
    const mins = Math.floor(ageSec / 60)
    ageLabel = `${mins}m ago`
  } else {
    const hours = Math.floor(ageSec / 3600)
    ageLabel = `${hours}h ago`
  }

  // green < 1min, yellow 1-5min, red > 5min
  const color = ageSec < 60 ? 'green.500' : ageSec < 300 ? 'yellow.500' : 'red.500'

  return (
    <>
      <style>{`
        @keyframes staleness-spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
      <HStack gap={2} align="center">
        <Circle size="8px" bg={color} />
        <Text fontSize="xs" color="gray.500">
          {isFetching ? 'Refreshing...' : `Updated ${ageLabel}`}
        </Text>
        <IconButton
          aria-label="Refresh data"
          size="xs"
          variant="ghost"
          onClick={onRefresh}
          disabled={isFetching}
        >
          <LuRefreshCw
            style={isFetching ? { animation: 'staleness-spin 1s linear infinite' } : undefined}
          />
        </IconButton>
      </HStack>
    </>
  )
}
