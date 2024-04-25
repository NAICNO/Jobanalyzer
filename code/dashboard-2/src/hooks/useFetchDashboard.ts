import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'

const fetchDashboard = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-at-a-glance.json`
  const response = await axios.get(endpoint)
  return response.data
}

export const useFetchDashboard = (clusterName: string) => {
  const axios = useAxios()
  return useQuery<DashboardTableItem[]>(
    {
      queryKey: [QueryKeys.DASHBOARD_TABLE, clusterName],
      queryFn: () => fetchDashboard(axios, clusterName),
    }
  )
}
