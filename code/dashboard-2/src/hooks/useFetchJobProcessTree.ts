import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QUERY_API_ENDPOINT, QueryKeys } from '../Constants.ts'
import { generateProcessTree } from '../util/TreeUtils.ts'

const fetchJobProcessTree = async (axios: AxiosInstance, clusterName: string, jobId: string) => {
  const endpoint = '/sample'
  const query = `cluster=${clusterName}&job=${jobId}&fmt=csv,pid,ppid,cmd`
  const url = `${endpoint}?${query}`

  const response = await axios.get(url)
  return response.data
}

export const useFetchJobProcessTree = (clusterName: string, jobId: string) => {
  const axios = useAxios(QUERY_API_ENDPOINT)
  return useQuery(
    {
      gcTime: 0,
      queryKey: [QueryKeys.JOB_PROCESS_TREE, clusterName, jobId],
      queryFn: () => fetchJobProcessTree(axios, clusterName, jobId),
      select: csv => generateProcessTree(csv)
    }
  )
}
