import { useState } from 'react'
import { Input, Button, Card, CardBody } from '@nextui-org/react'
import { IconLock } from '@tabler/icons-react'
import { apiPost } from '../api'

export default function LoginPage({ onLogin }: { onLogin: () => void }) {
  const [pwd, setPwd] = useState('')
  const [err, setErr] = useState('')

  const handleLogin = async () => {
    try {
      const res = await apiPost<{ ok: boolean }>('/auth', { password: pwd })
      if (res.ok) onLogin()
      else setErr('密码错误')
    } catch {
      setErr('服务器错误')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-96">
        <CardBody className="gap-4 p-8">
          <div className="text-center">
            <IconLock size={40} className="mx-auto mb-2 text-primary" />
            <h2 className="text-xl font-bold">Godis 认证</h2>
            <p className="text-sm text-default-500 mt-1">请输入访问密码</p>
          </div>
          <Input
            type="password"
            placeholder="密码"
            value={pwd}
            onValueChange={setPwd}
            onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
          />
          {err && <p className="text-sm text-danger">{err}</p>}
          <Button color="primary" onPress={handleLogin} className="w-full">登录</Button>
        </CardBody>
      </Card>
    </div>
  )
}
