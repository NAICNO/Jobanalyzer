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

const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout/>,
    children: [
      {
        index: true,
        element: <Navigate to="dashboard/ml" replace/>
      },
      {
        path: 'v2',
        children: [
          {
            path: ':clusterName/overview',
            element: <ClusterOverview />,
          },
          {
            path: ':clusterName/nodes',
            element: <NodesPage/>,
          },
          {
            path: ':clusterName/nodes/:nodename',
            element: <NodesPage/>,
          },
          {
            path: ':clusterName/nodes/:nodename/topology',
            element: <NodeTopologyPage/>,
          },
          {
            path: ':clusterName/partitions',
            element: <PartitionsPage/>,
          },
          {
            path: ':clusterName/partitions/:partitionName',
            element: <PartitionsPage/>,
          },
          {
            path: ':clusterName/jobs',
            element: <JobsPage/>,
          },
          {
            path: ':clusterName/jobs/:jobId',
            element: <JobsPage/>,
          },
          {
            path: ':clusterName/queries',
            element: <QueriesPage/>,
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
