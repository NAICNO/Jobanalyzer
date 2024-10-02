import { TextWithLink } from './TextWithLink.ts'

export interface ViolatingUser {
  user: TextWithLink;
  count: number;
  earliest: Date;
  latest: Date;
}
