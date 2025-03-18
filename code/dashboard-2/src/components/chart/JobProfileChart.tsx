import { useEffect, useState } from 'react'
import {
  Area,
  AreaChart,
  Brush,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from 'recharts'
import { Heading, VStack } from '@chakra-ui/react'
import { DataKey } from 'recharts/types/util/types'
import moment from 'moment/moment'

import { ChartSeriesConfig, JobProfileDataItem } from '../../types'

interface JobProfileChartProps {
  profileName: string;
  dataItems: JobProfileDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  syncId?: string;
}

const dateTimeFormatter = (datetime: number) => {
  return moment(datetime).format('MMM D, HH:mm')
}

export const JobProfileChart = ({
  profileName,
  dataItems,
  seriesConfigs,
  syncId,
}: JobProfileChartProps) => {

  const [lineVisibility, setLineVisibility] = useState<boolean[]>(() => seriesConfigs.map(() => true))

  const handleLegendClick = (dataKey?: DataKey<string>) => {
    const index = seriesConfigs.findIndex((config) => config.dataKey === dataKey)
    setLineVisibility((prev) => {
      const newVisibility = [...prev]
      newVisibility[index] = !newVisibility[index]
      return newVisibility
    })
  }

  useEffect(() => {
    // To ensure lineVisibility array matches seriesConfigs length
    if (lineVisibility.length !== seriesConfigs.length) {
      setLineVisibility(seriesConfigs.map(() => true))
    }
  }, [seriesConfigs])

  if (!dataItems.length || !seriesConfigs.length) {
    return
  }

  return (
    <VStack spacing={2} width={'100%'}>
      <Heading size={'h4'}>{`${profileName} Profile`}</Heading>
      <ResponsiveContainer width="100%" height={600}>
        <AreaChart
          title={profileName}
          data={dataItems}
          margin={{top: 20, right: 20, left: 20, bottom: 40}}
          syncId={syncId}
        >
          <CartesianGrid strokeDasharray="3 3"/>
          <XAxis
            dataKey="time"
            angle={-45}
            textAnchor="end"
            height={120}
            interval={'equidistantPreserveStart'}
            tickFormatter={dateTimeFormatter}
          />
          <YAxis/>
          <Tooltip labelFormatter={dateTimeFormatter}/>
          <Legend
            verticalAlign={'top'}
            onClick={(payload) => {
              handleLegendClick(payload.dataKey)
            }}
          />
          {seriesConfigs.map((config, index) => (
            <Area
              key={config.dataKey}
              type="linear"
              dataKey={config.dataKey}
              stroke={config.lineColor}
              fill={config.lineColor}
              stackId="1"
              hide={!lineVisibility[index]}
            />
          ))}
          <Brush
            dataKey="time"
            height={40}
            tickFormatter={dateTimeFormatter}
            style={{overflow: 'visible'}}
          >
            <AreaChart data={dataItems}>
              {seriesConfigs.map((config, index) => (
                <Area
                  key={config.dataKey}
                  type="monotone"
                  dataKey={config.dataKey}
                  stroke={config.lineColor}
                  fill={config.lineColor}
                  hide={!lineVisibility[index]}
                />
              ))}
            </AreaChart>
          </Brush>
        </AreaChart>
      </ResponsiveContainer>
    </VStack>
  )
}
