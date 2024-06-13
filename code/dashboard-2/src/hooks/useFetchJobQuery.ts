import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { JOB_QUERY_API_ENDPOINT, QueryKeys, TRUE_VAL } from '../Constants.ts'
import JobQueryValues from '../types/JobQueryValues.ts'
import { splitMultiPattern } from '../util/query/HostGlobber.ts'
import { parseDateString } from '../util'

export const prepareQueryString = (jobQueryValues?: JobQueryValues) => {
  if(!jobQueryValues) {
    return ''
  }
  let query = `cluster=${jobQueryValues.clusterName}`

  const trimmedUsernames = jobQueryValues.usernames || ''

  const userNameList = trimmedUsernames ? trimmedUsernames.split(',').map(item => item.trim()) : []
  if (userNameList?.length === 0) {
    query += `&user=-`
  } else {
    userNameList?.forEach(userName => {
      query += `&user=${userName}`
    })
  }

  const nodeNameList = jobQueryValues.nodeNames ? splitMultiPattern(jobQueryValues.nodeNames) : []
  nodeNameList.forEach(nodeName => {
    query += `&host=${nodeName}`
  })

  const jobIdList = jobQueryValues.jobIds ? jobQueryValues.jobIds.split(',').map(id => parseInt(id)) : []
  jobIdList.forEach(jobId => {
    query += `&job=${jobId}`
  })

  const fromDate = jobQueryValues.fromDate
  query += `&from=${fromDate}`

  const toDate = jobQueryValues.toDate
  query += `&to=${toDate}`

  const minRuntime = jobQueryValues.minRuntime
  if (minRuntime) {
    query += `&min-runtime=${minRuntime}`
  }

  const minPeakCpuCores = jobQueryValues.minPeakCpuCores
  if (minPeakCpuCores) {
    query += `&min-cpu-peak=${minPeakCpuCores * 100}`
  }

  const minPeakResidentGb = jobQueryValues.minPeakResidentGb
  if (minPeakResidentGb) {
    query += `&min-res-peak=${minPeakResidentGb}`
  }

  const gpuUsage = jobQueryValues.gpuUsage
  if (gpuUsage !== 'either') {
    query += `&${gpuUsage}=${TRUE_VAL}`
  }

  const fmt = 'fmt=json,job,user,host,duration,start,end,cpu-peak,res-peak,mem-peak,gpu-peak,gpumem-peak,cmd'
  query += `&${fmt}`

  return query
}

const fetchFetchJobQuery = async (axios: AxiosInstance, jobQueryValues: JobQueryValues) => {
  const endpoint = '/jobs'
  const query = prepareQueryString(jobQueryValues)
  const url = `${endpoint}?${query}`

  const response = await axios.get<FetchedJobQueryResultItem[]>(url)
  return response.data
}


export const useFetchJobQuery = (jobQueryValues: JobQueryValues) => {
  const axios = useAxios(JOB_QUERY_API_ENDPOINT)
  return useQuery<FetchedJobQueryResultItem[], Error, JobQueryResultsTableItem[]>(
    {
      enabled: false,
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
