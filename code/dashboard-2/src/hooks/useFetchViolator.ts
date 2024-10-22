import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { POLICIES, QueryKeys } from '../Constants.ts'
import { Cluster, FetchedViolatingJob } from '../types'

const fetchViolator = async (axios: AxiosInstance, canonical: string, clusterName: string) => {
  const endpoint = `${canonical}/${clusterName}-violator-report.json`
  const response: AxiosResponse<FetchedViolatingJob[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchViolator = (cluster: Cluster, violator: string) => {
  const axios = useAxios()
  return useQuery(
    {
      queryKey: [QueryKeys.VIOLATOR, cluster.cluster, violator],
      queryFn: () => fetchViolator(axios, cluster.canonical, cluster.cluster),
      select: (data) => {

        const jobsOfUser = data.filter(job => job.user === violator)

        const policies = POLICIES[cluster.cluster]

        return jobsOfUser.map(job => {
          // As of now we have only one policy
          return {...job, policyName: policies[0].name}
        })
      }
    }
  )
}
