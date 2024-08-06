import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import moment from 'moment'
import { useEffect, useState } from 'react'
import { AxisDomain, DataKey } from 'recharts/types/util/types'

interface HostDetailsChartProps {
  dataItems: ChartDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  containerProps: {
    width: number | string;
    height: number | string;
  }
  yAxisDomain?: AxisDomain;
  xAxisDomain?: AxisDomain;
  isShowDataPoints?: boolean;
}

const HostDetailsChart = ({
  dataItems,
  seriesConfigs,
  containerProps,
  xAxisDomain,
  yAxisDomain,
  isShowDataPoints
}: HostDetailsChartProps) => {

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


  const dateTimeFormatter = (datetime: number) => {
    return moment(datetime).format('MMM D, HH:mm')
  }

  if (!dataItems.length || !seriesConfigs.length) {
    return
  }

  return (
    <ResponsiveContainer width={containerProps.width} height={containerProps.height}>
      <LineChart
        data={dataItems}
        key={JSON.stringify(dataItems)}
      >
        <CartesianGrid strokeDasharray="3 3"/>
        <XAxis
          dataKey="timestamp"
          type={'number'}
          scale={'time'}
          domain={xAxisDomain || ['auto', 'auto']}
          tickFormatter={dateTimeFormatter}
          angle={-45}
          interval={'equidistantPreserveStart'}
          tickMargin={30}
          height={100}

        />
        <YAxis
          domain={ yAxisDomain || ['auto', 'auto']}
        />
        <Tooltip labelFormatter={dateTimeFormatter}/>
        <Legend
          verticalAlign={'top'}
          height={50}
          onClick={(payload) => {
            handleLegendClick(payload.dataKey)
          }}
        />
        {seriesConfigs.map((config, index) => {
          return (
            <Line
              key={config.dataKey}
              type="linear"
              dataKey={config.dataKey}
              stroke={config.lineColor}
              dot={isShowDataPoints}
              name={config.label}
              strokeWidth={config.strokeWidth}
              hide={!lineVisibility[index]}
            />
          )
        })}
      </LineChart>
    </ResponsiveContainer>
  )
}

export default HostDetailsChart
