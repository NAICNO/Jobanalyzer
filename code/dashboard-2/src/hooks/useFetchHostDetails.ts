import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'
import moment from 'moment'

import useAxios from './useAxios.ts'
import { CHART_SERIES_CONFIGS, QueryKeys } from '../Constants.ts'
import { reformatHostDescriptions } from '../util'
import {
  ChartDataItem,
  ChartSeriesConfig,
  HostDetails,
  HostFetchedData,
} from '../types'

const fetchHostDetails = async (axios: AxiosInstance, hostname: string, frequency: string) => {
  const endpoint = `/${hostname}-${frequency}.json`
  const response = await axios.get<HostFetchedData>(endpoint)
  return response.data
}

export const useFetchHostDetails = (
  hostname: string,
  frequency: string,
  isShowData: boolean = true,
  isShowDowntime: boolean = false,
  enabled: boolean = true
) => {
  const axios = useAxios()
  return useQuery<HostFetchedData, Error, HostDetails>(
    {
      enabled,
      queryKey: [QueryKeys.HOSTNAME, hostname, frequency],
      queryFn: () => fetchHostDetails(axios, hostname, frequency),
      select: (hostFetchedData) => {

        const {
          labels: timestamps,
          rcpu,
          rmem,
          rres,
          rgpu,
          rgpumem,
          downhost,
          downgpu
        } = hostFetchedData

        // Clamp GPU data to get rid of occasional garbage, it's probably OK to do this even
        // if it's not ideal.
        const processedGpuData = rgpu ? rgpu.map(d => Math.min(d, 100)) : null

        // Downtime data are flags indicating that the host or gpu was down during specific periods -
        // during the hour / day starting with at the start time of the bucket.  To represent that in
        // the current plot, we carry each nonzero value forward to the next slot too, to get a
        // horizontal line covering the entire bucket.  To make that pretty, we delete the remaining
        // zero slots.


        const processedDownhostData = processDowntimeData(downhost, 15)
        const processedDowngpuData = processDowntimeData(downgpu, 30)


        // Assemble the chart data.
        const seriesNames = new Set<string>()

        const chartData: ChartDataItem[] = timestamps.map((timestamp, i) => {
          let dataEntry: Partial<ChartDataItem> = {timestamp: moment(timestamp).valueOf()}

          if (isShowData) {

            seriesNames.add('rcpu').add('rmem')

            dataEntry = {
              ...dataEntry,
              rcpu: rcpu[i],
              rmem: rmem[i],
            }
            if (rres) {
              dataEntry = {
                ...dataEntry,
                rres: rres[i]
              }
              seriesNames.add('rres')
            }
            if (rgpu) {
              dataEntry = {
                ...dataEntry,
                rgpu: processedGpuData && processedGpuData[i],
              }
              seriesNames.add('rgpu')
            }
            if (rgpumem) {
              dataEntry = {
                ...dataEntry,
                rgpumem: rgpumem[i]
              }
              seriesNames.add('rgpumem')
            }
          }

          if (isShowDowntime) {
            if (processedDownhostData) {
              dataEntry = {
                ...dataEntry,
                downhost: processedDownhostData[i]
              }
              seriesNames.add('downhost')
            }

            if (processedDowngpuData) {
              dataEntry = {
                ...dataEntry,
                downgpu: processedDowngpuData[i]
              }
              seriesNames.add('downgpu')
            }
          }

          return dataEntry as ChartDataItem
        })

        const seriesConfigs: ChartSeriesConfig[] = Array.from(seriesNames).map((name) => {
          return CHART_SERIES_CONFIGS[name]
        })


        return {
          system: {
            hostname: hostFetchedData.system.hostname,
            description: reformatHostDescriptions(hostFetchedData.system.description),
          },
          chart: {
            dataItems: chartData,
            seriesConfigs: seriesConfigs,
          }
        }
      }
    }
  )
}

const processDowntimeData = (data: (0 | 1)[] | undefined, multiplier: number) => {
  if (!data) return null
  const processedData = data.map(d => d * multiplier)
  for (let i = processedData.length - 1; i > 0; i--) {
    if (processedData[i - 1] > 0) {
      processedData[i] = processedData[i - 1]
    }
  }
  return processedData.filter(d => d !== 0)
}
