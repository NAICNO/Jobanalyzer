import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QUERY_API_ENDPOINT, QueryKeys } from '../Constants.ts'
import {
  FetchedJobQueryResultItem,
  JobQueryResultsTableItem,
  JobQueryValues, TextWithLink,
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
  const axios = useAxios(QUERY_API_ENDPOINT)
  return useQuery<FetchedJobQueryResultItem[], Error, JobQueryResultsTableItem[]>(
    {
      enabled: false,
      gcTime: 0,
      queryKey: [QueryKeys.JOB_QUERY, jobQueryValues],
      queryFn: () => fetchJobQuery(axios, jobQueryValues, fields),
      select: data => {
        const converted = data.map((fetchedItem) => {

          const query = new URLSearchParams(
            Object.entries(
              {
                jobId: fetchedItem.job,
                clusterName: jobQueryValues.clusterName,
                hostname: fetchedItem.host,
                user: fetchedItem.user,
                from: jobQueryValues.fromDate,
                to: jobQueryValues.toDate,
              }
            ).reduce<Record<string, string>>((acc, [key, val]) => {
              acc[key] = String(val)
              return acc
            }, {})
          ).toString()

          const job: TextWithLink = {text: fetchedItem.job, link: `/jobs/profile?${query}`, openInNewTab: true}

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
