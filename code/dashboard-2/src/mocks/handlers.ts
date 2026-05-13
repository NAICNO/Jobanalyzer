/**
 * MSW request handlers for demo mode.
 *
 * Intercepts all API calls to the demo cluster and returns synthetic data.
 * The base URL matches the demo cluster's apiBaseUrl from clusters.demo.json.
 */
import { http, HttpResponse } from 'msw'

import { mockClusterResponse, CLUSTER_NAME, NODE_NAMES } from './data/cluster'
import { mockNodesInfo, mockNodeStates, mockLastProbeTimestamps, getNodeInfo, getNodeStates } from './data/nodes'
import { mockPartitions } from './data/partitions'
import { mockJobsResponse, mockJobs, getJobById } from './data/jobs'
import { mockDetailedJob, mockJobReport } from './data/job-details'
import { mockErrorMessages, mockNodeErrorMessages } from './data/error-messages'
import { mockProcessTree } from './data/process-tree'
import { getBenchmarkRecords } from './data/benchmarks'
import {
  generateCpuTimeseries,
  generateMemoryTimeseries,
  generateGpuTimeseries,
  generateDiskTimeseries,
  generateNodeGpuUtil,
  generateNodeCpuTimeseries,
  generateNodeMemoryTimeseries,
  generateNodeGpuTimeseries,
  generateNodeDiskTimeseries,
  generateJobProcessTimeseries,
  generateJobGpuTimeseries,
} from './data/timeseries'

const BASE = 'https://demo.hpc.example.org/api/v2'

