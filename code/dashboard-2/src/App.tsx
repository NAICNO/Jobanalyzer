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
        element: <DashboardPage/>
      },
      {
        path: 'dashboard/:clusterName',
        element: <DashboardPage/>
      },
      {
        path: 'dashboard/help/node-selection',
        element: <NodeSelectionHelpPage/>
      },
      {
        path: ':clusterName/subcluster/:subclusterName',
        element: <SubclusterPage/>
      },
      {
        path: ':clusterName/violators',
        element: <ViolatorsPage/>
      },
      {
        path: ':clusterName/violators/:violator',
        element: <ViolatorPage/>
      },
      {
        path: ':clusterName/deadweight',
        element: <DeadWeightPage/>
      },
      {
        path: ':clusterName/:hostname',
        element: <HostDetailsPage/>
      },
      {
        path: ':clusterName/:hostname/violators',
        element: <ViolatorsPage/>
      },
      {
        path: ':clusterName/:hostname/:violator',
        element: <ViolatorPage/>
      },
      {
        path: ':clusterName/:hostname/deadweight',
        element: <DeadWeightPage/>
      },
      {
        path: 'jobquery',
        element: <JobQueryPage/>
      },
      {
        path: 'jobprofile',
        element: <JobProfilePage/>
      }
    ]
  }
])

function App() {
  return (
    <RouterProvider router={router}/>
  )
}

export default App
