import React, { useEffect, useState } from 'react'
import { Box, type BoxProps } from '@chakra-ui/react'

export interface ResizableColumnsProps {
  left: React.ReactNode
  right: React.ReactNode
  initialLeftWidth?: number
  minLeftWidth?: number
  maxLeftWidth?: number
  handleWidth?: number
  storageKey?: string
  height?: string
  leftProps?: BoxProps
  rightProps?: BoxProps
  handleProps?: BoxProps
}

export const ResizableColumns: React.FC<ResizableColumnsProps> = ({
  left,
  right,
  initialLeftWidth = 320,
  minLeftWidth = 220,
  maxLeftWidth = 640,
  handleWidth = 6,
  storageKey = 'resizable.leftWidth',
  height,
  leftProps,
  rightProps,
  handleProps,
}) => {
  const [leftWidth, setLeftWidth] = useState<number>(initialLeftWidth)

  useEffect(() => {
    try {
      const saved = localStorage.getItem(storageKey)
      if (saved) {
        const n = Number(saved)
        if (!Number.isNaN(n)) setLeftWidth(Math.min(Math.max(n, minLeftWidth), maxLeftWidth))
      }
    } catch { /* ignore */ }
  }, [storageKey])

  const startDrag = (startX: number) => {
    const startWidth = leftWidth
    const onMove = (ev: MouseEvent) => {
      const delta = ev.clientX - startX
      const next = Math.min(Math.max(startWidth + delta, minLeftWidth), maxLeftWidth)
      setLeftWidth(next)
    }
    const onUp = () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
      try { localStorage.setItem(storageKey, String(leftWidth)) } catch { /* ignore */ }
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
  }

  const onMouseDownHandle = (e: React.MouseEvent) => {
    e.preventDefault()
    startDrag(e.clientX)
  }

  const onTouchStartHandle = (e: React.TouchEvent) => {
    const touch = e.touches[0]
    if (!touch) return
    e.preventDefault()
    const startX = touch.clientX
    const startWidth = leftWidth
    const onTouchMove = (ev: TouchEvent) => {
      const t = ev.touches[0]
      if (!t) return
      const delta = t.clientX - startX
      const next = Math.min(Math.max(startWidth + delta, minLeftWidth), maxLeftWidth)
      setLeftWidth(next)
    }
    const onTouchEnd = () => {
      window.removeEventListener('touchmove', onTouchMove)
      window.removeEventListener('touchend', onTouchEnd)
      try { localStorage.setItem(storageKey, String(leftWidth)) } catch { /* ignore */ }
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
    window.addEventListener('touchmove', onTouchMove)
    window.addEventListener('touchend', onTouchEnd)
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
  }

  return (
    <Box display="flex" gap={0} alignItems="stretch" height={height} width="100%">
      <Box
        width={`${leftWidth}px`}
        minW={`${minLeftWidth}px`}
        maxW={`${maxLeftWidth}px`}
        overflow="hidden"
        borderRight="1px"
        borderColor="gray.200"
        {...leftProps}
      >
        {left}
      </Box>
      <Box
        role="separator"
        aria-orientation="vertical"
        width={`${handleWidth}px`}
        cursor="col-resize"
        bg="gray.100"
        _hover={{ bg: 'gray.200' }}
        _active={{ bg: 'gray.300' }}
        onMouseDown={onMouseDownHandle}
        onTouchStart={onTouchStartHandle}
        {...handleProps}
      />
      <Box flex={1} minW={0} {...rightProps}>
        {right}
      </Box>
    </Box>
  )
}
