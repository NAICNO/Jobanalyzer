import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'
import { makeFilter } from '../util/query/QueryUtils.ts'
import { sortDashboardTableRows } from '../util/TableUtils.ts'
import { Cluster, DashboardTableItem } from '../types'

const fetchDashboard = async (axios: AxiosInstance, canonical: string, clusterName: string) => {
  const endpoint = `${canonical}/${clusterName}-at-a-glance.json`
  const response = await axios.get(endpoint)
  return response.data
}

export const useFetchDashboard = (cluster: Cluster, query: string) => {
  const axios = useAxios()
  const clusterName = cluster.cluster
  const canonical = cluster.canonical
  return useQuery(
    {
      queryKey: [QueryKeys.DASHBOARD_TABLE, clusterName],
      queryFn: () => fetchDashboard(axios, canonical, clusterName),
      select: data => {
        const filter = makeFilter(query)
        const filtered = data.filter(filter)
        filtered.sort((a: DashboardTableItem, b: DashboardTableItem) => sortDashboardTableRows(a, b, cluster.uptime))

        return filtered.map((item: Partial<DashboardTableItem>) => {
          return {
            ...item,
            hostname: {text: item.hostname, link: `/${clusterName}/${item.hostname}`},
          }
        })
      },
    }
  )
}
