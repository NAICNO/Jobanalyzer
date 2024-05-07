interface ChartDataItem {
  timestamp: number;
  rcpu: number;
  rmem: number;
  rres: number;
  rgpu: number | null;
  rgpumem: number | null;
  downhost?: 0 | 1;
  downgpu?: 0 | 1;
}
