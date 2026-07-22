import { useEffect, useState } from 'react'
import { Card, CardBody, Skeleton } from '@nextui-org/react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, LineChart, Line, ResponsiveContainer, Legend } from 'recharts'
import { apiGet } from '../api'
import type { ServerInfo } from '../api'
import LogViewer from '../components/LogViewer'

interface ServerStats {
  keys_per_db: number[]
  cpu_pct: number
  memory_mb: string
  history: { time: string; cpu: number; mem: number }[]
}

const tooltipStyle = { background: 'var(--nextui-content1)', border: '1px solid var(--nextui-divider)', borderRadius: 8, padding: '6px 10px', fontSize: 13 }

export default function Dashboard() {
  const [info, setInfo] = useState<ServerInfo | null>(null)
  const [stats, setStats] = useState<ServerStats | null>(null)

  useEffect(() => {
    const fetchInfo = () => apiGet<ServerInfo>('/server/info').then(setInfo).catch(() => { })
    const fetchStats = () => apiGet<ServerStats>('/server/stats').then(setStats).catch(() => { })
    fetchInfo(); fetchStats()
    const t = setInterval(() => { fetchInfo(); fetchStats() }, 3000)
    return () => clearInterval(t)
  }, [])

  if (!info || !stats) return <div className="p-6"><Skeleton className="h-40 rounded-lg" /></div>

  const dbChartData = stats.keys_per_db
    .map((count, i) => ({ name: `DB ${i}`, keys: count }))
    .filter(c => c.keys > 0)

  const cards = [
    { label: '版本', value: info.version },
    { label: '运行时间', value: info.uptime },
    { label: '总键数', value: info.keys.toLocaleString() },
    { label: '内存', value: info.memory },
    { label: 'CPU', value: `${stats.cpu_pct.toFixed(1)}%` },  // 保留一位，折线图 tooltip 会显示两位
    { label: '数据库', value: info.databases },
  ]

  return (
    <div className="p-6 space-y-4">
      <div className="grid grid-cols-3 lg:grid-cols-6 gap-4">
        {cards.map((c) => (
          <Card key={c.label}>
            <CardBody className="text-center">
              <div className="text-sm text-default-500">{c.label}</div>
              <div className="text-2xl font-bold mt-1">{c.value}</div>
            </CardBody>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardBody>
            <div className="text-sm font-semibold mb-2 text-default-600">各数据库键数</div>
            {dbChartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={dbChartData} barSize={40}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" tick={{ fontSize: 12 }} />
                  <YAxis tick={{ fontSize: 12 }} />
                  <Tooltip contentStyle={tooltipStyle} formatter={(value: any) => Number(value).toFixed(2)} />
                  <Bar dataKey="keys" fill="hsl(var(--nextui-primary) / 0.7)" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-48 flex items-center justify-center text-default-400">暂无数据</div>
            )}
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="text-sm font-semibold mb-2 text-default-600">CPU / 内存</div>
            {stats.history && stats.history.length > 1 ? (
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={stats.history}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} interval={9} />
                  <YAxis yAxisId="left" tick={{ fontSize: 12 }} unit="%" />
                  <YAxis yAxisId="right" orientation="right" tick={{ fontSize: 12 }} unit="MB" />
                  <Tooltip contentStyle={tooltipStyle} formatter={(value: any) => Number(value).toFixed(2)} />
                  <Legend />
                  <Line yAxisId="left" type="monotone" dataKey="cpu" stroke="hsl(0, 80%, 60%)" name="CPU %" dot={false} strokeWidth={2} animationDuration={200} />
                  <Line yAxisId="right" type="monotone" dataKey="mem" stroke="hsl(210, 80%, 60%)" name="内存 MB" dot={false} strokeWidth={2} animationDuration={200} />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-48 flex items-center justify-center text-default-400">采集数据中...</div>
            )}
          </CardBody>
        </Card>
      </div>

      <LogViewer />
    </div>
  )
}
