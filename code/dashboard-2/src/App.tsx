import {
  createBrowserRouter,
  createRoutesFromElements,
  Navigate,
  Route,
  RouterProvider
} from 'react-router-dom'

import RootLayout from './layouts/RootLayout.tsx'
import DashboardPage from './pages/DashboardPage.tsx'
import ViolatorsPage from './pages/ViolatorsPage.tsx'
import ViolatorPage from './pages/ViolatorPage.tsx'
import DeadWeightPage from './pages/DeadWeightPage.tsx'
import NodeSelectionHelpPage from './pages/NodeSelectionHelpPage.tsx'
import HostDetailsPage from './pages/HostDetailsPage.tsx'

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route element={<RootLayout/>}>
      <Route index element={<Navigate to="dashboard/ml" replace/>}/>
      <Route path="dashboard" element={<DashboardPage/>}/>
      <Route path="dashboard/:clusterName" element={<DashboardPage/>}/>
      <Route path="dashboard/help/node-selection" element={<NodeSelectionHelpPage/>}/>
      <Route path=":clusterName/violators" element={<ViolatorsPage/>}/>
      <Route path=":clusterName/violators/:violator" element={<ViolatorPage/>}/>
      <Route path=":clusterName/deadweight" element={<DeadWeightPage/>}/>
      <Route path=":clusterName/:hostname" element={<HostDetailsPage/>}/>
    </Route>
  )
)

function App() {
  return (
    <RouterProvider router={router}/>
  )
}

export default App
