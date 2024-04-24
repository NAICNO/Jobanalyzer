import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'

const fetchViolations = async (axios: AxiosInstance, clusterName: string) => {
  const endpoint = `/${clusterName}-violator-report.json`
  const response: AxiosResponse<ViolatingJob[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchViolations = (clusterName: string) => {
  const axios = useAxios()
  return useQuery(
    {
      queryKey: [QueryKeys.VIOLATIONS, clusterName],
      queryFn: () => fetchViolations(axios, clusterName),
      select: (data) => {

        const users: Record<string, ViolatingUser> = {}

        for (const violatingJob of data) {
          let u = users[violatingJob.user]
          if (u) {
            u.count++
            u.earliest = u.earliest < violatingJob['started-on-or-before'] ? u.earliest : violatingJob['started-on-or-before']
            u.latest = u.latest > violatingJob['last-seen'] ? u.latest : violatingJob['last-seen']
            // u.jobs.push(violatingJob)
          } else {
            users[violatingJob.user] = {
              user: {text: violatingJob.user, link: `${violatingJob.user}`},
              count: 1,
              earliest: violatingJob['started-on-or-before'],
              latest: violatingJob['last-seen'],
            }
          }
        }

        const byUser = Object.values(users)
          .sort((a, b) => a.user.text.localeCompare(b.user.text))

        data.sort((a, b) => a['last-seen'] > b['last-seen'] ? -1 : 1)
        return {
          byUser,
          byJob: data
        }
      }
    }
  )
}
