import { useState, useRef, useCallback, useMemo } from 'react'
import {
  Box,
  HStack,
  VStack,
  Text,
  Input,
  Button,
  Icon,
  IconButton,
  Badge,
  Popover,
  Portal,
  Separator,
} from '@chakra-ui/react'
import { LuBookmark, LuTrash2, LuSave, LuRefreshCw, LuBookmarkPlus } from 'react-icons/lu'
import type { JobQueryPreset, JobQueryPresetFilters, UserSettings } from '../types/userSettings'
import { MAX_PRESETS } from '../types/userSettings'
import { toaster } from './ui/toaster'

interface JobQueryFormValues {
  user: string
  userId: string
  jobId: string
  states: string
  startAfter: string
  startBefore: string
  endAfter: string
  endBefore: string
  submitAfter: string
  submitBefore: string
  minDuration: string
  maxDuration: string
}

interface QueryPresetsProps {
  formValues: JobQueryFormValues
  onLoadPreset: (filters: JobQueryPresetFilters) => void
  settings: UserSettings
  updateSettings: (updater: (current: UserSettings) => UserSettings) => void
  isSaving: boolean
  activePresetId: string | null
  isDirty: boolean
}

const formatRelativeTime = (dateStr: string): string => {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diffMs = now - then
  const diffMins = Math.floor(diffMs / 60_000)
  if (diffMins < 1) return 'now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  const diffWeeks = Math.floor(diffDays / 7)
  if (diffWeeks < 4) return `${diffWeeks}w ago`
  const diffMonths = Math.floor(diffDays / 30)
  return `${diffMonths}mo ago`
}

const buildFilterSummary = (filters: JobQueryPresetFilters): string => {
  const parts: string[] = []
  if (filters.user) parts.push(`User: ${filters.user}`)
  if (filters.userId) parts.push(`UID: ${filters.userId}`)
  if (filters.jobId) parts.push(`Job: ${filters.jobId}`)
  if (filters.states) parts.push(`States: ${filters.states}`)
  if (filters.minDuration) parts.push(`Min: ${filters.minDuration}s`)
  if (filters.maxDuration) parts.push(`Max: ${filters.maxDuration}s`)
  if (filters.startAfter) parts.push(`Start >= ${new Date(filters.startAfter).toLocaleDateString()}`)
  if (filters.startBefore) parts.push(`Start <= ${new Date(filters.startBefore).toLocaleDateString()}`)
  if (filters.endAfter) parts.push(`End >= ${new Date(filters.endAfter).toLocaleDateString()}`)
  if (filters.endBefore) parts.push(`End <= ${new Date(filters.endBefore).toLocaleDateString()}`)
  if (filters.submitAfter) parts.push(`Submit >= ${new Date(filters.submitAfter).toLocaleDateString()}`)
  if (filters.submitBefore) parts.push(`Submit <= ${new Date(filters.submitBefore).toLocaleDateString()}`)
  return parts.length > 0 ? parts.join(' \u2022 ') : 'No filters'
}

const formValuesToFilters = (values: JobQueryFormValues): JobQueryPresetFilters => {
  const filters: JobQueryPresetFilters = {}
  if (values.user) filters.user = values.user
  if (values.userId) filters.userId = values.userId
  if (values.jobId) filters.jobId = values.jobId
  if (values.states) filters.states = values.states
  if (values.startAfter) filters.startAfter = values.startAfter
  if (values.startBefore) filters.startBefore = values.startBefore
  if (values.endAfter) filters.endAfter = values.endAfter
  if (values.endBefore) filters.endBefore = values.endBefore
  if (values.submitAfter) filters.submitAfter = values.submitAfter
  if (values.submitBefore) filters.submitBefore = values.submitBefore
  if (values.minDuration) filters.minDuration = values.minDuration
  if (values.maxDuration) filters.maxDuration = values.maxDuration
  return filters
}

