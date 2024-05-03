import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'

const fetchHostnames = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-hostnames.json`
  const response = await axios.get(endpoint)
  return response.data
}

export const useFetchHostnames = (clusterName: string) => {
  const axios = useAxios()
  return useQuery<string[]>(
    {
      queryKey: [QueryKeys.HOSTNAME_LIST, clusterName],
      queryFn: () => fetchHostnames(axios, clusterName),
    }
  )
}
