export interface FetchedJobProfileResultItem {
  time: string;
  job: number;
  points: ProcessPoint[];
}

interface ProcessPoint {
  command: string;
  pid: number;
  cpu: number;
  mem: number;
  res: number;
  gpu: number;
  gpumem: number;
  nproc: number;
}
