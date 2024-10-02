import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { JOB_QUERY_API_ENDPOINT, QueryKeys } from '../Constants.ts'
import {
  FetchedJobQueryResultItem,
  JobQueryJobId,
  JobQueryResultsTableItem,
  JobQueryValues,
} from '../types'
import { parseDateString } from '../util'
import { prepareJobQueryString } from '../util/query/QueryUtils.ts'

const fetchJobQuery = async (axios: AxiosInstance, jobQueryValues: JobQueryValues, fields: string[], format?: string) => {
  const endpoint = '/jobs'
  const query = prepareJobQueryString(jobQueryValues, fields, format)
  const url = `${endpoint}?${query}`

  const response = await axios.get<FetchedJobQueryResultItem[]>(url)
  return response.data
}

export const useFetchJobQuery = (jobQueryValues: JobQueryValues, fields: string[]) => {
  const axios = useAxios(JOB_QUERY_API_ENDPOINT)
  return useQuery<FetchedJobQueryResultItem[], Error, JobQueryResultsTableItem[]>(
    {
      enabled: false,
      gcTime: 0,
      queryKey: [QueryKeys.JOB_QUERY, jobQueryValues],
      queryFn: () => fetchJobQuery(axios, jobQueryValues, fields),
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
            ...fetchedItem,
            job: job,
            start: parseDateString(fetchedItem.start),
            end: parseDateString(fetchedItem.end),
          }
        })
        //sort by end date descending
        converted.sort((a, b) => b.end.getTime() - a.end.getTime())
        return converted
      }
    }
  )
}
