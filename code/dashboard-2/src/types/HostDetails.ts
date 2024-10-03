import { ChartDataItem, ChartSeriesConfig } from './chart'

export interface HostDetails {
  chart: {
    dataItems: ChartDataItem[];
    seriesConfigs: ChartSeriesConfig[];
  }
  system: {
    hostname: string;
    description: string;
  }
}
