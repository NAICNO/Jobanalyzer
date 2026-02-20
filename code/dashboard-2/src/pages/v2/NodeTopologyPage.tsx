import { Spinner, Alert, Box, HStack, Text, Badge, IconButton } from '@chakra-ui/react'
import { useParams } from 'react-router'
import { useCallback, useMemo, useRef, useState } from 'react'
import { LuMinus, LuPlus, LuMaximize } from 'react-icons/lu'
import DOMPurify from 'dompurify'

import { useClusterClient } from '../../hooks/useClusterClient'
import { useNodeTopology } from '../../hooks/v2/useNodeQueries'

const MIN_SCALE = 0.1
const MAX_SCALE = 5
const ZOOM_STEP = 0.15

export const NodeTopologyPage = () => {
  const { clusterName, nodename } = useParams()

  const client = useClusterClient(clusterName)
  const { data, isLoading, isError, error } = useNodeTopology({ cluster: clusterName ?? '', nodename: nodename ?? '', client })

  const svg = useMemo<string | undefined>(() => {
    if (!data) return undefined
    let raw: string | undefined
    if (typeof data === 'string') {
      raw = data
    } else if (typeof data === 'object') {
      const obj = data as Record<string, unknown>
      raw = (obj.svg as string | undefined)
        || (obj.body as string | undefined)
        || (obj.data as string | undefined)
        || undefined
    }
    return raw ? DOMPurify.sanitize(raw, { USE_PROFILES: { svg: true } }) : undefined
  }, [data])

  // Pan & zoom state
  const [scale, setScale] = useState(1)
  const [translate, setTranslate] = useState({ x: 0, y: 0 })
  const isPanning = useRef(false)
  const panStart = useRef({ x: 0, y: 0 })
  const containerRef = useRef<HTMLDivElement>(null)

  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault()
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    const cursorX = e.clientX - rect.left
    const cursorY = e.clientY - rect.top

    setScale(prev => {
      const direction = e.deltaY < 0 ? 1 : -1
      const next = Math.min(MAX_SCALE, Math.max(MIN_SCALE, prev * (1 + direction * ZOOM_STEP)))
      const ratio = next / prev

      setTranslate(t => ({
        x: cursorX - ratio * (cursorX - t.x),
        y: cursorY - ratio * (cursorY - t.y),
      }))

      return next
    })
  }, [])

  const handlePointerDown = useCallback((e: React.PointerEvent) => {
    if (e.button !== 0) return
    isPanning.current = true
    panStart.current = { x: e.clientX - translate.x, y: e.clientY - translate.y }
    ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
  }, [translate])

  const handlePointerMove = useCallback((e: React.PointerEvent) => {
    if (!isPanning.current) return
    setTranslate({
      x: e.clientX - panStart.current.x,
      y: e.clientY - panStart.current.y,
    })
  }, [])

  const handlePointerUp = useCallback(() => {
    isPanning.current = false
  }, [])

  const zoomIn = useCallback(() => {
    setScale(prev => Math.min(MAX_SCALE, prev * (1 + ZOOM_STEP)))
  }, [])

  const zoomOut = useCallback(() => {
    setScale(prev => Math.max(MIN_SCALE, prev * (1 - ZOOM_STEP)))
  }, [])

  const resetView = useCallback(() => {
    setScale(1)
    setTranslate({ x: 0, y: 0 })
  }, [])

  if (!clusterName || !nodename) {
    return (
      <Box w="100vw" h="100vh" p={4}>
        <Alert.Root status="error">
          <Alert.Indicator />
          <Alert.Description>Missing cluster or node name in route.</Alert.Description>
        </Alert.Root>
      </Box>
    )
  }

  return (
    <Box w="100vw" h="100vh" overflow="hidden">
      <HStack px={4} py={2} borderBottomWidth="1px" borderColor="border" justify="space-between">
        <HStack gap={2}>
          <Text fontSize="sm" fontWeight="bold">Node Topology</Text>
          <Badge size="sm" variant="subtle">{nodename}</Badge>
          <Text fontSize="xs" color="fg.muted">{clusterName}</Text>
        </HStack>
        <HStack gap={1}>
          <Text fontSize="xs" color="fg.muted">{Math.round(scale * 100)}%</Text>
          <IconButton aria-label="Zoom in" size="xs" variant="ghost" onClick={zoomIn}><LuPlus /></IconButton>
          <IconButton aria-label="Zoom out" size="xs" variant="ghost" onClick={zoomOut}><LuMinus /></IconButton>
          <IconButton aria-label="Reset view" size="xs" variant="ghost" onClick={resetView}><LuMaximize /></IconButton>
        </HStack>
      </HStack>
      <Box
        ref={containerRef}
        w="100%"
        h="calc(100vh - 41px)"
        overflow="hidden"
        cursor={isPanning.current ? 'grabbing' : 'grab'}
        onWheel={handleWheel}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
      >
        {isLoading && (
          <Box display="flex" alignItems="center" justifyContent="center" h="100%">
            <Spinner size="lg" />
          </Box>
        )}
        {isError && (
          <Box p={4}>
            <Alert.Root status="error">
              <Alert.Indicator />
              <Alert.Description>
                {error instanceof Error ? error.message : 'Failed to load topology.'}
              </Alert.Description>
            </Alert.Root>
          </Box>
        )}
        {!isLoading && !isError && svg && (
          <Box
            style={{
              transform: `translate(${translate.x}px, ${translate.y}px) scale(${scale})`,
              transformOrigin: '0 0',
            }}
            css={{
              '& svg': { overflow: 'visible' },
              '& svg text': {
                fontSize: '10px',
              },
            }}
            dangerouslySetInnerHTML={{ __html: svg }}
          />
        )}
        {!isLoading && !isError && !svg && (
          <Box display="flex" alignItems="center" justifyContent="center" h="100%">
            <Text color="fg.muted">No topology available.</Text>
          </Box>
        )}
      </Box>
    </Box>
  )
}
