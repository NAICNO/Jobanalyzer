import axios, { AxiosInstance } from 'axios'

import { API_ENDPOINT } from '../Constants.ts'

export function getAxiosInstance(endpoint? : string): AxiosInstance {
  return axios.create({
    baseURL: endpoint || API_ENDPOINT,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
  })
}
