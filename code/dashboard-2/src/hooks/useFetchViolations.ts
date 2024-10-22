import { useQuery } from '@tanstack/react-query'
import { AxiosInstance, AxiosResponse } from 'axios'

import useAxios from './useAxios.ts'
import { QueryKeys } from '../Constants.ts'
import {
  Cluster,
  FetchedViolatingJob,
  JobQueryValues,
  ViolatingJob,
  ViolatingUser
} from '../types'
import { prepareShareableJobQueryLink } from '../util/query/QueryUtils.ts'

interface Filter {
  afterDate: Date | null
  hostname: string | null
}

const fetchViolations = async (axios: AxiosInstance, canonical: string, clusterName: string) => {
  const endpoint = `${canonical}/${clusterName}-violator-report.json`
  const response: AxiosResponse<FetchedViolatingJob[]> = await axios.get(endpoint)
  return response.data
}

export const useFetchViolations = (cluster: Cluster, filter: Filter | null = null, enabled: boolean = true) => {
  const axios = useAxios()
  return useQuery(
    {
      enabled,
      queryKey: [QueryKeys.VIOLATIONS, cluster.cluster],
      queryFn: () => fetchViolations(axios, cluster.canonical, cluster.cluster),
      select: (data) => {

        let violatingJobs: ViolatingJob[] = data.map((d) => {
          const jobQueryValues: JobQueryValues = {
            gpuUsage: '',
            minPeakCpuCores: null,
            minPeakResidentGb: null,
            minRuntime: null,
            nodeNames: d.hostname,
            clusterName: cluster.cluster,
            jobIds: d.id.toString(),
            usernames: d.user,
            fromDate: '',
            toDate: '',
          }
          const shareableLink = prepareShareableJobQueryLink(jobQueryValues)
          return {
            ...d,
            user: {text: d.user, link: shareableLink, openInNewTab: true},
          }
        })

        if (filter) {
          if (filter.hostname) {
            violatingJobs = violatingJobs.filter((d) => d.hostname === filter.hostname)
          }
          if (filter.afterDate) {
            violatingJobs = violatingJobs.filter((d) => new Date(d['last-seen']) > filter.afterDate!)
          }
        }

        const users: Record<string, ViolatingUser> = {}

        for (const violatingJob of violatingJobs) {
          const username = violatingJob.user.text
          const u = users[username]
          if (u) {
            u.count++
            u.earliest = u.earliest < violatingJob['started-on-or-before'] ? u.earliest : violatingJob['started-on-or-before']
            u.latest = u.latest > violatingJob['last-seen'] ? u.latest : violatingJob['last-seen']
          } else {
            users[username] = {
              user: {text: username, link: username},
              count: 1,
              earliest: violatingJob['started-on-or-before'],
              latest: violatingJob['last-seen'],
            }
          }
        }

        const byUser = Object.values(users)
          .sort((a, b) => a.user.text.localeCompare(b.user.text))

        violatingJobs.sort((a, b) => a['last-seen'] > b['last-seen'] ? -1 : 1)
        return {
          byUser,
          byJob: violatingJobs
        }
      }
    }
  )
}
