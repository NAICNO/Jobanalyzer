import {
  createBrowserRouter,
  Navigate,
  RouterProvider
} from 'react-router'

import RootLayout from './layouts/RootLayout.tsx'

import { ClusterOverview } from './pages/ClusterOverview.tsx'
import { NodesPage } from './pages/NodesPage.tsx'
import { PartitionsPage } from './pages/PartitionsPage.tsx'
import { JobsPage } from './pages/JobsPage.tsx'
import { JobDetailsPage } from './pages/JobDetailsPage.tsx'
import { QueriesPage } from './pages/QueriesPage.tsx'
import { NodeTopologyPage } from './pages/NodeTopologyPage.tsx'
import { ClusterSelectionPage } from './pages/ClusterSelectionPage.tsx'
import { ProcessTreeFullViewPage } from './pages/ProcessTreeFullViewPage.tsx'
import { BenchmarksPage } from './pages/BenchmarksPage.tsx'
import { ClusterRouteGuard } from './components/ClusterRouteGuard.tsx'
import { CallbackPage } from './pages/auth/CallbackPage.tsx'
import { APP_BASE_PREFIX } from './Constants.ts'

const router = createBrowserRouter([
  {
    path: '/:clusterName/jobs/:jobId/process-tree',
    element: <ClusterRouteGuard />,
    children: [
      {
        index: true,
        element: <ProcessTreeFullViewPage />,
      },
    ],
  },
  {
    path: '/:clusterName/nodes/:nodename/topology',
    element: <ClusterRouteGuard />,
    children: [
      {
        index: true,
        element: <NodeTopologyPage />,
      },
    ],
  },
  {
    path: '/',
    element: <RootLayout/>,
    children: [
      {
        index: true,
        element: <Navigate to="select-cluster" replace/>
      },
      {
        path: 'auth',
        children: [
          {
            path: 'callback',
            element: <CallbackPage />,
          },
        ],
      },
      {
        path: 'select-cluster',
        element: <ClusterSelectionPage />,
      },
      {
        path: ':clusterName',
        element: <ClusterRouteGuard />,
        children: [
          {
            path: 'overview',
            element: <ClusterOverview />,
          },
          {
            path: 'nodes',
            element: <NodesPage />,
          },
          {
            path: 'nodes/:nodename',
            element: <NodesPage />,
          },
          {
            path: 'partitions',
            element: <PartitionsPage />,
          },
          {
            path: 'partitions/:partitionName',
            element: <PartitionsPage />,
          },
          {
            path: 'jobs',
            element: <JobsPage />,
          },
          {
            path: 'jobs/running',
            element: <JobsPage />,
          },
          {
            path: 'jobs/query',
            element: <QueriesPage />,
          },
          {
            path: 'jobs/:jobId',
            element: <JobDetailsPage />,
          },
          {
            path: 'benchmarks',
            element: <BenchmarksPage />,
          }
        ]
      },
    ]
  }
], { basename: APP_BASE_PREFIX })

function App() {
  return (
    <RouterProvider router={router}/>
  )
}

export default App
