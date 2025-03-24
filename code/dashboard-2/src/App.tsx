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
import JobQueryPage from './pages/JobQueryPage.tsx'
import SubclusterPage from './pages/SubclusterPage.tsx'
import JobProfilePage from './pages/JobProfilePage.tsx'

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route element={<RootLayout/>}>
      <Route index element={<Navigate to="dashboard/ml" replace/>}/>
      <Route path="dashboard" element={<DashboardPage/>}/>
      <Route path="dashboard/:clusterName" element={<DashboardPage/>}/>
      <Route path="dashboard/help/node-selection" element={<NodeSelectionHelpPage/>}/>
      <Route path=":clusterName/subcluster/:subclusterName" element={<SubclusterPage/>}/>
      <Route path=":clusterName/violators" element={<ViolatorsPage/>}/>
      <Route path=":clusterName/violators/:violator" element={<ViolatorPage/>}/>
      <Route path=":clusterName/deadweight" element={<DeadWeightPage/>}/>
      <Route path=":clusterName/:hostname" element={<HostDetailsPage/>}/>
      <Route path=":clusterName/:hostname/violators" element={<ViolatorsPage/>}/>
      <Route path=":clusterName/:hostname/:violator" element={<ViolatorPage/>}/>
      <Route path=":clusterName/:hostname/deadweight" element={<DeadWeightPage/>}/>
      <Route path="jobquery" element={<JobQueryPage/>}/>
      <Route path="jobprofile" element={<JobProfilePage/>}/>
    </Route>
  )
)

function App() {
  return (
    <RouterProvider router={router}/>
  )
}

export default App
