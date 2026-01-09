import { LuServer } from 'react-icons/lu'
import { GrNodes } from 'react-icons/gr'
import { GiFox } from 'react-icons/gi'
import type { IconType } from 'react-icons'

export interface OIDCEndpoints {
  authorization: string
  token: string
  userInfo: string
}

export interface ClusterConfig {
  id: string
  name: string
  description?: string
  icon: IconType
  apiBaseUrl: string
  authEndpoint: OIDCEndpoints
}

// Available clusters configuration
export const AVAILABLE_CLUSTERS: ClusterConfig[] = [
  {
    id: 'ex3.simula.no',
    name: 'eX3',
    description: 'eX3 Cluster at Simula',
    icon: LuServer,
    apiBaseUrl: 'https://localhost:12200/api/v2',
    authEndpoint: {
      authorization: 'https://auth.ex3.simula.no/authorize',
      token: 'https://auth.ex3.simula.no/token',
      userInfo: 'https://auth.ex3.simula.no/userinfo',
    },
  },
  {
    id: 'mlx.hpc.uio.no',
    name: 'ML Nodes',
    description: 'Machine Learning Cluster at UiO',
    icon: GrNodes,
    apiBaseUrl: 'https://naic-monitor.uio.no/api/v2',
    authEndpoint: {
      authorization: 'https://auth.mlx.hpc.uio.no/authorize',
      token: 'https://auth.mlx.hpc.uio.no/token',
      userInfo: 'https://auth.mlx.hpc.uio.no/userinfo',
    },
  },
  {
    id: 'fox.educloud.no',
    name: 'Fox',
    description: 'Fox Cluster at EduCloud',
    icon: GiFox,
    apiBaseUrl: 'https://naic-monitor.uio.no/api/v2',
    authEndpoint: {
      authorization: 'https://auth.fox.educloud.no/authorize',
      token: 'https://auth.fox.educloud.no/token',
      userInfo: 'https://auth.fox.educloud.no/userinfo',
    },
  },
]

// Helper to get cluster config by ID
export const getClusterConfig = (clusterId: string): ClusterConfig | undefined => {
  return AVAILABLE_CLUSTERS.find((cluster) => cluster.id === clusterId)
}

// Helper to get full cluster name (e.g., 'mlx.hpc.uio.no') for URL routing
// Since cluster IDs are now full names, this just returns the ID
export const getClusterFullName = (clusterId: string): string => {
  return clusterId
}

// Helper to get cluster ID from full name
// Since cluster IDs are now full names, this just returns the input
export const getClusterIdFromFullName = (fullName: string): string => {
  return fullName
}
