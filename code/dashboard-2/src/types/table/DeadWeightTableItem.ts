interface DeadWeightTableItem {
  hostname: string;
  user: string;
  id: string;
  cmd: string;
  'started-on-or-before': Date;
  'first-violation': Date;
  'last-seen': Date;
}
