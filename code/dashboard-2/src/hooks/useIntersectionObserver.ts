import { useState, useCallback, useRef } from 'react'

interface UseIntersectionObserverOptions {
  rootMargin?: string
  threshold?: number | number[]
}

export function useIntersectionObserver(options: UseIntersectionObserverOptions = {}) {
  const [hasBeenVisible, setHasBeenVisible] = useState(false)
  const observerRef = useRef<IntersectionObserver | null>(null)

  const ref = useCallback(
    (node: Element | null) => {
      if (observerRef.current) {
        observerRef.current.disconnect()
        observerRef.current = null
      }

      if (hasBeenVisible || !node) return

      observerRef.current = new IntersectionObserver(
        ([entry]) => {
          if (entry.isIntersecting) {
            setHasBeenVisible(true)
            observerRef.current?.disconnect()
            observerRef.current = null
          }
        },
        {
          rootMargin: options.rootMargin ?? '200px',
          threshold: options.threshold ?? 0,
        },
      )

      observerRef.current.observe(node)
    },
    [hasBeenVisible, options.rootMargin, options.threshold],
  )

  return { ref, hasBeenVisible }
}
