import { ChartSeriesConfig } from './ChartSeriesConfig.ts'
import { ProfileInfo } from '../ProfileInfo.ts'

export type JobProfileDataItem = {
  time: string;
} & Record<string, number>

export interface JobProfileChartData {
  dataItems: JobProfileDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  profileInfo: ProfileInfo;
}