export const QueryPresets = ({
  formValues,
  onLoadPreset,
  settings,
  updateSettings,
  isSaving,
  activePresetId,
  isDirty,
}: QueryPresetsProps) => {
  const [menuOpen, setMenuOpen] = useState(false)
  const [saveMode, setSaveMode] = useState(false)
  const [saveName, setSaveName] = useState('')
  const saveInputRef = useRef<HTMLInputElement>(null)

  const presets = useMemo(() => settings.jobQueryPresets ?? [], [settings.jobQueryPresets])
  const activePreset = useMemo(
    () => presets.find((p) => p.id === activePresetId),
    [presets, activePresetId],
  )

  const hasFilters = useMemo(() => {
    return Object.values(formValues).some((v) => v !== '')
  }, [formValues])

  const slotsRemaining = MAX_PRESETS - presets.length

  const handleSave = useCallback(() => {
    const trimmed = saveName.trim()
    if (!trimmed) return

    const newPreset: JobQueryPreset = {
      id: crypto.randomUUID(),
      name: trimmed,
      createdAt: new Date().toISOString(),
      filters: formValuesToFilters(formValues),
    }

    updateSettings((current) => ({
      ...current,
      jobQueryPresets: [...(current.jobQueryPresets ?? []), newPreset],
    }))

    toaster.create({
      title: 'Query saved',
      description: `"${trimmed}" has been saved`,
      type: 'success',
    })

    setSaveMode(false)
    setSaveName('')
    setMenuOpen(false)
  }, [saveName, formValues, updateSettings])

  const handleUpdate = useCallback(() => {
    if (!activePresetId) return

    updateSettings((current) => ({
      ...current,
      jobQueryPresets: (current.jobQueryPresets ?? []).map((p) =>
        p.id === activePresetId
          ? { ...p, filters: formValuesToFilters(formValues) }
          : p,
      ),
    }))

    toaster.create({
      title: 'Query updated',
      description: `"${activePreset?.name}" has been updated`,
      type: 'success',
    })

    setMenuOpen(false)
  }, [activePresetId, activePreset?.name, formValues, updateSettings])

  const handleDelete = useCallback(
    (presetId: string, presetName: string) => {
      updateSettings((current) => ({
        ...current,
        jobQueryPresets: (current.jobQueryPresets ?? []).filter((p) => p.id !== presetId),
      }))

      toaster.create({
        title: 'Query deleted',
        description: `"${presetName}" has been removed`,
        type: 'info',
      })
    },
    [updateSettings],
  )

  const handleLoad = useCallback(
    (preset: JobQueryPreset) => {
      onLoadPreset(preset.filters)
      setMenuOpen(false)
    },
    [onLoadPreset],
  )

  const handleSaveKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleSave()
    } else if (e.key === 'Escape') {
      e.preventDefault()
      setSaveMode(false)
      setSaveName('')
    }
  }

  return (
    <Popover.Root
      open={menuOpen}
      onOpenChange={(e) => {
        setMenuOpen(e.open)
        if (!e.open) {
          setSaveMode(false)
          setSaveName('')
        }
      }}
      positioning={{ placement: 'bottom-end' }}
      lazyMount
      unmountOnExit
    >
      <Popover.Trigger asChild>
        <Button
          variant="ghost"
          size="sm"
          gap={1.5}
          fontWeight="medium"
          color="fg.muted"
          _hover={{ color: 'fg', bg: 'gray.100' }}
          onClick={(e) => e.stopPropagation()}
        >
          <Icon fontSize="sm">
            <LuBookmark />
          </Icon>
          Saved Queries
          {presets.length > 0 && (
            <Badge
              size="sm"
              variant="solid"
              colorPalette="blue"
              fontFamily="mono"
              fontSize="2xs"
              px={1.5}
              minW="auto"
            >
              {presets.length}
            </Badge>
          )}
        </Button>
      </Popover.Trigger>

      <Portal>
        <Popover.Positioner>
          <Popover.Content
            w="360px"
            shadow="lg"
            borderRadius="lg"
            border="1px solid"
            borderColor="gray.200"
            p={0}
            overflow="hidden"
          >
            {saveMode ? (
              /* ── Save Form ── */
              <Box p={4}>
                <Text fontSize="sm" fontWeight="semibold" mb={3}>
                  Save Query
                </Text>
                <Input
                  ref={saveInputRef}
                  value={saveName}
                  onChange={(e) => setSaveName(e.target.value)}
                  placeholder="e.g. my-failed-jobs"
                  size="sm"
                  fontFamily="mono"
                  fontSize="sm"
                  autoFocus
                  maxLength={50}
                  onKeyDown={handleSaveKeyDown}
                />
                {saveName.length > 40 && (
                  <Text fontSize="2xs" color="orange.500" mt={1} fontFamily="mono">
                    {50 - saveName.length} chars remaining
                  </Text>
                )}
                <HStack mt={3} justify="flex-end" gap={2}>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setSaveMode(false)
                      setSaveName('')
                    }}
                  >
                    Cancel
                  </Button>
                  <Button
                    size="sm"
                    colorPalette="blue"
                    disabled={!saveName.trim() || isSaving}
                    loading={isSaving}
                    onClick={handleSave}
                  >
                    Save
                  </Button>
                </HStack>
              </Box>
            ) : (
              /* ── Presets Menu ── */
              <VStack gap={0} align="stretch">
                {/* Actions */}
                <Box px={2} pt={2} pb={1}>
                  {activePresetId && isDirty && (
                    <Box
                      as="button"
                      w="100%"
                      display="flex"
                      alignItems="center"
                      gap={2}
                      px={3}
                      py={2}
                      rounded="md"
                      fontSize="sm"
                      fontWeight="medium"
                      color="blue.600"
                      _hover={{ bg: 'blue.50' }}
                      transition="background 0.15s"
                      onClick={handleUpdate}
                    >
                      <Icon fontSize="sm">
                        <LuRefreshCw />
                      </Icon>
                      Update &quot;{activePreset?.name}&quot;
                    </Box>
                  )}

                  <Box
                    as="button"
                    w="100%"
                    display="flex"
                    alignItems="center"
                    gap={2}
                    px={3}
                    py={2}
                    rounded="md"
                    fontSize="sm"
                    _hover={{ bg: 'gray.100' }}
                    transition="background 0.15s"
                    disabled={!hasFilters || slotsRemaining <= 0}
                    opacity={!hasFilters || slotsRemaining <= 0 ? 0.5 : 1}
                    cursor={!hasFilters || slotsRemaining <= 0 ? 'not-allowed' : 'pointer'}
                    onClick={() => {
                      if (hasFilters && slotsRemaining > 0) setSaveMode(true)
                    }}
                  >
                    <Icon fontSize="sm">
                      <LuSave />
                    </Icon>
                    Save as new query...
                  </Box>
                </Box>

                {presets.length > 0 && <Separator />}

                {/* Preset List */}
                {presets.length === 0 ? (
                  <Box px={4} py={6} textAlign="center">
                    <Icon fontSize="2xl" color="gray.400" mb={2}>
                      <LuBookmarkPlus />
                    </Icon>
                    <Text fontSize="sm" color="gray.500" mb={1}>
                      No saved queries yet.
                    </Text>
                    <Text fontSize="xs" color="gray.400" mb={3}>
                      Save your current filters to quickly reload them later.
                    </Text>
                    {hasFilters && (
                      <Button
                        size="sm"
                        variant="outline"
                        colorPalette="blue"
                        onClick={() => setSaveMode(true)}
                      >
                        Save current filters
                      </Button>
                    )}
                  </Box>
                ) : (
                  <Box maxH="300px" overflowY="auto" px={2} py={1}>
                    {presets.map((preset) => {
                      const isActive = preset.id === activePresetId
                      return (
                        <Box
                          key={preset.id}
                          px={3}
                          py={2.5}
                          rounded="md"
                          cursor="pointer"
                          borderLeftWidth={isActive ? '3px' : '3px'}
                          borderLeftColor={isActive ? 'blue.500' : 'transparent'}
                          bg={isActive ? 'blue.50' : 'transparent'}
                          _hover={{ bg: isActive ? 'blue.100' : 'gray.50' }}
                          transition="background 0.15s"
                          onClick={() => handleLoad(preset)}
                          role="group"
                        >
                          <HStack justify="space-between" align="start">
                            <VStack align="start" gap={0.5} flex={1} minW={0}>
                              <HStack gap={1.5}>
                                <Text
                                  fontSize="sm"
                                  fontWeight="semibold"
                                  truncate
                                  maxW="200px"
                                >
                                  {preset.name}
                                </Text>
                                {isActive && (
                                  <Badge
                                    size="sm"
                                    colorPalette="blue"
                                    variant="subtle"
                                    fontSize="2xs"
                                  >
                                    active
                                  </Badge>
                                )}
                              </HStack>
                              <Text
                                fontSize="xs"
                                color="gray.500"
                                truncate
                                maxW="250px"
                              >
                                {buildFilterSummary(preset.filters)}
                              </Text>
                            </VStack>
                            <HStack gap={1} flexShrink={0}>
                              <Text
                                fontSize="xs"
                                color="gray.400"
                                fontFamily="mono"
                                whiteSpace="nowrap"
                              >
                                {formatRelativeTime(preset.createdAt)}
                              </Text>
                              <IconButton
                                aria-label="Delete query"
                                size="xs"
                                variant="ghost"
                                color="gray.400"
                                _hover={{ color: 'red.500', bg: 'red.50' }}
                                transition="color 0.15s"
                                onClick={(e) => {
                                  e.stopPropagation()
                                  handleDelete(preset.id, preset.name)
                                }}
                              >
                                <LuTrash2 />
                              </IconButton>
                            </HStack>
                          </HStack>
                        </Box>
                      )
                    })}
                  </Box>
                )}

                {/* Slots remaining warning */}
                {presets.length >= 18 && (
                  <>
                    <Separator />
                    <Box px={4} py={2}>
                      <Text fontSize="xs" color="gray.400" fontFamily="mono" textAlign="center">
                        {slotsRemaining} {slotsRemaining === 1 ? 'slot' : 'slots'} remaining ({MAX_PRESETS} max)
                      </Text>
                    </Box>
                  </>
                )}
              </VStack>
            )}
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
}

/* ── Active Preset Badge ── */

interface ActivePresetBadgeProps {
  preset: JobQueryPreset | undefined
  isDirty: boolean
}

export const ActivePresetBadge = ({ preset, isDirty }: ActivePresetBadgeProps) => {
  if (!preset) return null

  return (
    <HStack
      gap={1}
      px={2}
      py={0.5}
      bg="blue.50"
      rounded="md"
      borderWidth="1px"
      borderColor="blue.200"
    >
      <Box w="6px" h="6px" rounded="full" bg="blue.500" />
      <Text fontSize="xs" color="blue.700" fontWeight="medium" truncate maxW="180px">
        {preset.name}
      </Text>
      {isDirty && (
        <Box w="6px" h="6px" rounded="full" bg="orange.400" />
      )}
    </HStack>
  )
}
