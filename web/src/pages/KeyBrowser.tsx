import { useEffect, useMemo, useRef, useState } from 'react'
import { Input, Button, Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip, Checkbox, Tooltip, Select, SelectItem } from '@nextui-org/react'
import { IconSearch, IconTrash, IconRefresh, IconEye, IconCaretDown, IconCaretUp, IconArrowsSort, IconPlus } from '@tabler/icons-react'
import { apiGet, apiPost } from '../api'
import type { KeyItem, KeyDetail } from '../api'
import KeyDetailPanel from '../components/KeyDetailPanel'
import ConfirmModal from '../components/ConfirmModal'
import AddKeyModal from '../components/AddKeyModal'

const typeColors: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'secondary'> = {
  string: 'primary', hash: 'success', list: 'warning', set: 'danger', zset: 'secondary',
}

type SortField = 'key' | 'type' | 'ttl'

export default function KeyBrowser() {
  const [keys, setKeys] = useState<KeyItem[]>([])
  const [pattern, setPattern] = useState('*')
  const [db, setDb] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [detailKey, setDetailKey] = useState<string | null>(null)
  const detailKeyRef = useRef<string | null>(null)
  useEffect(() => { detailKeyRef.current = detailKey }, [detailKey])
  const [detail, setDetail] = useState<KeyDetail | null>(null)
  const [sortField, setSortField] = useState<SortField>('key')
  const [sortAsc, setSortAsc] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 15

  // Add key modal
  const [addOpen, setAddOpen] = useState(false)

  // Confirm modal
  const [confirmOpen, setConfirmOpen] = useState(false)

  const loadKeys = () => {
    setDetailKey(null); setDetail(null)
    refreshKeys()
  }

  const refreshKeys = () => {
    apiGet<{ keys: KeyItem[]; total: number }>(`/keys?db=${db}&pattern=${encodeURIComponent(pattern)}&page=${page}&page_size=${pageSize}`)
      .then(r => { setKeys(r.keys); setTotal(r.total) }).catch(console.error)
  }
  useEffect(() => {
    setPage(1)
  }, [db, pattern])

  useEffect(() => {
    loadKeys()
    const t = setInterval(refreshKeys, 5000)
    return () => clearInterval(t)
  }, [db, pattern, page])

  const sortedKeys = useMemo(() => {
    const sorted = [...keys].sort((a, b) => {
      let cmp = 0
      if (sortField === 'key') cmp = a.key.localeCompare(b.key)
      else if (sortField === 'type') cmp = a.type.localeCompare(b.type) || a.key.localeCompare(b.key)
      else if (sortField === 'ttl') { const ta = a.ttl === -1 ? Infinity : a.ttl; const tb = b.ttl === -1 ? Infinity : b.ttl; cmp = ta - tb || a.key.localeCompare(b.key) }
      return sortAsc ? cmp : -cmp
    })
    return sorted
  }, [keys, sortField, sortAsc])

  const handleSort = (field: SortField) => { if (sortField === field) setSortAsc(!sortAsc); else { setSortField(field); setSortAsc(true) } }
  const sortIcon = (field: SortField) => {
    if (sortField !== field) return <IconArrowsSort size={14} className="text-default-400" />
    return sortAsc ? <IconCaretUp size={14} className="text-primary" /> : <IconCaretDown size={14} className="text-primary" />
  }

  const viewKey = async (key: string) => {
    setDetailKey(key)
    try { const d = await apiGet<KeyDetail>(`/key?db=${db}&key=${encodeURIComponent(key)}`); setDetail(d) }
    catch { setDetail(null) }
  }

  const delKeys = async () => {
    await apiPost('/keys/delete', { keys: [...selected] })
    setSelected(new Set()); loadKeys()
  }

  const onEdited = () => { loadKeys(); if (detailKey) viewKey(detailKey) }

  const toggleKey = (key: string) => { setSelected(prev => { const next = new Set(prev); if (next.has(key)) next.delete(key); else next.add(key); return next }) }
  const toggleAll = () => { if (selected.size === sortedKeys.length) setSelected(new Set()); else setSelected(new Set(sortedKeys.map(k => k.key))) }

  return (
    <div className="p-6 flex gap-4">
      <ConfirmModal open={confirmOpen} title="确认删除" message={`将删除 ${selected.size} 个 key，不可撤销`} onConfirm={() => { delKeys(); setConfirmOpen(false) }} onCancel={() => setConfirmOpen(false)} />
      <AddKeyModal open={addOpen} onClose={() => setAddOpen(false)} onCreated={loadKeys} db={db} />

      <div className="flex-1 space-y-4">
        <div className="flex gap-2 items-center">
          <Select size="sm" className="w-24" defaultSelectedKeys={['0']}
            onChange={(e: any) => setDb(parseInt(e.target.value))}>
            {Array.from({ length: 16 }, (_, i) => (
              <SelectItem key={String(i)} textValue={`DB ${i}`}>DB {i}</SelectItem>
            ))}
          </Select>
          <Input className="max-w-xs" placeholder="匹配模式" value={pattern} onValueChange={setPattern}
            startContent={<IconSearch size={16} />}
            onKeyDown={(e: any) => e.key === 'Enter' && loadKeys()}
          />
          <Button color="primary" variant="flat" onPress={loadKeys}><IconRefresh size={16} /> 搜索</Button>
          <Button color="success" variant="flat" onPress={() => setAddOpen(true)}><IconPlus size={16} /> 新增</Button>
          {selected.size > 0 && (
            <Button color="danger" onPress={() => setConfirmOpen(true)}><IconTrash size={16} /> 删除 ({selected.size})</Button>
          )}
        </div>

        <Table aria-label="Keys" removeWrapper>
          <TableHeader>
            <TableColumn><Checkbox size="sm" isSelected={selected.size === sortedKeys.length && sortedKeys.length > 0} onValueChange={toggleAll} /></TableColumn>
            <TableColumn><button className="flex items-center gap-1 font-semibold px-2 py-1 rounded hover:bg-default-100" onClick={() => handleSort('key')}><span className="text-sm">Key</span> {sortIcon('key')}</button></TableColumn>
            <TableColumn><button className="flex items-center gap-1 font-semibold px-2 py-1 rounded hover:bg-default-100" onClick={() => handleSort('type')}><span className="text-sm">类型</span> {sortIcon('type')}</button></TableColumn>
            <TableColumn><button className="flex items-center gap-1 font-semibold px-2 py-1 rounded hover:bg-default-100" onClick={() => handleSort('ttl')}><span className="text-sm">TTL</span> {sortIcon('ttl')}</button></TableColumn>
            <TableColumn><span className="font-semibold text-sm">操作</span></TableColumn>
          </TableHeader>
          <TableBody emptyContent="暂无数据">
            {sortedKeys.map(k => (
              <TableRow key={k.key} className={k.key === detailKey ? 'bg-primary-100 dark:bg-primary-500/20' : ''}>
                <TableCell><Checkbox size="sm" isSelected={selected.has(k.key)} onValueChange={() => toggleKey(k.key)} /></TableCell>
                <TableCell className="font-mono text-sm cursor-pointer hover:text-primary" onClick={() => viewKey(k.key)}>{k.key}</TableCell>
                <TableCell><Chip size="sm" color={typeColors[k.type] || 'default'} variant="flat">{k.type}</Chip></TableCell>
                <TableCell className="text-sm">{k.ttl === -1 ? '永久' : `${k.ttl}s`}</TableCell>
                <TableCell><Tooltip content="查看详情"><button onClick={() => viewKey(k.key)} className="text-default-500 hover:text-primary"><IconEye size={16} /></button></Tooltip></TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        {total > pageSize && (
          <div className="flex items-center justify-between mt-2 text-sm">
            <span className="text-default-500">
              共 {total} 条，第 {(page - 1) * pageSize + 1}-{Math.min(page * pageSize, total)} 条
            </span>
            <div className="flex gap-1">
              <Button size="sm" variant="flat" isDisabled={page <= 1} onPress={() => setPage(page - 1)}>上一页</Button>
              <span className="px-3 py-1 text-default-500">{page} / {Math.ceil(total / pageSize)}</span>
              <Button size="sm" variant="flat" isDisabled={page * pageSize >= total} onPress={() => setPage(page + 1)}>下一页</Button>
            </div>
          </div>
        )}
      </div>

      {detailKey && detail && (
        <div className="w-96 shrink-0 mt-12">
          <KeyDetailPanel keyName={detailKey} detail={detail} db={db} onEdited={onEdited}
            onRename={(newName: string) => { setDetailKey(newName); onEdited() }}
            onClose={() => { setDetailKey(null); setDetail(null) }} />
        </div>
      )}
    </div>
  )
}
