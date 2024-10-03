export interface FetchedDeadWeight {
  hostname: string;
  id: string;
  user: string;
  cmd: string;
  'started-on-or-before': Date;
  'first-violation': Date;
  'last-seen': Date;
}
