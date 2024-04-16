import axios, { AxiosInstance } from 'axios'

import { API_ENDPOINT } from '../Constants.ts'

export function getAxiosInstance(): AxiosInstance {
  return axios.create({
    baseURL: API_ENDPOINT,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
  })
}
