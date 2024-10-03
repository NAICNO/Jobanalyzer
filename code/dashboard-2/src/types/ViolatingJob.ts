import { TextWithLink } from './TextWithLink.ts'

export interface ViolatingJob {
  hostname: string;
  id: number;
  user: TextWithLink;
  cmd: string;
  'started-on-or-before': Date;
  'first-violation': string;
  'last-seen': Date;
  'cpu-peak': number;
  'rcpu-avg': number;
  'rcpu-peak': number;
  'rmem-avg': number;
  'rmem-peak': number;
  policyName?: string;
}
