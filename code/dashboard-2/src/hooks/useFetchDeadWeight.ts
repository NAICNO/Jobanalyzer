import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'
import JobQueryValues from '../types/JobQueryValues.ts'
import { prepareShareableJobQueryLink } from '../util/query/QueryUtils.ts'

interface Filter {
  afterDate: Date | null
  hostname: string | null
}

const fetchDeadWeight = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-deadweight-report.json`
  const response: AxiosResponse<FetchedDeadWeight[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchDeadWeight = (clusterName: string, filter: Filter | null = null, enabled: boolean = true) => {
  const axios = useAxios()
  return useQuery<FetchedDeadWeight[], Error, DeadWeight[]>(
    {
      enabled,
      queryKey: [QueryKeys.DEAD_WEIGHT, clusterName],
      queryFn: () => fetchDeadWeight(axios, clusterName),
      select: data => {
        let deadWeights: DeadWeight[] = data.map((d: FetchedDeadWeight) => {
          const commonJobQueryValues = {
            nodeNames: d.hostname,
            clusterName: clusterName,
            gpuUsage: '',
            minPeakCpuCores: null,
            minPeakResidentGb: null,
            minRuntime: null,
            fromDate: '',
            toDate: '',
          }

          const jobQueryValuesWithJobId: JobQueryValues = {
            ...commonJobQueryValues,
            usernames: '',
            jobIds: d.id.toString(),
          }

          const jobQueryValuesWithUser: JobQueryValues = {
            ...commonJobQueryValues,
            usernames: d.user,
            jobIds: '',
          }

          const shareableLinkWithJobId = prepareShareableJobQueryLink(jobQueryValuesWithJobId)
          const shareableLinkWithUser = prepareShareableJobQueryLink(jobQueryValuesWithUser)

          return {
            ...d,
            id: {text: d.id, link: shareableLinkWithJobId, openInNewTab: true},
            user: {text: d.user, link: shareableLinkWithUser, openInNewTab: true},
          }
        })

        if (filter) {
          if (filter.hostname) {
            deadWeights = deadWeights.filter(d => d.hostname === filter.hostname)
          }
          if (filter.afterDate) {
            deadWeights = deadWeights.filter(d => new Date(d['last-seen']) > filter.afterDate!)
          }
        }
        return deadWeights
      }
    }
  )
}
