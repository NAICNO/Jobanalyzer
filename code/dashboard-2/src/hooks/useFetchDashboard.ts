import { useQuery } from '@tanstack/react-query'
import { AxiosInstance } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'
import { makeFilter } from '../util/query/QueryUtils.ts'
import { sortDashboardTableRows } from '../util/TableUtils.ts'
import { Cluster } from '../types/Cluster.ts'

const fetchDashboard = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-at-a-glance.json`
  const response = await axios.get(endpoint)
  return response.data
}

export const useFetchDashboard = (cluster: Cluster, query: string) => {
  const axios = useAxios()
  const clusterName = cluster.cluster
  return useQuery(
    {
      queryKey: [QueryKeys.DASHBOARD_TABLE, clusterName],
      queryFn: () => fetchDashboard(axios, clusterName),
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
