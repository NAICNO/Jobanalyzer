import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { POLICIES, QueryKeys } from '../Constants.ts'
import { FetchedViolatingJob } from '../types'

const fetchViolator = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-violator-report.json`
  const response: AxiosResponse<FetchedViolatingJob[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchViolator = (clusterName: string, violator: string) => {
  const axios = useAxios()
  return useQuery(
    {
      queryKey: [QueryKeys.VIOLATOR, clusterName, violator],
      queryFn: () => fetchViolator(axios, clusterName),
      select: (data) => {

        const jobsOfUser = data.filter(job => job.user === violator)

        const policies = POLICIES[clusterName]

        return jobsOfUser.map(job => {
          // As of now we have only one policy
          return {...job, policyName: policies[0].name}
        })
      }
    }
  )
}
