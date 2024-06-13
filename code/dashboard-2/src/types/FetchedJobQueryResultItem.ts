interface FetchedJobQueryResultItem {
  job: string;
  user: string;
  host: string;
  duration: string;
  start: string;
  end: string;
  'cpu-peak': number;
  'res-peak': number;
  'mem-peak': number;
  'gpu-peak': number;
  'gpumem-peak': number;
  cmd: string;
}
