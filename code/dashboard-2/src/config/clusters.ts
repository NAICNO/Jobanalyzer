import type { IconType } from 'react-icons'
import { resolveIcon } from './iconRegistry'
import { APP_BASE_PREFIX } from '../Constants'

const basePrefix = APP_BASE_PREFIX.endsWith('/') ? APP_BASE_PREFIX : APP_BASE_PREFIX + '/'

export interface OIDCEndpoints {
  authorization: string
  token: string
  userInfo: string
  clientId: string
  redirectUri: string
  scope?: string
}

export interface ClusterConfig {
  id: string
  name: string
  shortName: string
  description?: string
  icon: IconType
  apiBaseUrl: string
  authEndpoint: OIDCEndpoints
  requiresAuth: boolean
}

interface ClusterConfigJson {
  id: string
  name: string
  shortName: string
  description?: string
  icon?: string
  apiBaseUrl: string
  authEndpoint: {
    authorization: string
    token: string
    userInfo: string
    clientId: string
    redirectUri?: string
    scope?: string
  }
  requiresAuth: boolean
}

interface ClustersJson {
  clusters: ClusterConfigJson[]
}

// Module-level array — mutated in place so all importers see the same reference
export const AVAILABLE_CLUSTERS: ClusterConfig[] = []

let _loaded = false

function validateClustersJson(json: unknown): asserts json is ClustersJson {
  if (!json || typeof json !== 'object') {
    throw new Error('clusters.json must contain a JSON object')
  }

  const obj = json as Record<string, unknown>
  if (!Array.isArray(obj.clusters) || obj.clusters.length === 0) {
    throw new Error('clusters.json must contain a non-empty "clusters" array')
  }

  for (const [i, cluster] of obj.clusters.entries()) {
    const prefix = `clusters[${i}]`
    if (!cluster || typeof cluster !== 'object') {
      throw new Error(`${prefix} must be an object`)
    }
    const c = cluster as Record<string, unknown>
    for (const field of ['id', 'name', 'shortName', 'apiBaseUrl'] as const) {
      if (typeof c[field] !== 'string' || (c[field] as string).length === 0) {
        throw new Error(`${prefix}.${field} must be a non-empty string`)
      }
    }
    if (typeof c.requiresAuth !== 'boolean') {
      throw new Error(`${prefix}.requiresAuth must be a boolean`)
    }
    if (!c.authEndpoint || typeof c.authEndpoint !== 'object') {
      throw new Error(`${prefix}.authEndpoint must be an object`)
    }
    const auth = c.authEndpoint as Record<string, unknown>
    for (const field of ['authorization', 'token', 'userInfo', 'clientId'] as const) {
      if (typeof auth[field] !== 'string' || (auth[field] as string).length === 0) {
        throw new Error(`${prefix}.authEndpoint.${field} must be a non-empty string`)
      }
    }
  }
}

export async function loadClusterConfig(): Promise<ClusterConfig[]> {
  if (_loaded) return AVAILABLE_CLUSTERS

  const response = await fetch(`${basePrefix}clusters.json`)
  if (!response.ok) {
    throw new Error(`Failed to load clusters.json: ${response.status} ${response.statusText}`)
  }

  const json: unknown = await response.json()
  validateClustersJson(json)

  const defaultRedirectUri = window.location.origin + basePrefix + 'auth/callback'

  // Clear and populate the shared array in place
  AVAILABLE_CLUSTERS.length = 0
  for (const raw of json.clusters) {
    AVAILABLE_CLUSTERS.push({
      id: raw.id,
      name: raw.name,
      shortName: raw.shortName,
      description: raw.description,
      icon: resolveIcon(raw.icon),
      apiBaseUrl: raw.apiBaseUrl,
      authEndpoint: {
        authorization: raw.authEndpoint.authorization,
        token: raw.authEndpoint.token,
        userInfo: raw.authEndpoint.userInfo,
        clientId: raw.authEndpoint.clientId,
        redirectUri: raw.authEndpoint.redirectUri ?? defaultRedirectUri,
        scope: raw.authEndpoint.scope,
      },
      requiresAuth: raw.requiresAuth,
    })
  }

  _loaded = true
  return AVAILABLE_CLUSTERS
}

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
