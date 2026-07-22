import { useEffect, useRef, useState } from 'react'
import { Card, CardBody } from '@nextui-org/react'
import { apiGet } from '../api'

function highlightLine(line: string): string {
  if (line.includes('[ERROR]')) return 'text-red-600 dark:text-red-400 font-semibold'
  if (line.includes('[WARN]')) return 'text-amber-600 dark:text-yellow-400 font-semibold'
  if (line.includes('[DEBUG]')) return 'text-gray-400 dark:text-gray-500'
  if (/exceed|expire|trigger/i.test(line)) return 'text-cyan-600 dark:text-cyan-400'
  return 'text-gray-800 dark:text-gray-200'
}

export default function LogViewer() {
  const [logs, setLogs] = useState<string[]>([])
  const logRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const fetchLogs = () => apiGet<{ logs: string[] }>('/logs').then(r => setLogs(r.logs || [])).catch(() => { })
    fetchLogs()
    const t = setInterval(fetchLogs, 2000)
    return () => clearInterval(t)
  }, [])

  useEffect(() => {
    if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight
  }, [logs])

  return (
    <Card className="h-full flex flex-col">
      <CardBody className="p-0 flex flex-col flex-1">
        <div className="text-xs font-semibold px-4 py-2 border-b border-divider bg-default-50 flex items-center gap-2 shrink-0">
          实时日志
          <span className="text-default-400 font-normal">({logs.length} lines)</span>
          <span className="ml-auto inline-flex items-center gap-2 text-default-400">
            <span className="inline-block w-2 h-2 rounded-full bg-green-500" /> INFO
            <span className="inline-block w-2 h-2 rounded-full bg-amber-500" /> WARN
            <span className="inline-block w-2 h-2 rounded-full bg-red-500" /> ERR
          </span>
        </div>
        <div ref={logRef} className="overflow-y-auto p-3 h-full font-mono text-sm leading-relaxed bg-default-100 rounded-b-lg dark:border-gray-600" style={{ maxHeight: '20rem' }}>
          {logs.length === 0 ? (
            <span className="text-default-500">暂无日志...</span>
          ) : (
            logs.map((l, i) => (
              <div key={i} className={highlightLine(l)}>{l}</div>
            ))
          )}
        </div>
      </CardBody>
    </Card>
  )
}
