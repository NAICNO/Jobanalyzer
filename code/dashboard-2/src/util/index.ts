import { CLUSTER_INFO } from '../Constants.ts'

export const isValidateClusterName = (clusterName?: string | null): boolean => {
  return clusterName ? !!CLUSTER_INFO[clusterName] : false;
}
