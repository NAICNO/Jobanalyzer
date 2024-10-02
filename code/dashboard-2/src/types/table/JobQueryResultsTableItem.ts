import { JobQueryJobId } from './JobQueryJobId.ts'

export interface JobQueryResultsTableItem {
  job: JobQueryJobId;
  user: string;
  host: string;
  duration: string;
  start: Date;
  end: Date;
  'cpu-peak': number;
  'res-peak': number;
  'mem-peak': number;
  'gpu-peak': number;
  'gpumem-peak': number;
  cmd: string;
}
