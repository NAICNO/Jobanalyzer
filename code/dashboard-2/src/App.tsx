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
import { QueriesPage } from './pages/v2/QueriesPage.tsx'
import { NodeTopologyPage } from './pages/v2/NodeTopologyPage.tsx'
import { ClusterSelectionPage } from './pages/v2/ClusterSelectionPage.tsx'
import { ClusterRouteGuard } from './components/ClusterRouteGuard.tsx'

const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout/>,
    children: [
      {
        index: true,
        element: <Navigate to="v2/mlx.hpc.uio.no/overview" replace/>
      },
      {
        path: 'v2',
        children: [
          {
            path: 'select-cluster',
            element: <ClusterSelectionPage />,
          },
          {
            path: ':clusterName/overview',
            element: <ClusterRouteGuard><ClusterOverview /></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/nodes',
            element: <ClusterRouteGuard><NodesPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/nodes/:nodename',
            element: <ClusterRouteGuard><NodesPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/nodes/:nodename/topology',
            element: <ClusterRouteGuard><NodeTopologyPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/partitions',
            element: <ClusterRouteGuard><PartitionsPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/partitions/:partitionName',
            element: <ClusterRouteGuard><PartitionsPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/jobs',
            element: <ClusterRouteGuard><JobsPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/jobs/:jobId',
            element: <ClusterRouteGuard><JobsPage/></ClusterRouteGuard>,
          },
          {
            path: ':clusterName/queries',
            element: <ClusterRouteGuard><QueriesPage/></ClusterRouteGuard>,
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
