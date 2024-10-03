import { TextWithLink } from '../TextWithLink.ts'

export interface ViolatingUserTableItem {
  user: TextWithLink;
  count: number;
  earliest: Date;
  latest: Date;
}
