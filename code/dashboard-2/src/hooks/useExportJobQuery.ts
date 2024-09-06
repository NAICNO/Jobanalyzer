import JobQueryValues from '../types/JobQueryValues.ts'
import useAxios from './useAxios.ts'
import { JOB_QUERY_API_ENDPOINT, QueryKeys } from '../Constants.ts'
import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'
import { prepareJobQueryString } from '../util/query/QueryUtils.ts'

const exportJobQuery = async (axios: AxiosInstance, jobQueryValues: JobQueryValues, fields: string[], format?: string) => {
  const endpoint = '/jobs'
  const query = prepareJobQueryString(jobQueryValues, fields, format)
  const url = `${endpoint}?${query}`

  const response = await axios.get(url, {responseType: 'blob'})
  return response.data
}

export const useExportJobQuery = (jobQueryValues: JobQueryValues, exportOptions: ExportOptions) => {
  const axios = useAxios(JOB_QUERY_API_ENDPOINT)
  return useQuery(
    {
      enabled: false,
      gcTime: 0,
      retryOnMount: false,
      queryKey: [QueryKeys.EXPORT_JOB_QUERY],
      queryFn: () => exportJobQuery(axios, jobQueryValues, exportOptions.fields, exportOptions.format),
    }
  )
}
