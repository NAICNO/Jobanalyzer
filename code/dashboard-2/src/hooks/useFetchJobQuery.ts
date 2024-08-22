import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { JOB_QUERY_API_ENDPOINT, QueryKeys } from '../Constants.ts'
import JobQueryValues from '../types/JobQueryValues.ts'
import { parseDateString } from '../util'
import { prepareJobQueryString } from '../util/query/QueryUtils.ts'

const fetchFetchJobQuery = async (axios: AxiosInstance, jobQueryValues: JobQueryValues) => {
  const endpoint = '/jobs'
  const query = prepareJobQueryString(jobQueryValues)
  const url = `${endpoint}?${query}`

  const response = await axios.get<FetchedJobQueryResultItem[]>(url)
  return response.data
}

export const useFetchJobQuery = (jobQueryValues: JobQueryValues) => {
  const axios = useAxios(JOB_QUERY_API_ENDPOINT)
  return useQuery<FetchedJobQueryResultItem[], Error, JobQueryResultsTableItem[]>(
    {
      enabled: false,
      gcTime: 0,
      queryKey: [QueryKeys.JOB_QUERY, jobQueryValues],
      queryFn: () => fetchFetchJobQuery(axios, jobQueryValues),
      select: data => {
        const converted = data.map((fetchedItem) => {
          const job: JobQueryJobId = {
            jobId: fetchedItem.job,
            clusterName: jobQueryValues.clusterName,
            hostName: fetchedItem.host,
            from: jobQueryValues.fromDate,
            to: jobQueryValues.toDate,
          }
          return {
            job: job,
            user: fetchedItem.user,
            host: fetchedItem.host,
            duration: fetchedItem.duration,
            start: parseDateString(fetchedItem.start),
            end: parseDateString(fetchedItem.end),
            cpuPeak: fetchedItem['cpu-peak'],
            resPeak: fetchedItem['res-peak'],
            memPeak: fetchedItem['mem-peak'],
            gpuPeak: fetchedItem['gpu-peak'],
            gpumemPeak: fetchedItem['gpumem-peak'],
            cmd: fetchedItem.cmd,
          }
        })
        //sort by end date descending
        converted.sort((a, b) => b.end.getTime() - a.end.getTime())
        return converted
      }
    }
  )
}
