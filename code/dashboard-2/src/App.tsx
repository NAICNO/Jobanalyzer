import {
  createBrowserRouter,
  Navigate,
  RouterProvider
} from 'react-router'

import RootLayout from './layouts/RootLayout.tsx'
import DashboardPage from './pages/DashboardPage.tsx'
import ViolatorsPage from './pages/ViolatorsPage.tsx'
import ViolatorPage from './pages/ViolatorPage.tsx'
import DeadWeightPage from './pages/DeadWeightPage.tsx'
import NodeSelectionHelpPage from './pages/NodeSelectionHelpPage.tsx'
import HostDetailsPage from './pages/HostDetailsPage.tsx'
import JobQueryPage from './pages/JobQueryPage.tsx'
import SubclusterPage from './pages/SubclusterPage.tsx'
import JobProfilePage from './pages/JobProfilePage.tsx'
import JobProcessTreePage from './pages/JobProcessTreePage.tsx'

import { ClusterOverview } from './pages/v2/ClusterOverview.tsx'
import { NodesPage } from './pages/v2/NodesPage.tsx'
import { PartitionsPage } from './pages/v2/PartitionsPage.tsx'
import { JobsPage } from './pages/v2/JobsPage.tsx'
import { JobDetailsPage } from './pages/v2/JobDetailsPage.tsx'
import { QueriesPage } from './pages/v2/QueriesPage.tsx'
import { NodeTopologyPage } from './pages/v2/NodeTopologyPage.tsx'
import { ClusterSelectionPage } from './pages/v2/ClusterSelectionPage.tsx'
import { ProcessTreeFullViewPage } from './pages/v2/ProcessTreeFullViewPage.tsx'
import { BenchmarksPage } from './pages/v2/BenchmarksPage.tsx'
import { ClusterRouteGuard } from './components/ClusterRouteGuard.tsx'
import { CallbackPage } from './pages/auth/CallbackPage.tsx'

const router = createBrowserRouter([
  {
    path: '/v2/:clusterName/jobs/:jobId/process-tree',
    element: <ClusterRouteGuard />,
    children: [
      {
        index: true,
        element: <ProcessTreeFullViewPage />,
      },
    ],
  },
  {
    path: '/v2/:clusterName/nodes/:nodename/topology',
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
        element: <Navigate to="v2/select-cluster" replace/>
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
        path: 'v2',
        children: [
          {
            index: true,
            element: <Navigate to="select-cluster" replace/>
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
          }
        ]
      },
      {
        path: 'dashboard',
        children: [
          {
            index: true,
            element: <DashboardPage/>,
          },
          {
            path: ':clusterName',
            element: <DashboardPage/>,
          },
          {
            path: 'help/node-selection',
            element: <NodeSelectionHelpPage/>,
          },
        ]
      },
      {
        path: ':clusterName',
        children: [
          {
            path: 'subcluster/:subclusterName',
            element: <SubclusterPage/>,
          },
          {
            path: 'violators',
            element: <ViolatorsPage/>,
          },
          {
            path: 'violators/:violator',
            element: <ViolatorPage/>,
          },
          {
            path: 'deadweight',
            element: <DeadWeightPage/>,
          },
          {
            path: ':hostname',
            children: [
              {
                index: true,
                element: <HostDetailsPage/>,
              },
              {
                path: 'violators',
                element: <ViolatorsPage/>,
              },
              {
                path: 'violators/:violator',
                element: <ViolatorPage/>,
              },
              {
                path: 'deadweight',
                element: <DeadWeightPage/>,
              }
            ]
          }
        ]
      },
      {
        path: 'jobquery',
        children: [
          {
            index: true,
            element: <JobQueryPage/>,
          },
          {
            path: 'profile',
            element: <JobProfilePage/>,
          },
          {
            path: 'tree',
            element: <JobProcessTreePage/>,
          }
        ]
      },
    ]
  }
])

function App() {
  return (
    <RouterProvider router={router}/>
  )
}

export default App
