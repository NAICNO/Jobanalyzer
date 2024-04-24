import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'

const fetchDeadWeight = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-deadweight-report.json`
  const response: AxiosResponse<DeadWeight[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchDeadWeight = (clusterName: string) => {
  const axios = useAxios()
  return useQuery(
    {
      queryKey: [QueryKeys.DEAD_WEIGHT, clusterName],
      queryFn: () => fetchDeadWeight(axios, clusterName),
    }
  )
}
