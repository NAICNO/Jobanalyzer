import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import moment from 'moment'
import { useState } from 'react'
import { DataKey } from 'recharts/types/util/types'

interface HostDetailsChartProps {
  dataItems: ChartDataItem[];
  seriesConfigs: ChartSeriesConfig[];
  containerProps: {
    width: number | string;
    height: number | string;
  }
  isShowDataPoints?: boolean;
}

const HostDetailsChart = ({dataItems, seriesConfigs, containerProps, isShowDataPoints}: HostDetailsChartProps) => {

  const [lineVisibility, setLineVisibility] = useState<boolean[]>(() => seriesConfigs.map(() => true))

  const handleLegendClick = (dataKey?: DataKey<string>) => {
    const index = seriesConfigs.findIndex((config) => config.dataKey === dataKey)
    setLineVisibility((prev) => {
      const newVisibility = [...prev]
      newVisibility[index] = !newVisibility[index]
      return newVisibility
    })
  }

  const dateTimeFormatter = (datetime: number) => {
    return moment(datetime).format('MMM D, HH:mm')
  }

  return (
    <ResponsiveContainer width={containerProps.width} height={containerProps.height}>
      <LineChart
        data={dataItems}
      >
        <CartesianGrid strokeDasharray="3 3"/>
        <XAxis
          dataKey="timestamp"
          type={'number'}
          scale={'time'}
          domain={['auto', 'auto']}
          tickFormatter={dateTimeFormatter}
          angle={-45}
          interval={'equidistantPreserveStart'}
          tickMargin={30}
          height={100}

        />
        <YAxis
          domain={([dataMin, dataMax]) => {
            const min = dataMin
            const max = Math.round(dataMax / 100) * 100
            return [min, max]
          }}
        />
        <Tooltip labelFormatter={dateTimeFormatter}/>
        <Legend
          verticalAlign={'top'}
          height={50}
          onClick={(payload) => {
            handleLegendClick(payload.dataKey)
          }}
        />
        {seriesConfigs.map((config) => {
          return (
            <Line
              key={config.dataKey}
              type="linear"
              dataKey={config.dataKey}
              stroke={config.lineColor}
              dot={isShowDataPoints}
              name={config.label}
              strokeWidth={config.strokeWidth}
              hide={lineVisibility[seriesConfigs.indexOf(config)]}
            />
          )
        })}
      </LineChart>
    </ResponsiveContainer>
  )
}

export default HostDetailsChart
