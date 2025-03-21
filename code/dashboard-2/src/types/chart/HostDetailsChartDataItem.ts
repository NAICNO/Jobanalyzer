export interface HostDetailsChartDataItem {
  timestamp: number;
  rcpu: number;
  rmem: number;
  rres: number;
  rgpu: number | null;
  rgpumem: number | null;
  downhost?: number;
  downgpu?: number;
}
