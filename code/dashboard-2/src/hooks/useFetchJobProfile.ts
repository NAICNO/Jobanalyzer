import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { PROFILING_INFO, QUERY_API_ENDPOINT, QueryKeys } from '../Constants.ts'
import { ChartSeriesConfig, FetchedJobProfileResultItem, JobProfileChartData, JobProfileDataItem, } from '../types'
import { generateTolRainbowColors } from '../util'

const fetchJobProfile = async (axios: AxiosInstance, clusterName: string, hostname: string, jobId: string, from: string, to: string) => {
  const endpoint = '/profile'
  const query = `cluster=${clusterName}&job=${jobId}&host=${hostname}&from=${from}&to=${to}&fmt=json`
  const url = `${endpoint}?${query}`

  const response = await axios.get<FetchedJobProfileResultItem[]>(url)
  return response.data
}

export const useFetchJobProfile = (clusterName: string, hostname: string, jobId: string, from: string, to: string) => {
  const axios = useAxios(QUERY_API_ENDPOINT)
  return useQuery<FetchedJobProfileResultItem[], Error, JobProfileChartData[]>(
    {
      gcTime: 0,
      queryKey: [QueryKeys.JOB_PROFILE, clusterName, hostname, jobId],
      queryFn: () => fetchJobProfile(axios, clusterName, hostname, jobId, from, to),
      select: (data) => transformData(data),
      initialData: () => {
        return []
      }
    }
  )
}

const transformData = (
  data: FetchedJobProfileResultItem[],
) => {
  const commandPidSet = new Set<string>()
  // Precompute raw data items with the original point values
  const rawDataItems = data.map(entry => {
    const raw: { time: string, [key: string]: any } = {time: entry.time}
    entry.points.forEach(point => {
      const key = `${point.command}-${point.pid}`
      commandPidSet.add(key)
      raw[key] = point
    })
    return raw
  })

  const commandPidKeys = Array.from(commandPidSet)
  const colors = generateTolRainbowColors(commandPidKeys.length)

  return PROFILING_INFO.map((profileInfo) => {
    const dataItems: JobProfileDataItem[] = rawDataItems.map(raw => {
      const transformed: JobProfileDataItem = {time: raw.time} as JobProfileDataItem
      commandPidKeys.forEach(key => {
        transformed[key] = raw[key]
          ? Math.round(raw[key][profileInfo.key] * profileInfo.scaleFactor * 100) / 100
          : 0
      })
      return transformed
    })

    const seriesConfigs: ChartSeriesConfig[] = commandPidKeys.map((key, index) => ({
      dataKey: key,
      label: key,
      lineColor: colors[index],
    }))

    return {dataItems, seriesConfigs, profileInfo}
  })
}
