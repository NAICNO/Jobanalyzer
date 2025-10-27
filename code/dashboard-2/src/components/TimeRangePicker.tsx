import { useState } from 'react'
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Input,
  Field,
  Separator,
} from '@chakra-ui/react'
import { Menu } from '@chakra-ui/react'
import { FiClock, FiCalendar, FiGlobe, FiChevronDown, FiChevronRight, FiInfo, FiPlusCircle, FiExternalLink } from 'react-icons/fi'

export interface TimeRange {
  type: 'relative' | 'absolute' | 'custom'
  label: string
  value: string
  start?: Date
  end?: Date
}

interface TimeRangePickerProps {
  value?: TimeRange
  onChange?: (range: TimeRange) => void
  timezone?: string
}

const relativeTimeRanges: TimeRange[] = [
  { type: 'relative', label: 'Today', value: 'today' },
  { type: 'relative', label: 'Yesterday', value: 'yesterday' },
  { type: 'relative', label: 'Last 15 minutes', value: '15m' },
  { type: 'relative', label: 'Last 30 minutes', value: '30m' },
  { type: 'relative', label: 'Last 1 hour', value: '1h' },
  { type: 'relative', label: 'Last 3 hours', value: '3h' },
  { type: 'relative', label: 'Last 6 hours', value: '6h' },
  { type: 'relative', label: 'Last 12 hours', value: '12h' },
  { type: 'relative', label: 'Last 1 day', value: '1d' },
  { type: 'relative', label: 'Last 2 days', value: '2d' },
  { type: 'relative', label: 'Last 7 days', value: '7d' },
  { type: 'relative', label: 'Last 14 days', value: '14d' },
]

const parseRelativeToMinutes = (value: string): number | null => {
  const m = value.match(/^(\d+)([mhd])$/)
  if (!m) return null
  const num = Number(m[1])
  const unit = m[2]
  if (unit === 'm') return num
  if (unit === 'h') return num * 60
  if (unit === 'd') return num * 1440
  return null
}

const durationOptions = relativeTimeRanges
  .map((r) => {
    const minutes = parseRelativeToMinutes(r.value)
    if (minutes == null) return null
    const pretty = r.label.replace(/^Last\s+/, '')
    return { minutes, label: `± ${pretty}`, value: r.value }
  })
  .filter((x): x is { minutes: number; label: string; value: string } => !!x)


