import { HostDetailsChartDataItem, ChartSeriesConfig } from './chart'

export interface HostDetails {
  chart: {
    dataItems: HostDetailsChartDataItem[];
    seriesConfigs: ChartSeriesConfig[];
  }
  system: {
    hostname: string;
    description: string;
  }
}
