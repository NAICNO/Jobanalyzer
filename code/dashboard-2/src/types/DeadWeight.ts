import { TextWithLink } from './TextWithLink.ts'

export interface DeadWeight {
  hostname: string;
  id: TextWithLink;
  user: TextWithLink;
  cmd: string;
  'started-on-or-before': Date;
  'first-violation': Date;
  'last-seen': Date;
}
