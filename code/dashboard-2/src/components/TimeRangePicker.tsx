import { useEffect, useState, useRef } from 'react'
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Input,
  Field,
  Separator,
  Portal,
} from '@chakra-ui/react'
import { Menu } from '@chakra-ui/react'
import { FiClock, FiCalendar, FiGlobe, FiChevronRight, FiInfo, FiPlusCircle, FiExternalLink, FiChevronLeft, FiChevronDown } from 'react-icons/fi'

type SidePaneMode = 'absolute' | 'around' | null

/**
 * TimeRange describes the selected range.
 * - type 'relative': value like '5m', '1h', '7d', '1w'. When endAt='now', it's a moving window.
 * - type 'absolute': start and end are concrete Date objects.
 * - type 'custom': allow freeform labels; downstream decides how to interpret value.
 */
export interface TimeRange {
  type: 'relative' | 'absolute' | 'custom'
  label: string
  value: string
  start?: Date
  end?: Date
  // If endAt is 'now', the range is considered moving and may auto-refresh
  endAt?: 'now' | 'fixed'
  // Auto refresh interval in seconds; 0 or undefined disables auto refresh
  refreshIntervalSec?: number
}

interface TimeRangePickerProps {
  value?: TimeRange
  onChange?: (range: TimeRange) => void
  timezone?: string
}

const relativeTimeRanges: TimeRange[] = [
  { type: 'relative', label: 'Last 5 minutes', value: '5m', endAt: 'now' },
  { type: 'relative', label: 'Last 15 minutes', value: '15m', endAt: 'now' },
  { type: 'relative', label: 'Last 30 minutes', value: '30m', endAt: 'now' },
  { type: 'relative', label: 'Last 1 hour', value: '1h', endAt: 'now' },
  { type: 'relative', label: 'Last 3 hours', value: '3h', endAt: 'now' },
  { type: 'relative', label: 'Last 6 hours', value: '6h', endAt: 'now' },
  { type: 'relative', label: 'Last 1 day', value: '1d', endAt: 'now' },
  { type: 'relative', label: 'Last 7 days', value: '7d', endAt: 'now' },
]

