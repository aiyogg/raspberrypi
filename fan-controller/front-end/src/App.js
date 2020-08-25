import React from 'react'
import { Chart, Line, Point, Slider } from 'bizcharts'
import './App.css'

class App extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      temperatures: [],
    }
  }
  componentWillMount() {
    fetch('http://192.168.0.100:10001/get')
      .then((rsp) => rsp.json())
      .then((json) => {
        console.log(json)
        if (json.code === 0) {
          this.setState({
            temperatures: json.temperatures,
          })
        }
      })
  }
  render() {
    const data = this.state.temperatures
    const scale = {
      time: {
        formatter: toLocalTime,
      },
      temperature: {
        alias: '温度',
      },
    }
    function toLocalTime(t) {
      return new Date(t)
        .toLocaleTimeString('zh-CN', { hour12: false })
        .substring(0, 5)
    }

    return (
      <div className="chart">
        <Chart
          scale={scale}
          padding={[30, 20, 50, 40]}
          autoFit
          height={600}
          data={data}
        >
          <Line
            shape="smooth"
            position="time*temperature"
            color="l (270) 0:#1E9600 .5:#FFF200 1:#FF0000"
          />
          <Point size={2} position="time*temperature" />
          <Slider start={0.5} end={1} formatter={toLocalTime} />
        </Chart>
      </div>
    )
  }
}

export default App
