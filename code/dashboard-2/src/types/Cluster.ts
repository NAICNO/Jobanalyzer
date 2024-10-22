import { Subcluster } from './SubCluster.ts'

export interface Cluster {
  cluster: string;
  canonical: string;
  subclusters: Subcluster[];
  uptime: boolean;
  violators: boolean;
  deadweight: boolean;
  defaultQuery: string;
  hasDowntime: boolean;
  name: string;
  description: string;
  prefix: string;
  policy: string;
}