const parseRelativeToMinutes = (value: string): number | null => {
  const m = value.match(/^(\d+)([mhdw])$/)
  if (!m) return null
  const num = Number(m[1])
  const unit = m[2]
  if (unit === 'm') return num
  if (unit === 'h') return num * 60
  if (unit === 'd') return num * 1440
  if (unit === 'w') return num * 10080
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
  const [sidePaneMode, setSidePaneMode] = useState<SidePaneMode>(null)
  const [sidePanePosition, setSidePanePosition] = useState<{ top: number; left?: number; right?: number }>({ top: 100 })
  const [endAt] = useState<'now' | 'fixed'>(value?.endAt ?? 'now')
  const [refreshIntervalSec] = useState<number>(value?.refreshIntervalSec ?? 0)
  const [startDate, setStartDate] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endDate, setEndDate] = useState('')
  const [endTime, setEndTime] = useState('')
  const [aroundDate, setAroundDate] = useState('')
  const [aroundClock, setAroundClock] = useState('')
  const [aroundDurationMin, setAroundDurationMin] = useState(durationOptions[0]?.minutes ?? 15)
  const menuRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)

  const selectedRange = value || relativeTimeRanges[3] // Default to "Last 1 hour"

  // Parse duration string (e.g., '15m', '1h', '7d') to milliseconds
  const parseDurationToMs = (durationStr: string): number | null => {
    const match = durationStr.match(/^(\d+)([mhdw])$/)
    if (!match) return null
    const num = parseInt(match[1], 10)
    const unit = match[2]
    switch (unit) {
      case 'm': return num * 60 * 1000
      case 'h': return num * 60 * 60 * 1000
      case 'd': return num * 24 * 60 * 60 * 1000
      case 'w': return num * 7 * 24 * 60 * 60 * 1000
      default: return null
    }
  }

  const handleNavigate = (direction: 'prev' | 'next') => {
    // For relative ranges with endAt='now', shift the time window
    if (selectedRange.type === 'relative' && selectedRange.endAt === 'now') {
      const durationMs = parseDurationToMs(selectedRange.value)
      if (!durationMs) return

      const now = new Date()
      const shift = direction === 'prev' ? -durationMs : durationMs
      const newEnd = new Date(now.getTime() + shift)
      const newStart = new Date(newEnd.getTime() - durationMs)

      // Create an absolute range for the shifted window
      const shiftedRange: TimeRange = {
        type: 'absolute',
        label: `${formatDateTime(newStart)} - ${formatDateTime(newEnd)}`,
        value: `${newStart.toISOString()}_${newEnd.toISOString()}`,
        start: newStart,
        end: newEnd,
        endAt: 'fixed',
      }
      onChange?.(shiftedRange)
      return
    }

    // For absolute ranges, shift by the duration
    if (selectedRange.type === 'absolute' && selectedRange.start && selectedRange.end) {
      const duration = selectedRange.end.getTime() - selectedRange.start.getTime()
      const shift = direction === 'prev' ? -duration : duration
      const newStart = new Date(selectedRange.start.getTime() + shift)
      const newEnd = new Date(selectedRange.end.getTime() + shift)

      const shiftedRange: TimeRange = {
        type: 'absolute',
        label: `${formatDateTime(newStart)} - ${formatDateTime(newEnd)}`,
        value: `${newStart.toISOString()}_${newEnd.toISOString()}`,
        start: newStart,
        end: newEnd,
        endAt: 'fixed',
      }
      onChange?.(shiftedRange)
    }
  }

  // Format date/time for display in 12-hour format
  const formatDateTime = (date: Date): string => {
    let hours = date.getHours()
    const ampm = hours >= 12 ? 'PM' : 'AM'
    hours = hours % 12
    hours = hours ? hours : 12 // 0 should be 12
    const minutes = String(date.getMinutes()).padStart(2, '0')
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${month}/${day} ${hours}:${minutes} ${ampm}`
  }

  const handleSelectRange = (range: TimeRange) => {
    onChange?.(range)
    setIsOpen(false)
    setSidePaneMode(null)
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
        endAt: 'fixed',
        refreshIntervalSec: 0,
      }
      onChange?.(absoluteRange)
      setIsOpen(false)
      setSidePaneMode(null)
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
      endAt: 'fixed',
      refreshIntervalSec: 0,
    }
    onChange?.(range)
    setIsOpen(false)
    setSidePaneMode(null)
  }

  // Auto refresh effect when using a relative range ending at now
  // Emits onChange periodically with the same selected relative range to trigger upstream refresh
  // Disabled when the menu is open to avoid spurious updates while editing
  useEffect(() => {
    const shouldTick = !isOpen && endAt === 'now' && refreshIntervalSec > 0 && selectedRange.type !== 'absolute'
    if (!shouldTick) return
    const id = setInterval(() => {
      // fire with a fresh object reference to trigger parent updates
      onChange?.({ ...selectedRange, endAt: 'now', refreshIntervalSec })
    }, refreshIntervalSec * 1000)
    return () => clearInterval(id)
  }, [isOpen, endAt, refreshIntervalSec, selectedRange, onChange])

  // Calculate side pane position when it opens
  useEffect(() => {
    if (!sidePaneMode || !menuRef.current) return
    const menuRect = menuRef.current.getBoundingClientRect()
    const sidePaneWidth = 340
    const viewportWidth = window.innerWidth
    const spaceOnRight = viewportWidth - menuRect.right
    const spaceOnLeft = menuRect.left

    // Prefer right if there's enough space, otherwise left
    const position: { top: number; left?: number; right?: number } = {
      top: menuRect.top,
    }
    if (spaceOnRight >= sidePaneWidth) {
      position.left = menuRect.right + 8
    } else if (spaceOnLeft >= sidePaneWidth) {
      position.right = viewportWidth - menuRect.left + 8
    } else {
      // Fallback: position on right regardless
      position.left = menuRect.right + 8
    }
    setSidePanePosition(position)
  }, [sidePaneMode])

  return (
    <HStack gap={0} align="stretch" w="fit-content">
      {/* Previous button */}
      <Button
        variant="outline"
        size="sm"
        px={2}
        borderRightRadius={0}
        borderRightWidth={0}
        onClick={() => handleNavigate('prev')}
        aria-label="Previous time range"
      >
        <Box as={FiChevronLeft} boxSize={4} />
      </Button>

      {/* Main time range selector */}
      <Menu.Root open={isOpen} onOpenChange={(e) => setIsOpen(e.open)}>
        <Menu.Trigger asChild>
          <Button ref={triggerRef} variant="outline" size="sm" borderRadius={0}>
            <HStack gap={2}>
              <HStack gap={2}>
                <Box as={FiClock} boxSize={4} />
                <Text whiteSpace="nowrap">{selectedRange.label}</Text>
                <Text color="gray.500" fontSize="sm">{timezone}</Text>
              </HStack>
              <Box as={FiChevronDown} boxSize={3.5} color="gray.500" />
            </HStack>
          </Button>
        </Menu.Trigger>
        <Menu.Positioner>
          <Menu.Content ref={menuRef} minW="300px" maxH="400px" overflowY="auto">
          <Box p={1.5}>
            <VStack align="stretch" gap={1}>
              {/* Custom Input */}
              <Box>
                <Text fontSize="xs" fontWeight="semibold" color="blue.600" mb={1}>
                Relative time (15m, 1h, 1d, 1w)
                </Text>
                <HStack>
                  <Input
                    size="xs"
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

              <Separator my={0.5} />

              {/* Relative Time Ranges */}
              <VStack align="stretch" gap={0}>
                {relativeTimeRanges.map((range) => (
                  <Menu.Item
                    key={range.value}
                    value={range.value}
                    onClick={() => handleSelectRange(range)}
                    bg={selectedRange.value === range.value ? 'blue.50' : 'white'}
                    _hover={{ bg: 'gray.50' }}
                    cursor="pointer"
                    py={1}
                    px={2}
                    rounded="sm"
                  >
                    <HStack justify="space-between" w="100%">
                      <Text fontSize="xs">{range.label}</Text>
                      <Text fontSize="xs" color="gray.500" fontFamily="mono">
                        {range.value}
                      </Text>
                    </HStack>
                  </Menu.Item>
                ))}
              </VStack>

              <Separator my={0.5} />

              {/* Start and End Times */}
              <Box>
                <Button
                  variant="ghost"
                  size="xs"
                  w="100%"
                  justifyContent="space-between"
                  onClick={() => setSidePaneMode('absolute')}
                  py={1}
                >
                  <HStack gap={1.5}>
                    <Box as={FiCalendar} boxSize={3.5} />
                    <Text fontSize="xs">Start and end times</Text>
                  </HStack>
                  <Box as={FiChevronRight} boxSize={3.5} />
                </Button>
              </Box>

              {/* Around a Time */}
              <Box>
                <Button
                  variant="ghost"
                  size="xs"
                  w="100%"
                  justifyContent="space-between"
                  onClick={() => setSidePaneMode('around')}
                  py={1}
                >
                  <HStack gap={1.5}>
                    <Box as={FiClock} boxSize={3.5} />
                    <Text fontSize="xs">Around a time</Text>
                  </HStack>
                  <Box as={FiChevronRight} boxSize={3.5} />
                </Button>
              </Box>

              {/* Timezone */}
              <Separator my={0.5} />
              <Box>
                <Button
                  variant="ghost"
                  size="xs"
                  w="100%"
                  justifyContent="space-between"
                  py={1}
                >
                  <HStack gap={1.5}>
                    <Box as={FiGlobe} boxSize={3.5} />
                    <Text fontSize="xs">Time zone: {timezone} (UTC+1)</Text>
                  </HStack>
                  <Box as={FiChevronRight} boxSize={3.5} />
                </Button>
              </Box>
            </VStack>
          </Box>
        </Menu.Content>
      </Menu.Positioner>

      {/* Side Pane for Absolute and Around forms */}
      {sidePaneMode && (
        <Portal>
          {/* Backdrop overlay */}
          <Box
            position="fixed"
            top={0}
            left={0}
            right={0}
            bottom={0}
            bg="blackAlpha.300"
            zIndex={1499}
            onClick={() => setSidePaneMode(null)}
          />
          {/* Side pane content */}
          <Box
            position="fixed"
            top={`${sidePanePosition.top}px`}
            left={sidePanePosition.left !== undefined ? `${sidePanePosition.left}px` : undefined}
            right={sidePanePosition.right !== undefined ? `${sidePanePosition.right}px` : undefined}
            bg="white"
            boxShadow="0 4px 12px rgba(0,0,0,0.15)"
            rounded="md"
            w="340px"
            maxH="70vh"
            overflowY="auto"
            zIndex={1500}
            p={2.5}
          >
            <VStack align="stretch" gap={1.5}>
              <Text fontSize="sm" fontWeight="semibold">
                {sidePaneMode === 'absolute' ? 'Start and end times' : 'Around a time'}
              </Text>

              {sidePaneMode === 'absolute' ? (
                <>
                  {/* Start time */}
                  <VStack align="stretch" gap={1}>
                    <Text fontWeight="semibold" fontSize="sm">Start time</Text>
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
                    </Field.Root>
                    <HStack gap={2}>
                      <Input
                        type="time"
                        step={1}
                        size="sm"
                        value={startTime}
                        onChange={(e) => setStartTime(e.target.value)}
                      />
                      <Box borderWidth="1px" rounded="sm" px={1.5} py={0.5} fontSize="xs" color="gray.600">
                        {timezone}
                      </Box>
                    </HStack>
                  </VStack>

                  {/* End time */}
                  <VStack align="stretch" gap={1}>
                    <Text fontWeight="semibold" fontSize="sm">End time</Text>
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
                    </Field.Root>
                    <HStack gap={2}>
                      <Input
                        type="time"
                        step={1}
                        size="sm"
                        value={endTime}
                        onChange={(e) => setEndTime(e.target.value)}
                      />
                      <Box borderWidth="1px" rounded="sm" px={1.5} py={0.5} fontSize="xs" color="gray.600">
                        {timezone}
                      </Box>
                    </HStack>
                  </VStack>

                  {/* Actions */}
                  <HStack justify="end" gap={2} pt={2}>
                    <Button variant="ghost" size="sm" onClick={() => setSidePaneMode(null)}>Cancel</Button>
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
                  <HStack align="start" gap={1.5} p={2} bg="gray.50" rounded="sm">
                    <Box as={FiInfo} boxSize={3.5} mt={0.5} color="gray.600" />
                    <Text fontSize="xs" color="gray.700">
                      Tip: Paste ISO 8601 timestamps into date/time fields (e.g., 2025-10-26T23:51:49.902Z)
                    </Text>
                  </HStack>

                  {/* Change format */}
                  <Button variant="plain" colorPalette="blue" justifyContent="flex-start" px={0} size="sm">
                    <HStack gap={1.5}>
                      <Box as={FiPlusCircle} boxSize={3.5} />
                      <Text fontSize="xs">Change date & time format</Text>
                      <Box as={FiExternalLink} boxSize={3} />
                    </HStack>
                  </Button>
                </>
              ) : (
                <>
                  {/* Around a time form */}
                  <Text fontWeight="semibold" fontSize="sm">Around a time</Text>
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
                  </Field.Root>
                  <HStack gap={2}>
                    <Input
                      type="time"
                      step={1}
                      size="sm"
                      value={aroundClock}
                      onChange={(e) => setAroundClock(e.target.value)}
                    />
                    <Box borderWidth="1px" rounded="sm" px={1.5} py={0.5} fontSize="xs" color="gray.600">
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

                  {/* Actions */}
                  <HStack justify="end" gap={2} pt={2}>
                    <Button variant="ghost" size="sm" onClick={() => setSidePaneMode(null)}>Cancel</Button>
                    <Button
                      size="sm"
                      colorPalette="blue"
                      onClick={handleAroundTime}
                      disabled={!aroundDate || !aroundClock}
                    >
                      Apply
                    </Button>
                  </HStack>

                  {/* Tip */}
                  <HStack align="start" gap={1.5} p={2} bg="gray.50" rounded="sm">
                    <Box as={FiInfo} boxSize={3.5} mt={0.5} color="gray.600" />
                    <Text fontSize="xs" color="gray.700">
                      Tip: Paste ISO 8601 timestamps into date/time fields (e.g., 2025-10-26T23:52:32.386Z)
                    </Text>
                  </HStack>

                  <Button variant="plain" colorPalette="blue" justifyContent="flex-start" px={0} size="sm">
                    <HStack gap={1.5}>
                      <Box as={FiPlusCircle} boxSize={3.5} />
                      <Text fontSize="xs">Change date & time format</Text>
                      <Box as={FiExternalLink} boxSize={3} />
                    </HStack>
                  </Button>
                </>
              )}
            </VStack>
          </Box>
        </Portal>
      )}
    </Menu.Root>

    {/* Next button */}
    <Button
      variant="outline"
      size="sm"
      px={2}
      borderLeftRadius={0}
      borderLeftWidth={0}
      onClick={() => handleNavigate('next')}
      aria-label="Next time range"
    >
      <Box as={FiChevronRight} boxSize={4} />
    </Button>
  </HStack>
  )
}
