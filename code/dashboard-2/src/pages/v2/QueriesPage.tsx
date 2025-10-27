import { Box, Text } from '@chakra-ui/react'
import { TimeRangePicker, TimeRange } from '../../components/TimeRangePicker'
import { useState } from 'react'

export const QueriesPage = () => {
  const [timeRange, setTimeRange] = useState<TimeRange>()

  return (
    <Box>
      <Text fontSize="2xl" fontWeight="bold">
        Queries
      </Text>
      <TimeRangePicker
        value={timeRange}
        onChange={setTimeRange}
        timezone="CET"
      />
    </Box>
  )
}
