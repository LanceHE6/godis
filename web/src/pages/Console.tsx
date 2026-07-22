import { useEffect, useState } from 'react'
import { Input, Button, Textarea } from '@nextui-org/react'
import { IconSend, IconTrash } from '@tabler/icons-react'
import { apiGet, apiPost } from '../api'

export default function Console({ cmd, setCmd, history, setHistory, result, setResult }: {
  cmd: string; setCmd: (v: string) => void
  history: string[]; setHistory: (v: string[]) => void
  result: string; setResult: (v: string) => void
}) {
  const [allCmds, setAllCmds] = useState<string[]>([])
  const [hints, setHints] = useState<string[]>([])

  useEffect(() => {
    apiGet<{ commands: string[] }>('/commands').then(r => {
      setAllCmds(r.commands?.sort() || [])
    })
  }, [])

  const upperCmd = cmd.toUpperCase()
  useEffect(() => {
    if (cmd.length > 0) {
      setHints(allCmds.filter(c => c.startsWith(upperCmd)).slice(0, 8))
    } else {
      setHints([])
    }
  }, [cmd, allCmds])

  const addHistory = (c: string) => {
    // 去重：移到最前面，保留最近 50 条
    setHistory([c, ...history.filter(h => h !== c)].slice(0, 49))
  }

  const execute = async () => {
    if (!cmd.trim()) return
    try {
      const res = await apiPost<{ reply: string }>('/exec', { command: cmd })
      const out = `> ${cmd}\n${res.reply}`
      setResult(out + (result ? '\n\n' + result : ''))  // 新结果放上面
      addHistory(cmd)
      setCmd('')
      setHints([])
    } catch (e: any) {
      setResult(`> ${cmd}\nERR: ${e}` + (result ? '\n\n' + result : ''))
    }
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex gap-2 relative">
        <div className="flex-1 relative">
          <Input className="font-mono" placeholder="输入 Godis 命令" value={cmd}
            onValueChange={setCmd}
            onKeyDown={(e: any) => e.key === 'Enter' && execute()}
          />
          {hints.length > 0 && (
            <div className="absolute top-full left-0 z-50 mt-1 w-64 bg-content1 border border-divider rounded-lg shadow-lg p-1">
              {hints.map(h => (
                <button key={h} className="w-full text-left px-3 py-1.5 rounded text-sm font-mono hover:bg-primary-100 dark:hover:bg-primary-50"
                  onClick={() => { setCmd(h); setHints([]) }}>
                  {h}
                </button>
              ))}
            </div>
          )}
        </div>
        <Button color="primary" onPress={execute}><IconSend size={16} /> 执行</Button>
        <Button variant="flat" onPress={() => setResult('')}><IconTrash size={16} /> 清空</Button>
      </div>

      {history.length > 0 && (
        <div className="flex gap-1 flex-wrap">
          {history.slice(0, 10).map((h, i) => (
            <button key={i} className="text-xs bg-default-100 px-2 py-1 rounded hover:bg-default-200 font-mono"
              onClick={() => setCmd(h)}>{h}</button>
          ))}
        </div>
      )}

      <Textarea className="font-mono text-sm" minRows={15} maxRows={30} readOnly value={result}
        placeholder="结果将显示在这里..." />
    </div>
  )
}
