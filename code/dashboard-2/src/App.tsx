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
          }
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
        path: 'jobs',
        children: [
          {
            index: true,
            path: 'query',
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