export const TimeRangePicker = ({
  value,
  onChange,
  timezone = 'CET',
}: TimeRangePickerProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const [customInput, setCustomInput] = useState('')
  const [showAbsolute, setShowAbsolute] = useState(false)
  const [showAroundTime, setShowAroundTime] = useState(false)
  const [startDate, setStartDate] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endDate, setEndDate] = useState('')
  const [endTime, setEndTime] = useState('')
  const [aroundDate, setAroundDate] = useState('')
  const [aroundClock, setAroundClock] = useState('')
  const [aroundDurationMin, setAroundDurationMin] = useState(durationOptions[0]?.minutes ?? 15)

  const selectedRange = value || relativeTimeRanges[4] // Default to "Last 1 hour"

  const handleSelectRange = (range: TimeRange) => {
    onChange?.(range)
    setIsOpen(false)
    setShowAbsolute(false)
    setShowAroundTime(false)
  }

  const handleCustomInput = () => {
    if (customInput.trim()) {
      const customRange: TimeRange = {
        type: 'custom',
        label: customInput,
        value: customInput,
      }
      onChange?.(customRange)
      setIsOpen(false)
    }
  }

  const combineDateTime = (dateStr: string, timeStr: string) => {
    if (!dateStr || !timeStr) return undefined
    // Expect dateStr: YYYY-MM-DD, timeStr: HH:MM[:SS]
    const iso = `${dateStr}T${timeStr}`
    const d = new Date(iso)
    return isNaN(d.getTime()) ? undefined : d
  }

  const handleAbsoluteRange = () => {
    const start = combineDateTime(startDate, startTime)
    const end = combineDateTime(endDate, endTime)
    if (start && end) {
      const absoluteRange: TimeRange = {
        type: 'absolute',
        label: `${start.toLocaleDateString()} - ${end.toLocaleDateString()}`,
        value: 'custom',
        start,
        end,
      }
      onChange?.(absoluteRange)
      setIsOpen(false)
      setShowAbsolute(false)
    }
  }

  const handleAroundTime = () => {
    const center = combineDateTime(aroundDate, aroundClock)
    if (!center) return
    const deltaMs = aroundDurationMin * 60 * 1000
    const start = new Date(center.getTime() - deltaMs)
    const end = new Date(center.getTime() + deltaMs)
    const range: TimeRange = {
      type: 'absolute',
      label: `Around ${center.toLocaleString()} (± ${aroundDurationMin}m)`,
      value: 'around',
      start,
      end,
    }
    onChange?.(range)
    setIsOpen(false)
    setShowAroundTime(false)
  }

  return (
    <Menu.Root open={isOpen} onOpenChange={(e) => setIsOpen(e.open)}>
      <Menu.Trigger asChild>
        <Button variant="outline" size="sm">
          <HStack gap={2}>
            <Box as={FiClock} boxSize={4} />
            <Text>{selectedRange.label}</Text>
            <Text color="gray.500" fontSize="sm">{timezone}</Text>
          </HStack>
        </Button>
      </Menu.Trigger>
      <Menu.Positioner>
        <Menu.Content minW="400px" maxH="600px" overflowY="auto">
          <Box p={3}>
            <VStack align="stretch" gap={2}>
              {/* Custom Input */}
              <Box>
                <Text fontSize="sm" fontWeight="semibold" color="blue.600" mb={2}>
                Relative time (15m, 1h, 1d, 1w)
                </Text>
                <HStack>
                  <Input
                    size="sm"
                    placeholder="e.g., 2h, 3d, 1w"
                    value={customInput}
                    onChange={(e) => setCustomInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        handleCustomInput()
                      }
                    }}
                  />
                </HStack>
              </Box>

              <Separator my={2} />

              {/* Relative Time Ranges */}
              <VStack align="stretch" gap={1}>
                {relativeTimeRanges.map((range) => (
                  <Menu.Item
                    key={range.value}
                    value={range.value}
                    onClick={() => handleSelectRange(range)}
                    bg={selectedRange.value === range.value ? 'blue.50' : 'white'}
                    _hover={{ bg: 'gray.50' }}
                    cursor="pointer"
                    py={2}
                    px={3}
                    rounded="md"
                  >
                    <HStack justify="space-between" w="100%">
                      <Text fontSize="sm">{range.label}</Text>
                      <Text fontSize="sm" color="gray.500" fontFamily="mono">
                        {range.value}
                      </Text>
                    </HStack>
                  </Menu.Item>
                ))}
              </VStack>

              <Separator my={2} />

              {/* Start and End Times */}
              <Box>
                <Button
                  variant="ghost"
                  size="sm"
                  w="100%"
                  justifyContent="space-between"
                  onClick={() => {
                    setShowAbsolute(!showAbsolute)
                    setShowAroundTime(false)
                  }}
                >
                  <HStack gap={2}>
                    <Box as={FiCalendar} boxSize={4} />
                    <Text>Start and end times</Text>
                  </HStack>
                  <Box as={showAbsolute ? FiChevronDown : FiChevronRight} boxSize={4} />
                </Button>

                {showAbsolute && (
                  <VStack align="stretch" gap={4} mt={2} pl={4}>
                    {/* Start time */}
                    <VStack align="stretch" gap={2}>
                      <Text fontWeight="semibold">Start time</Text>
                      <Field.Root>
                        <Field.Label fontSize="xs">Date</Field.Label>
                        <Input
                          type="date"
                          size="sm"
                          value={startDate}
                          onChange={(e) => setStartDate(e.target.value)}
                          onPaste={(e) => {
                            const text = e.clipboardData.getData('text')
                            const d = new Date(text)
                            if (!isNaN(d.getTime())) {
                              e.preventDefault()
                              const yyyy = d.getFullYear()
                              const mm = String(d.getMonth() + 1).padStart(2, '0')
                              const dd = String(d.getDate()).padStart(2, '0')
                              const hh = String(d.getHours()).padStart(2, '0')
                              const mi = String(d.getMinutes()).padStart(2, '0')
                              const ss = String(d.getSeconds()).padStart(2, '0')
                              setStartDate(`${yyyy}-${mm}-${dd}`)
                              setStartTime(`${hh}:${mi}:${ss}`)
                            }
                          }}
                        />
                        <Field.HelperText fontSize="xs">Date format: m/d/yyyy</Field.HelperText>
                      </Field.Root>
                      <HStack>
                        <Input
                          type="time"
                          step={1}
                          size="sm"
                          value={startTime}
                          onChange={(e) => setStartTime(e.target.value)}
                        />
                        <Box borderWidth="1px" rounded="md" px={2} py={1} fontSize="sm" color="gray.600">
                          {timezone}
                        </Box>
                      </HStack>
                    </VStack>

                    {/* End time */}
                    <VStack align="stretch" gap={2}>
                      <Text fontWeight="semibold">End time</Text>
                      <Field.Root>
                        <Field.Label fontSize="xs">Date</Field.Label>
                        <Input
                          type="date"
                          size="sm"
                          value={endDate}
                          onChange={(e) => setEndDate(e.target.value)}
                          onPaste={(e) => {
                            const text = e.clipboardData.getData('text')
                            const d = new Date(text)
                            if (!isNaN(d.getTime())) {
                              e.preventDefault()
                              const yyyy = d.getFullYear()
                              const mm = String(d.getMonth() + 1).padStart(2, '0')
                              const dd = String(d.getDate()).padStart(2, '0')
                              const hh = String(d.getHours()).padStart(2, '0')
                              const mi = String(d.getMinutes()).padStart(2, '0')
                              const ss = String(d.getSeconds()).padStart(2, '0')
                              setEndDate(`${yyyy}-${mm}-${dd}`)
                              setEndTime(`${hh}:${mi}:${ss}`)
                            }
                          }}
                        />
                        <Field.HelperText fontSize="xs">Date format: m/d/yyyy</Field.HelperText>
                      </Field.Root>
                      <HStack>
                        <Input
                          type="time"
                          step={1}
                          size="sm"
                          value={endTime}
                          onChange={(e) => setEndTime(e.target.value)}
                        />
                        <Box borderWidth="1px" rounded="md" px={2} py={1} fontSize="sm" color="gray.600">
                          {timezone}
                        </Box>
                      </HStack>
                    </VStack>

                    {/* Actions */}
                    <HStack justify="end" gap={3}>
                      <Button variant="ghost" onClick={() => setShowAbsolute(false)}>Cancel</Button>
                      <Button
                        size="sm"
                        colorPalette="blue"
                        onClick={handleAbsoluteRange}
                        disabled={!startDate || !startTime || !endDate || !endTime}
                      >
                        Apply
                      </Button>
                    </HStack>

                    {/* Tip */}
                    <HStack align="start" gap={2} p={3} bg="gray.50" rounded="md">
                      <Box as={FiInfo} boxSize={4} mt={1} color="gray.600" />
                      <Text fontSize="sm" color="gray.700">
                        Tip: Paste timestamps (ISO 8601 format) into the date or time fields. Example: 2025-10-26T23:51:49.902Z
                      </Text>
                    </HStack>

                    {/* Change format */}
                    <Button variant="plain" colorPalette="blue" justifyContent="flex-start" px={0}>
                      <HStack gap={2}>
                        <Box as={FiPlusCircle} boxSize={4} />
                        <Text>Change date & time format</Text>
                        <Box as={FiExternalLink} boxSize={3.5} />
                      </HStack>
                    </Button>
                  </VStack>
                )}
              </Box>

              {/* Around a Time */}
              <Box>
                <Button
                  variant="ghost"
                  size="sm"
                  w="100%"
                  justifyContent="space-between"
                  onClick={() => {
                    setShowAroundTime(!showAroundTime)
                    setShowAbsolute(false)
                  }}
                >
                  <HStack gap={2}>
                    <Box as={FiClock} boxSize={4} />
                    <Text>Around a time</Text>
                  </HStack>
                  <Box as={showAroundTime ? FiChevronDown : FiChevronRight} boxSize={4} />
                </Button>

                {showAroundTime && (
                  <VStack align="stretch" gap={4} mt={2} pl={4}>
                    <Text fontWeight="semibold">Around a time</Text>
                    <Field.Root>
                      <Field.Label fontSize="xs">Date</Field.Label>
                      <Input
                        type="date"
                        size="sm"
                        value={aroundDate}
                        onChange={(e) => setAroundDate(e.target.value)}
                        onPaste={(e) => {
                          const text = e.clipboardData.getData('text')
                          const d = new Date(text)
                          if (!isNaN(d.getTime())) {
                            e.preventDefault()
                            const yyyy = d.getFullYear()
                            const mm = String(d.getMonth() + 1).padStart(2, '0')
                            const dd = String(d.getDate()).padStart(2, '0')
                            const hh = String(d.getHours()).padStart(2, '0')
                            const mi = String(d.getMinutes()).padStart(2, '0')
                            const ss = String(d.getSeconds()).padStart(2, '0')
                            setAroundDate(`${yyyy}-${mm}-${dd}`)
                            setAroundClock(`${hh}:${mi}:${ss}`)
                          }
                        }}
                      />
                      <Field.HelperText fontSize="xs">Date format: m/d/yyyy</Field.HelperText>
                    </Field.Root>
                    <HStack>
                      <Input
                        type="time"
                        step={1}
                        size="sm"
                        value={aroundClock}
                        onChange={(e) => setAroundClock(e.target.value)}
                      />
                      <Box borderWidth="1px" rounded="md" px={2} py={1} fontSize="sm" color="gray.600">
                        {timezone}
                      </Box>
                    </HStack>

                    <Field.Root>
                      <Field.Label fontSize="xs">Duration</Field.Label>
                      <select
                        value={aroundDurationMin}
                        onChange={(e) => setAroundDurationMin(Number(e.target.value))}
                        style={{ padding: '6px', borderRadius: '6px', border: '1px solid var(--chakra-colors-gray-200)' }}
                      >
                        {durationOptions.map((opt) => (
                          <option key={opt.value} value={opt.minutes}>{opt.label}</option>
                        ))}
                      </select>
                    </Field.Root>

                    <HStack justify="end" gap={3}>
                      <Button variant="ghost" onClick={() => setShowAroundTime(false)}>Cancel</Button>
                      <Button
                        size="sm"
                        colorPalette="blue"
                        onClick={handleAroundTime}
                        disabled={!aroundDate || !aroundClock}
                      >
                        Apply
                      </Button>
                    </HStack>

                    <HStack align="start" gap={2} p={3} bg="gray.50" rounded="md">
                      <Box as={FiInfo} boxSize={4} mt={1} color="gray.600" />
                      <Text fontSize="sm" color="gray.700">
                        Tip: Paste timestamps (ISO 8601 format) into the date or time fields. Example: 2025-10-26T23:52:32.386Z
                      </Text>
                    </HStack>

                    <Button variant="plain" colorPalette="blue" justifyContent="flex-start" px={0}>
                      <HStack gap={2}>
                        <Box as={FiPlusCircle} boxSize={4} />
                        <Text>Change date & time format</Text>
                        <Box as={FiExternalLink} boxSize={3.5} />
                      </HStack>
                    </Button>
                  </VStack>
                )}
              </Box>

              {/* Timezone */}
              <Separator my={2} />
              <Box>
                <Button
                  variant="ghost"
                  size="sm"
                  w="100%"
                  justifyContent="space-between"
                >
                  <HStack gap={2}>
                    <Box as={FiGlobe} boxSize={4} />
                    <Text>Time zone: {timezone} (UTC+1)</Text>
                  </HStack>
                  <Box as={FiChevronRight} boxSize={4} />
                </Button>
              </Box>
            </VStack>
          </Box>
        </Menu.Content>
      </Menu.Positioner>
    </Menu.Root>
  )
}
