import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'
import { Cluster } from '../types'

const fetchHostnames = async (axios: AxiosInstance, canonical: string, clusterName: string) => {
  const endpoint = `${canonical}/${clusterName}-hostnames.json`
  const response = await axios.get(endpoint)
  return response.data
}

export const useFetchHostnames = (cluster: Cluster) => {
  const axios = useAxios()
  return useQuery<string[]>(
    {
      queryKey: [QueryKeys.HOSTNAME_LIST, cluster.cluster],
      queryFn: () => fetchHostnames(axios, cluster.canonical, cluster.cluster),
    }
  )
}
