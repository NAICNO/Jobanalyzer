import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'

const fetchHostDetails = async (axios: AxiosInstance, hostname: string, frequency: string) => {
  const endpoint = `/${hostname}-${frequency}.json`
  const response = await axios.get(endpoint)
  return response.data
}

export const useFetchHostDetails = (hostname: string, frequency: string, enabled: boolean = true) => {
  const axios = useAxios()
  return useQuery<HostDetails>(
    {
      enabled,
      queryKey: [QueryKeys.HOSTNAME, hostname, frequency],
      queryFn: () => fetchHostDetails(axios, hostname, frequency),
    }
  )
}
