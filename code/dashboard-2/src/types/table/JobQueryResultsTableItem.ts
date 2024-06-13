interface JobQueryResultsTableItem {
  job: JobQueryJobId;
  user: string;
  host: string;
  duration: string;
  start: Date;
  end: Date;
  cpuPeak: number;
  resPeak: number;
  memPeak: number;
  gpuPeak: number;
  gpumemPeak: number;
  cmd: string;
}
