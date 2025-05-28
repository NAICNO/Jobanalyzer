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
import {
  Checkbox,
  HStack,
  Heading,
  Portal,
  Select,
  Spacer,
  Text,
  VStack,
  createListCollection,
} from '@chakra-ui/react'
import { DataKey } from 'recharts/types/util/types'
import moment from 'moment/moment'

import { ChartSeriesConfig, JobProfileDataItem, ProfileInfo } from '../../types'

interface JobProfileChartProps {
  profileInfo: ProfileInfo;
  dataItems: JobProfileDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  syncId?: string;
}

const dateTimeFormatter = (datetime: number) => {
  return moment(datetime).format('MMM D, HH:mm')
}

export const JobProfileChart = ({
  profileInfo,
  dataItems,
  seriesConfigs,
  syncId,
}: JobProfileChartProps) => {

  const chartTypeOptions = createListCollection({
    items: [
      {value: 'area', label: 'Area'},
      {value: 'line', label: 'Line'}
    ]
  })
  const [lineVisibility, setLineVisibility] = useState<boolean[]>(() => seriesConfigs.map(() => true))
  const [dotsVisibility, setDotsVisibility] = useState<boolean>(true)
  const [chartType, setChartType] = useState<string[]>(['area'])

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
    <VStack gap={2} width={'100%'} alignItems={'start'}>
      <HStack width="100%" paddingLeft={8} paddingRight={4}>
        <Heading size={'lg'}>{profileInfo.text}</Heading>
        <Spacer/>
        <Text>Chart Type: </Text>
        <Select.Root
          collection={chartTypeOptions}
          value={chartType}
          onValueChange={(e) => setChartType(e.value)}
          width="120px"
          size={'sm'}
        >
          <Select.HiddenSelect/>
          <Select.Control>
            <Select.Trigger>
              <Select.ValueText/>
            </Select.Trigger>
            <Select.IndicatorGroup>
              <Select.Indicator/>
            </Select.IndicatorGroup>
          </Select.Control>
          <Portal>
            <Select.Positioner>
              <Select.Content>
                <Select.Item item="area">
                  <Select.ItemText>Area</Select.ItemText>
                </Select.Item>
                <Select.Item item="line">
                  <Select.ItemText>Line</Select.ItemText>
                </Select.Item>
              </Select.Content>
            </Select.Positioner>
          </Portal>
        </Select.Root>
        <Checkbox.Root
          colorPalette={'blue'}
          variant={'subtle'}
          checked={dotsVisibility}
          onCheckedChange={(e) => setDotsVisibility(!!e.checked)}
        >
          <Checkbox.HiddenInput/>
          <Checkbox.Control/>
          <Checkbox.Label>
            Show Dots
          </Checkbox.Label>
        </Checkbox.Root>
      </HStack>
      {chartType[0] === 'area' ? (
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
            <YAxis label={{value: profileInfo.yAxisLabel, angle: -90, position: 'insideLeft'}}/>
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
            <YAxis label={{value: profileInfo.yAxisLabel, angle: -90, position: 'insideLeft'}}/>
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
                    type="monotone"
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
