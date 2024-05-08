import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'

interface Filter {
  afterDate: Date | null
  hostname: string | null
}

const fetchDeadWeight = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-deadweight-report.json`
  const response: AxiosResponse<DeadWeight[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchDeadWeight = (clusterName: string, filter: Filter | null = null, enabled: boolean = true) => {
  const axios = useAxios()
  return useQuery(
    {
      enabled,
      queryKey: [QueryKeys.DEAD_WEIGHT, clusterName],
      queryFn: () => fetchDeadWeight(axios, clusterName),
      select: data => {
        if (filter) {
          if (filter.hostname) {
            data = data.filter(d => d.hostname === filter.hostname)
          }
          if (filter.afterDate) {
            data = data.filter(d => new Date(d['last-seen']) > filter.afterDate!)
          }
        }
        return data
      }
    }
  )
}
