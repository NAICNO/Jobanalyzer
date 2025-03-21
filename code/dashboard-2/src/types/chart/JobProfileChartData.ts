import { ChartSeriesConfig } from './ChartSeriesConfig.ts'
import { ProfileInfo } from '../ProfileInfo.ts'

export interface JobProfileDataItem {
  time: string;
  [dataKey: string]: number;
}

export interface JobProfileChartData {
  dataItems: JobProfileDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  profileInfo: ProfileInfo;
}
