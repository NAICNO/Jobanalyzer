interface ViolatingJobTableItem {
  hostname: string;
  user: string;
  id: string;
  cmd: string;
  'started-on-or-before': Date;
  'last-seen': Date;
  'cpu-peak': number;
  'rcpu-avg': number;
  'rcpu-peak': number;
  'rmem-avg': number;
  'rmem-peak': number;
  policyName?: string;
}
