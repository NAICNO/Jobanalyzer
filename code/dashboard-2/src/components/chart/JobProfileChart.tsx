import React, { useEffect, useState } from 'react'
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
  Box,
  Checkbox,
  HStack,
  Heading,
  SegmentGroup,
  SegmentGroupValueChangeDetails,
  Text,
  VStack,
} from '@chakra-ui/react'
import { DataKey } from 'recharts/types/util/types'
import { FaChartArea, FaChartLine } from 'react-icons/fa'

import { ChartSeriesConfig, JobProfileDataItem, ProfileInfo } from '../../types'
import { CustomTooltip } from './CustomTooltip'
import { CustomLegend } from './CustomLegend'
import { dateTimeFormatter } from '../../util'

interface JobProfileChartProps {
  profileInfo: ProfileInfo;
  dataItems: JobProfileDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  syncId?: string;
  chartSize?: { width: number | string; height: number | string };
  maxTooltipItems?: number;
  maxLegendItems?: number;
  wrapperStyles?: React.CSSProperties;
}

export const JobProfileChart = ({
  profileInfo,
  dataItems,
  seriesConfigs,
  syncId,
  chartSize,
  maxTooltipItems,
  maxLegendItems,
  wrapperStyles,
}: JobProfileChartProps) => {

  const [lineVisibility, setLineVisibility] = useState<boolean[]>(() => seriesConfigs.map(() => true))
  const [dotsVisibility, setDotsVisibility] = useState<boolean>(false)
  const [chartType, setChartType] = useState<'area' | 'line'>('area')

  const handleChartTypeChange = (change: SegmentGroupValueChangeDetails) => {
    setChartType(change.value || 'area') // Default to 'area' if no value is provided
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
    <Box style={wrapperStyles}>
      <VStack gap={2} width={'100%'} alignItems={'start'}>
        <HStack width="100%" paddingLeft={8} paddingRight={4}>
          <VStack alignItems={'start'} width={'100%'}>
            <Heading size={'md'} mb={'-10px'}>{profileInfo.title}</Heading>
            <Text textStyle="sm" minH="1.0rem">
              {profileInfo.description || '\u00A0'}
            </Text>
          </VStack>
          <VStack alignItems={'start'}>
            <HStack>
              <SegmentGroup.Root
                defaultValue={chartType}
                size="xs"
                colorPalette="blue"
                onValueChange={handleChartTypeChange}
              >
                <SegmentGroup.Indicator/>
                <SegmentGroup.Items items={[
                  {
                    value: 'area',
                    label: (
                      <HStack>
                        <FaChartArea/>
                        Area
                      </HStack>
                    ),
                  },
                  {
                    value: 'line',
                    label: (
                      <HStack>
                        <FaChartLine/>
                        Line
                      </HStack>
                    ),
                  },
                ]}/>
              </SegmentGroup.Root>

            </HStack>
            <Checkbox.Root
              size={'sm'}
              colorPalette={'blue'}
              checked={dotsVisibility}
              onCheckedChange={(e) => setDotsVisibility(!!e.checked)}
            >
              <Checkbox.HiddenInput/>
              <Checkbox.Control/>
              <Checkbox.Label>
                <Text fontWeight="normal">Show Dots</Text>
              </Checkbox.Label>
            </Checkbox.Root>
          </VStack>
        </HStack>
        <ChartComponent
          chartType={chartType}
          dataItems={dataItems}
          seriesConfigs={seriesConfigs}
          showDots={dotsVisibility}
          profileInfo={profileInfo}
          maxTooltipItems={maxTooltipItems}
          maxLegendItems={maxLegendItems}
          width={chartSize?.width}
          height={chartSize?.height}
          syncId={syncId}
        />
      </VStack>
    </Box>
  )
}

interface ChartComponentProps {
  chartType: 'area' | 'line'
  dataItems: JobProfileDataItem[]
  seriesConfigs: ChartSeriesConfig[]
  showDots: boolean
  syncId?: string
  profileInfo: ProfileInfo
  maxTooltipItems?: number
  maxLegendItems?: number
  width?: number | string
  height?: number | string
}


export const ChartComponent = ({
  chartType,
  dataItems,
  seriesConfigs,
  showDots,
  syncId,
  profileInfo,
  maxTooltipItems,
  maxLegendItems,
  width = '100%',
  height = 400,
}: ChartComponentProps) => {

  const Chart = chartType === 'area' ? AreaChart : LineChart
  const Series = chartType === 'area' ? Area : Line

  const [lineVisibility, setLineVisibility] = useState<boolean[]>(() => seriesConfigs.map(() => true))

  const hiddenSeriesCount = seriesConfigs.length - maxLegendItems
  const handleLegendClick = (dataKey?: DataKey<string>) => {
    const index = seriesConfigs.findIndex((config) => config.dataKey === dataKey)
    setLineVisibility((prev) => {
      const newVisibility = [...prev]
      newVisibility[index] = !newVisibility[index]
      return newVisibility
    })
  }

  return (
    <ResponsiveContainer
      width={width}
      height={height}
      style={{margin: '5px'}}
    >
      <Chart
        data={dataItems}
        syncId={syncId}
      >
        <CartesianGrid strokeDasharray="3 3"/>
        <XAxis
          dataKey="time"
          angle={-45}
          textAnchor="end"
          height={90}
          interval={'equidistantPreserveStart'}
          tickFormatter={dateTimeFormatter}
          style={{
            fontSize: '12px',
          }}
        />
        <YAxis
          label={{
            value: profileInfo.yAxisLabel,
            angle: -90,
            position: 'insideLeft'
          }}
          style={{
            fontSize: '12px',
          }}
        />
        <Tooltip
          content={<CustomTooltip maxItems={maxTooltipItems}/>}
        />
        <Legend
          wrapperStyle={{
            paddingBottom: 16,
            paddingLeft: 12,
            paddingRight: 4,
            maxHeight: '120px',
          }}
          content={<CustomLegend
            maxLegendItems={maxLegendItems}
            hiddenSeriesCount={hiddenSeriesCount}
            onLegendClick={handleLegendClick}
          />}
          verticalAlign={'top'}
        />
        {seriesConfigs.map((config, index) => (
          <Series
            key={config.dataKey}
            type="linear"
            dataKey={config.dataKey}
            stroke={config.lineColor}
            strokeWidth={1.25}
            hide={!lineVisibility[index]}
            dot={showDots ? {stroke: config.lineColor, r: 2} : false}
            activeDot={showDots ? {stroke: config.lineColor, r: 3} : undefined}
            fill={chartType === 'area' ? config.lineColor : undefined}
            stackId={chartType === 'area' ? '1' : undefined}
          />
        ))}
        <Brush
          dataKey="time"
          height={25}
          tickFormatter={dateTimeFormatter}
          style={{overflow: 'visible'}}
        >
          <Chart data={dataItems}>
            {seriesConfigs.map((config, index) => (
              <Series
                key={config.dataKey}
                type="monotone"
                dataKey={config.dataKey}
                stroke={config.lineColor}
                hide={!lineVisibility[index]}
                dot={false}
              />
            ))}
          </Chart>
        </Brush>
      </Chart>
    </ResponsiveContainer>
  )
}
