import {
  createBrowserRouter,
  createRoutesFromElements,
  Navigate,
  Route,
  RouterProvider
} from 'react-router-dom'

import RootLayout from './layouts/RootLayout.tsx'
import MlNodesHomePage from './pages/MlNodesHomePage.tsx'
import FoxHomePage from './pages/FoxHomePage.tsx'
import SagaHomePage from './pages/SagaHomePage.tsx'

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route element={<RootLayout/>}>
      <Route index element={<Navigate to="ml" replace/>}/>
      <Route path="ml" element={<MlNodesHomePage/>}/>
      <Route path="fox" element={<FoxHomePage/>}/>
      <Route path="saga" element={<SagaHomePage/>}/>
    </Route>
  )
)

function App() {
  return (
    <RouterProvider router={router}/>
  )
}

export default App
