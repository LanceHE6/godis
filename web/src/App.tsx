import { useEffect, useState } from 'react'
import { Tabs, Tab } from '@nextui-org/react'
import { IconKey, IconTerminal2, IconChartBar } from '@tabler/icons-react'
import { apiGet } from './api'
import Header from './components/Header'
import Dashboard from './pages/Dashboard'
import KeyBrowser from './pages/KeyBrowser'
import Console from './pages/Console'
import LoginPage from './pages/LoginPage'

const AUTH_KEY = 'godis_authed'

export default function App() {
  const [authed, setAuthed] = useState(() => localStorage.getItem(AUTH_KEY) === '1')
  const [needAuth, setNeedAuth] = useState(false)
  const [loading, setLoading] = useState(true)
  const [dark, setDark] = useState(true)
  const [consoleCmd, setConsoleCmd] = useState('')
  const [consoleHistory, setConsoleHistory] = useState<string[]>([])
  const [consoleResult, setConsoleResult] = useState('')

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
  }, [dark])

  useEffect(() => {
    apiGet<{ requirepass: boolean }>('/auth').then(r => {
      setNeedAuth(r.requirepass)
      if (!r.requirepass) { setAuthed(true); localStorage.setItem(AUTH_KEY, '1') }
    }).finally(() => setLoading(false))
  }, [])

  const handleLogin = () => { setAuthed(true); localStorage.setItem(AUTH_KEY, '1') }
  const handleLogout = () => { setAuthed(false); localStorage.removeItem(AUTH_KEY) }

  if (loading) return null

  const body = (needAuth && !authed)
    ? <LoginPage onLogin={handleLogin} />
    : (
      <Tabs aria-label="Navigation" className="px-6 pt-4" color="primary">
        <Tab key="dashboard" title={<span className="flex items-center gap-1"><IconChartBar size={16} /> 仪表盘</span>}>
          <Dashboard />
        </Tab>
        <Tab key="keys" title={<span className="flex items-center gap-1"><IconKey size={16} /> 键管理</span>}>
          <KeyBrowser />
        </Tab>
        <Tab key="console" title={<span className="flex items-center gap-1"><IconTerminal2 size={16} /> 控制台</span>}>
          <Console cmd={consoleCmd} setCmd={setConsoleCmd}
            history={consoleHistory} setHistory={setConsoleHistory}
            result={consoleResult} setResult={setConsoleResult} />
        </Tab>
      </Tabs>
    )

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Header dark={dark} onToggleDark={setDark} authed={authed} onLogout={handleLogout} />
      {body}
    </div>
  )
}
