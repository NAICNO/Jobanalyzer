interface Cluster {
  cluster: string;
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
