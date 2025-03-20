import { useEffect, useState } from 'react'
import {
  Area,
  AreaChart,
  Brush,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from 'recharts'
import { Checkbox, Heading, HStack, Select, Spacer, Text, VStack } from '@chakra-ui/react'
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
  const [dotsVisibility, setDotsVisibility] = useState<boolean>(true)
  const [chartType, setChartType] = useState<'area' | 'line'>('area')

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
    <VStack spacing={2} width={'100%'} alignItems={'start'}>
      <HStack width="100%" paddingLeft={8} paddingRight={4}>
        <Heading size={'h4'}>{`${profileName} Profile`}</Heading>
        <Spacer/>
        <Text>Chart Type: </Text>
        <Select
          value={chartType}
          onChange={(e) => setChartType(e.target.value as 'area' | 'line')}
          width="120px"
          size={'sm'}
        >
          <option value="area">Area</option>
          <option value="line">Line</option>
        </Select>
        <Checkbox isChecked={dotsVisibility} onChange={(e) => setDotsVisibility(e.target.checked)}>
          Show Dots
        </Checkbox>
      </HStack>
      {chartType === 'area' ? (
        <ResponsiveContainer width="100%" height={700}>
          <AreaChart
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
              wrapperStyle={{
                paddingBottom: '20px',
                fontSize: '14px',
                maxHeight: '150px',
                overflowY: 'auto'
              }}
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
                strokeWidth={1.25}
                fill={config.lineColor}
                stackId="1"
                hide={!lineVisibility[index]}
                dot={dotsVisibility ? {stroke: config.lineColor, r: 2} : false}
                activeDot={dotsVisibility ? {stroke: config.lineColor, r: 3} : undefined}
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
      ) : (
        <ResponsiveContainer width="100%" height={600}>
          <LineChart
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
              wrapperStyle={{
                paddingBottom: '20px',
                fontSize: '14px',
                maxHeight: '150px',
                overflowY: 'auto'
              }}
              onClick={(payload) => {
                handleLegendClick(payload.dataKey)
              }}
            />
            {seriesConfigs.map((config, index) => (
              <Line
                key={config.dataKey}
                type="linear"
                dataKey={config.dataKey}
                stroke={config.lineColor}
                strokeWidth={1.25}
                hide={!lineVisibility[index]}
                dot={dotsVisibility ? {stroke: config.lineColor, r: 2} : false}
                activeDot={dotsVisibility ? {stroke: config.lineColor, r: 3} : undefined}
              />
            ))}
            <Brush
              dataKey="time"
              height={40}
              tickFormatter={dateTimeFormatter}
              style={{overflow: 'visible'}}
            >
              <LineChart data={dataItems}>
                {seriesConfigs.map((config, index) => (
                  <Line
                    dot={false}
                    key={config.dataKey}
                    type="linear"
                    dataKey={config.dataKey}
                    stroke={config.lineColor}
                    hide={!lineVisibility[index]}
                  />
                ))}
              </LineChart>
            </Brush>
          </LineChart>
        </ResponsiveContainer>
      )}
    </VStack>
  )
}
