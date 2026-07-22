const BASE = '/api'

export async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export interface KeyItem {
  key: string
  type: string
  ttl: number
  size: number
}

export interface KeyDetail {
  key: string
  type: string
  ttl: number
  value?: string
  values?: string[]
  fields?: Record<string, string>
  members?: { Member: string; Score: number }[] | string[]
}

export interface ServerInfo {
  version: string
  uptime: string
  keys: number
  memory: string
  clients: number
  port: number
  databases: number
}
