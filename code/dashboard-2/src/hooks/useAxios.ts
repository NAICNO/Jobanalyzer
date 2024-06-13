import { getAxiosInstance } from '../api/ApiUtils.ts'

const useAxios = (endpoint?: string) => {
  return getAxiosInstance(endpoint)
}

export default useAxios
