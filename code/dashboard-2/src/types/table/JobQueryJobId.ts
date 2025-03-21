export interface JobQueryJobId {
  jobId: string;
  clusterName: string;
  hostname: string;
  user: string;
  from?: string;
  to?: string;
}
