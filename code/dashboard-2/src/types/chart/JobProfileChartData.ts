import { ChartSeriesConfig } from './ChartSeriesConfig.ts'

export interface JobProfileDataItem {
  time: string;
  [dataKey: string]: number;
}

export interface JobProfileChartData {
  dataItems: JobProfileDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  profileName: string;
}