export const handlers = [
  // =========================================================================
  // Root / health
  // =========================================================================

  http.get(`${BASE}/`, () => HttpResponse.json({ message: 'ok' })),

  // =========================================================================
  // Cluster
  // =========================================================================

  http.get(`${BASE}/cluster`, () =>
    HttpResponse.json([mockClusterResponse]),
  ),

  // =========================================================================
  // Nodes
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/nodes`, ({ params }) => {
    if (params.cluster !== CLUSTER_NAME) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json(NODE_NAMES)
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/info`, () =>
    HttpResponse.json(mockNodesInfo),
  ),

  http.get(`${BASE}/cluster/:cluster/nodes/info/pages`, ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') ?? '1')
    const pageSize = parseInt(url.searchParams.get('page_size') ?? '25')
    const allNodes = Object.values(mockNodesInfo)
    const total = allNodes.length
    const pages = Math.ceil(total / pageSize)
    const start = (page - 1) * pageSize
    return HttpResponse.json({
      nodes: allNodes.slice(start, start + pageSize),
      total,
      page,
      size: pageSize,
      pages,
    })
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/states`, () =>
    HttpResponse.json(mockNodeStates),
  ),

  http.get(`${BASE}/cluster/:cluster/nodes/last_probe_timestamp`, () =>
    HttpResponse.json(mockLastProbeTimestamps),
  ),

  // Single node info (returns dict keyed by node name)
  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/info`, ({ params }) => {
    const nodename = params.nodename as string
    const info = getNodeInfo(nodename)
    if (!info) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json({ [nodename]: info })
  }),

  // Single node states
  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/states`, ({ params }) => {
    const state = getNodeStates(params.nodename as string)
    if (!state) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json(state)
  }),

  // Single node error messages
  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/error_messages`, ({ params }) => {
    return HttpResponse.json(mockNodeErrorMessages(params.nodename as string))
  }),

  // Node topology (SVG)
  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/topology`, () => {
    // Return empty — topology is optional
    return HttpResponse.json({ topo_svg: null, topo_text: null })
  }),

  // =========================================================================
  // Node GPU utilization
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/nodes/process/gpu/util`, () =>
    HttpResponse.json(generateNodeGpuUtil()),
  ),

  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/process/gpu/util`, ({ params }) => {
    const nodeName = params.nodename as string
    const all = generateNodeGpuUtil([nodeName])
    return HttpResponse.json(all)
  }),

  // =========================================================================
  // Cluster-level timeseries
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/nodes/cpu/timeseries`, ({ request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateCpuTimeseries(undefined, hours))
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/memory/timeseries`, ({ request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateMemoryTimeseries(undefined, hours))
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/gpu/timeseries`, ({ request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateGpuTimeseries(undefined, hours))
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/diskstats/timeseries`, ({ request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateDiskTimeseries(undefined, hours))
  }),

  // =========================================================================
  // Per-node timeseries
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/cpu/timeseries`, ({ params, request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateNodeCpuTimeseries(params.nodename as string, hours))
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/memory/timeseries`, ({ params, request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateNodeMemoryTimeseries(params.nodename as string, hours))
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/gpu/timeseries`, ({ params, request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateNodeGpuTimeseries(params.nodename as string, hours))
  }),

  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/diskstats/timeseries`, ({ params, request }) => {
    const hours = getTimeRangeHours(request)
    return HttpResponse.json(generateNodeDiskTimeseries(params.nodename as string, hours))
  }),

  // =========================================================================
  // Node-level process GPU timeseries
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/nodes/process/gpu/timeseries`, () =>
    HttpResponse.json({}),
  ),

  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/process/gpu/timeseries`, () =>
    HttpResponse.json({}),
  ),

  // =========================================================================
  // Partitions
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/partitions`, () =>
    HttpResponse.json(mockPartitions),
  ),

  // =========================================================================
  // Jobs
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/jobs`, () =>
    HttpResponse.json(mockJobsResponse),
  ),

  // Single job by ID
  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/info`, ({ params }) => {
    const jobId = parseInt(params.jobId as string)
    const job = getJobById(jobId)
    if (!job) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json(job)
  }),

  http.get(`${BASE}/cluster/:cluster/jobs/:jobId`, ({ params }) => {
    const jobId = parseInt(params.jobId as string)
    // Prefer the detailed version for job detail pages
    if (jobId === mockDetailedJob.job_id) {
      return HttpResponse.json(mockDetailedJob)
    }
    const job = getJobById(jobId)
    if (!job) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json(job)
  }),

  // Job report
  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/report`, () =>
    HttpResponse.json(mockJobReport),
  ),

  // Job process timeseries (API returns an array)
  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/process/timeseries`, ({ params }) => {
    const jobId = parseInt(params.jobId as string)
    const job = getJobById(jobId)
    const nodeName = job?.nodes?.[0] ?? 'gpu001'
    return HttpResponse.json([{
      job: jobId,
      epoch: 0,
      nodes: {
        [nodeName]: generateJobProcessTimeseries(nodeName),
      },
    }])
  }),

  // Job process GPU timeseries (API returns an array)
  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/process/gpu/timeseries`, ({ params }) => {
    const jobId = parseInt(params.jobId as string)
    const job = getJobById(jobId)
    const nodeName = job?.nodes?.[0] ?? 'gpu001'
    return HttpResponse.json([{
      job: jobId,
      epoch: 0,
      nodes: {
        [nodeName]: generateJobGpuTimeseries(nodeName),
      },
    }])
  }),

  // Job process tree
  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/process/tree`, () =>
    HttpResponse.json(mockProcessTree),
  ),

  // Job epoch endpoints
  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/epoch/:epoch`, ({ params }) => {
    const jobId = parseInt(params.jobId as string)
    const job = getJobById(jobId) ?? mockDetailedJob
    return HttpResponse.json(job)
  }),

  http.get(`${BASE}/cluster/:cluster/jobs/:jobId/epoch/:epoch/info`, ({ params }) => {
    const jobId = parseInt(params.jobId as string)
    const job = getJobById(jobId) ?? mockDetailedJob
    return HttpResponse.json(job)
  }),

  // Node-specific job GPU timeseries
  http.get(`${BASE}/cluster/:cluster/nodes/:nodename/jobs/:jobId/process/gpu/timeseries`, ({ params }) => {
    const nodeName = params.nodename as string
    return HttpResponse.json(generateJobGpuTimeseries(nodeName))
  }),

  // =========================================================================
  // Job query (paginated)
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/query/jobs`, () =>
    HttpResponse.json(mockJobsResponse),
  ),

  http.get(`${BASE}/cluster/:cluster/query/jobs/pages`, ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') ?? '1')
    const pageSize = parseInt(url.searchParams.get('page_size') ?? '25')
    const total = mockJobs.length
    const pages = Math.ceil(total / pageSize)
    const start = (page - 1) * pageSize
    return HttpResponse.json({
      jobs: mockJobs.slice(start, start + pageSize),
      total,
      page,
      size: pageSize,
      pages,
    })
  }),

  // =========================================================================
  // Jobs process timeseries (cluster-wide)
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/jobs/process/timeseries`, () =>
    HttpResponse.json({}),
  ),

  http.get(`${BASE}/cluster/:cluster/jobs/process/gpu/timeseries`, () =>
    HttpResponse.json({}),
  ),

  // =========================================================================
  // Queries (predefined)
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/queries`, () =>
    HttpResponse.json({ queries: ['recent-gpu-jobs', 'failed-jobs-24h', 'long-running'] }),
  ),

  http.get(`${BASE}/cluster/:cluster/queries/:queryName`, () =>
    HttpResponse.json(mockJobsResponse),
  ),

  // =========================================================================
  // Benchmarks
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/benchmarks/:benchmarkName`, ({ params }) =>
    HttpResponse.json(getBenchmarkRecords(params.benchmarkName as string)),
  ),

  // =========================================================================
  // Error messages (cluster-wide)
  // =========================================================================

  http.get(`${BASE}/cluster/:cluster/error_messages`, () =>
    HttpResponse.json(mockErrorMessages),
  ),

  // =========================================================================
  // User / settings (no-auth cluster, return sensible defaults)
  // =========================================================================

  http.get(`${BASE}/user`, () =>
    HttpResponse.json({
      exp: Math.floor(Date.now() / 1000) + 3600,
      iat: Math.floor(Date.now() / 1000),
      jti: 'demo-jti',
      iss: 'demo',
      aud: 'demo',
      sub: 'demo-user',
      typ: 'Bearer',
      azp: 'demo-client',
      sid: 'demo-session',
      acr: 1,
      realm_access: { roles: ['user'] },
      resource_access: { account: { roles: ['user'] } },
      scope: 'openid profile email',
      email_verified: true,
      name: 'Demo User',
      preferred_username: 'demo_user',
      given_name: 'Demo',
      family_name: 'User',
      email: 'demo@example.org',
    }),
  ),

  http.get(`${BASE}/user/settings`, () =>
    HttpResponse.json(null),
  ),

  http.post(`${BASE}/user/settings`, () =>
    HttpResponse.json({ status: 'ok' }),
  ),

  http.put(`${BASE}/user/settings`, () =>
    HttpResponse.json({ status: 'ok' }),
  ),

  // =========================================================================
  // Cache clear (no-op)
  // =========================================================================

  http.get(`${BASE}/clear_cache`, () =>
    HttpResponse.json({ status: 'ok' }),
  ),
]

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Extract the time-range span in hours from query params. */
function getTimeRangeHours(request: Request): number {
  const url = new URL(request.url)
  const start = url.searchParams.get('start_time_in_s')
  const end = url.searchParams.get('end_time_in_s')
  if (start && end) {
    return Math.max(1, Math.round((parseInt(end) - parseInt(start)) / 3600))
  }
  return 24 // default 1 day
}
