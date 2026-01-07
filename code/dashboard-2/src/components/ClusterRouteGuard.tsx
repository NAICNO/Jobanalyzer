import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router'
import { useCluster } from '../hooks/useCluster'

/**
 * Route guard that redirects to cluster selection page if user has no clusters selected
 * Should wrap v2 routes that require a cluster
 */
export const ClusterRouteGuard = ({ children }: { children: React.ReactNode }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const { hasSelectedClusters } = useCluster()

  useEffect(() => {
    // Don't redirect if we're already on the selection page
    if (location.pathname === '/v2/select-cluster') {
      return
    }

    // Redirect to selection page if no clusters are selected
    if (!hasSelectedClusters) {
      navigate('/v2/select-cluster', { replace: true })
    }
  }, [hasSelectedClusters, navigate, location.pathname])

  // If no clusters selected and not on selection page, don't render children
  if (!hasSelectedClusters && location.pathname !== '/v2/select-cluster') {
    return null
  }

  return <>{children}</>
}
