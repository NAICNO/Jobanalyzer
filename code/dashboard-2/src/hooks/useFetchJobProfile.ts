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
      select: (data) => {
        return PROFILING_INFO.map(({key, text, scaleFactor}) => {
          return transformData(data, key, text, scaleFactor)
        })
      },
      initialData: () => {
        return []
      }
    }
  )
}

const transformData = (
  data: FetchedJobProfileResultItem[],
  profileType: string,
  profileName: string,
  scaleFactor: number,
) => {
  // First pass: collect all unique keys from data points
  const commandPidSet = new Set<string>()
  data.forEach(entry => {
    entry.points.forEach(point => {
      const key = `${point.command}-${point.pid}`
      commandPidSet.add(key)
    })
  })

  // Convert the set to an array
  const commandPidKeys = Array.from(commandPidSet)

  // Second pass: transform data items using the collected keys
  const dataItems = data.map(entry => {
    const transformed: JobProfileDataItem = {time: entry.time}
    entry.points.forEach(point => {
      const key = `${point.command}-${point.pid}`
      transformed[key] = point[profileType] * scaleFactor
    })
    return transformed
  })

  // Generate colors for each key
  const colors = generateTolRainbowColors(commandPidKeys.length)

  // Build seriesConfigs from the unique keys
  const seriesConfigs: ChartSeriesConfig[] = commandPidKeys.map((key, index) => {
    return {
      dataKey: key,
      label: key,
      lineColor: colors[index],
    }
  })

  return {dataItems, seriesConfigs, profileName}
}
