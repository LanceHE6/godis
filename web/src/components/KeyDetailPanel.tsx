import { useState } from 'react'
import { Input, Button, Textarea, Divider } from '@nextui-org/react'
import { IconX, IconDeviceFloppy, IconPencil, IconTrash } from '@tabler/icons-react'
import { apiPost } from '../api'
import type { KeyDetail } from '../api'
import ConfirmModal from '../components/ConfirmModal'

export default function KeyDetailPanel({ keyName, detail, db, onEdited, onRename, onClose }: {
  keyName: string; detail: KeyDetail; db: number; onEdited: () => void; onRename?: (name: string) => void; onClose: () => void
}) {
  const [editing, setEditing] = useState(false)
  const [editVal, setEditVal] = useState(detail.value || '')
  const [ttlMode, setTtlMode] = useState(false)
  const [ttlVal, setTtlVal] = useState('')
  const [renameMode, setRenameMode] = useState(false)
  const [newName, setNewName] = useState(keyName)
  const [msg, setMsg] = useState('')

  const [editField, setEditField] = useState<string | null>(null)
  const [editFieldVal, setEditFieldVal] = useState('')
  const [addMode, setAddMode] = useState(false)
  const [addField, setAddField] = useState('')
  const [addVal, setAddVal] = useState('')

  // confirm
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [confirmAction, setConfirmAction] = useState<(() => void) | null>(null)
  const [confirmTitle, setConfirmTitle] = useState('')

  const askConfirm = (title: string, fn: () => void) => {
    setConfirmTitle(title)
    setConfirmAction(() => fn)
    setConfirmOpen(true)
  }

  const doEdit = async (action: string, value: string, field?: string) => {
    try {
      const body: any = { key: keyName, action, value, db }
      if (field !== undefined) body.field = field
      const res = await apiPost<{ status?: string }>('/key/edit', body)
      if (!res.status) { setMsg('服务器返回错误'); return }
      setMsg('操作成功'); setTimeout(() => setMsg(''), 1500)
      onEdited()
    } catch { setMsg('请求失败') }
  }

  const saveValue = () => { doEdit('set_value', editVal); setEditing(false) }
  const saveTTL = () => { doEdit(ttlVal === '-1' ? 'persist' : 'set_ttl', ttlVal); setTtlMode(false) }
  const doRename = () => { const name = newName.trim().replace(/\s/g, ''); if (!name || name === keyName) return; doEdit('rename', name); onRename?.(name); setRenameMode(false) }

  const membersArr = detail.members as { Member: string; Score: number }[] | undefined

  return (
    <div className="border border-divider rounded-lg p-4 space-y-4 bg-content1 text-sm">
      <ConfirmModal open={confirmOpen} title={confirmTitle} message="此操作不可撤销" onConfirm={() => { confirmAction?.(); setConfirmOpen(false) }} onCancel={() => setConfirmOpen(false)} />

      <div className="flex justify-end">
        <button onClick={onClose}><IconX size={18} /></button>
      </div>

      <div className="flex items-center gap-2 min-w-0">
        <span className="text-default-500 shrink-0">Key:</span>
        {renameMode ? (
          <div className="flex items-center gap-1 flex-1 min-w-0">
            <Input size="sm" value={newName} onValueChange={setNewName} onKeyDown={(e: any) => e.key === 'Enter' && doRename()} />
            <Button size="sm" isIconOnly color="primary" onPress={doRename}><IconDeviceFloppy size={14} /></Button>
            <Button size="sm" isIconOnly variant="flat" onPress={() => setRenameMode(false)}><IconX size={14} /></Button>
          </div>
        ) : (
          <>
            <span className="font-mono font-bold truncate">{keyName}</span>
            <button onClick={() => { setNewName(keyName); setRenameMode(true) }} className="text-default-400 hover:text-primary shrink-0">
              <IconPencil size={14} />
            </button>
          </>
        )}
      </div>

      <div className="flex items-center gap-4">
        <span><span className="text-default-500">类型:</span> <ChipInner label={detail.type} /></span>
        {ttlMode ? (
          <div className="flex items-center gap-1">
            <Input size="sm" className="w-20" placeholder="秒/-1" value={ttlVal} onValueChange={setTtlVal} onKeyDown={(e: any) => e.key === 'Enter' && saveTTL()} />
            <Button size="sm" isIconOnly color="primary" onPress={saveTTL}><IconDeviceFloppy size={12} /></Button>
            <Button size="sm" isIconOnly variant="flat" onPress={() => setTtlMode(false)}><IconX size={12} /></Button>
          </div>
        ) : (
          <span className="flex items-center gap-1">
            <span className="text-default-500">TTL:</span> {detail.ttl === -1 ? '永久' : `${detail.ttl}s`}
            <button onClick={() => { setTtlVal(detail.ttl === -1 ? '-1' : String(detail.ttl)); setTtlMode(true) }} className="text-default-400 hover:text-primary">
              <IconPencil size={12} />
            </button>
          </span>
        )}
      </div>

      <Divider />
      {msg && <p className="text-xs text-success">{msg}</p>}

      {/* STRING */}
      {detail.type === 'string' && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-default-500">值:</span>
            {!editing && <button onClick={() => { setEditVal(detail.value || ''); setEditing(true) }}><IconPencil size={14} /></button>}
          </div>
          {editing ? (
            <div className="flex gap-1">
              <Textarea size="sm" value={editVal} onValueChange={setEditVal} minRows={3} className="flex-1" />
              <Button size="sm" isIconOnly color="primary" onPress={saveValue}><IconDeviceFloppy size={14} /></Button>
            </div>
          ) : (
            <pre className="text-sm bg-default-100 p-2 rounded whitespace-pre-wrap break-all">{detail.value || '(空)'}</pre>
          )}
        </div>
      )}

      {/* HASH: field 左, value 右 */}
      {detail.type === 'hash' && detail.fields && (
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-default-500">Fields ({Object.keys(detail.fields).length}):</span>
            <button onClick={() => { setAddMode(true); setAddField(''); setAddVal('') }} className="text-primary hover:text-primary-500 text-lg leading-none">+</button>
          </div>
          <div className="border border-divider rounded overflow-hidden">
            <div className="flex text-xs text-default-500 bg-default-50 px-3 py-1.5 border-b border-divider font-semibold">
              <span className="flex-1">Field</span>
              <span className="flex-1">Value</span>
              <span className="w-14"></span>
            </div>
            <div className="max-h-80 overflow-y-auto divide-y divide-divider">
              {addMode && (
                <div className="flex gap-1 px-2 py-1.5">
                  <Input size="sm" placeholder="field" value={addField} onValueChange={setAddField} className="flex-1" />
                  <Input size="sm" placeholder="value" value={addVal} onValueChange={setAddVal} className="flex-1" />
                  <Button size="sm" isIconOnly color="primary" onPress={() => { if (addField) { doEdit('hset', addVal, addField); setAddMode(false) } }}><IconDeviceFloppy size={12} /></Button>
                  <Button size="sm" isIconOnly variant="flat" onPress={() => setAddMode(false)}><IconX size={12} /></Button>
                </div>
              )}
              {Object.entries(detail.fields).map(([f, v]) => (
                <div key={f} className="flex items-center px-3 py-2">
                  {editField === f ? (
                    <>
                      <span className="font-mono text-sm flex-1">{f}</span>
                      <Input size="sm" value={editFieldVal} onValueChange={setEditFieldVal} className="flex-1" />
                      <div className="flex gap-1 w-14 justify-end">
                        <Button size="sm" isIconOnly color="primary" onPress={() => { doEdit('hset', editFieldVal, f); setEditField(null) }}><IconDeviceFloppy size={12} /></Button>
                        <Button size="sm" isIconOnly variant="flat" onPress={() => setEditField(null)}><IconX size={12} /></Button>
                      </div>
                    </>
                  ) : (
                    <>
                      <span className="font-mono text-sm font-semibold flex-1 truncate">{f}</span>
                      <span className="text-sm text-default-700 flex-1 truncate">{v}</span>
                      <div className="flex gap-1 w-14 justify-end">
                        <button className="text-default-400 hover:text-primary" onClick={() => { setEditField(f); setEditFieldVal(v) }}><IconPencil size={14} /></button>
                        <button className="text-default-400 hover:text-danger" onClick={() => askConfirm(`删除 field "${f}"?`, () => doEdit('hdel', '', f))}><IconTrash size={14} /></button>
                      </div>
                    </>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* LIST */}
      {detail.type === 'list' && detail.values && (
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-default-500">Elements ({detail.values.length}):</span>
            <button onClick={() => { setAddMode(true); setAddVal('') }} className="text-primary hover:text-primary-500 text-lg leading-none">+</button>
          </div>
          <div className="max-h-80 overflow-y-auto space-y-1">
            {addMode && (
              <div className="flex gap-1">
                <Input size="sm" placeholder="新元素值 (RPUSH)" value={addVal} onValueChange={setAddVal} className="flex-1" />
                <Button size="sm" isIconOnly color="primary" onPress={() => { if (addVal) { doEdit('rpush', addVal); setAddMode(false) } }}><IconDeviceFloppy size={12} /></Button>
                <Button size="sm" isIconOnly variant="flat" onPress={() => setAddMode(false)}><IconX size={12} /></Button>
              </div>
            )}
            {detail.values.map((v, i) => (
              <div key={i} className="bg-default-100 px-2 py-1.5 rounded flex items-center group">
                {editField === String(i) ? (
                  <div className="flex gap-1 flex-1 items-center">
                    <span className="text-default-500 text-sm shrink-0">[{i}]</span>
                    <Input size="sm" value={editFieldVal} onValueChange={setEditFieldVal} className="flex-1" />
                    <Button size="sm" isIconOnly color="primary" onPress={() => { doEdit('lset', editFieldVal, String(i)); setEditField(null) }}><IconDeviceFloppy size={12} /></Button>
                    <Button size="sm" isIconOnly variant="flat" onPress={() => setEditField(null)}><IconX size={12} /></Button>
                  </div>
                ) : (
                  <div className="flex items-center justify-between w-full">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="text-default-500 text-sm shrink-0">[{i}]</span>
                      <span className="font-mono text-sm truncate">{v}</span>
                    </div>
                    <div className="flex items-center gap-1 shrink-0 ml-2">
                      <button className="text-default-400 hover:text-primary" onClick={() => { setEditField(String(i)); setEditFieldVal(v) }}><IconPencil size={14} /></button>
                      <button className="text-default-400 hover:text-danger" onClick={() => askConfirm(`删除列表元素 "[${i}] ${v}"?`, () => doEdit('lrem', v, String(i)))}><IconTrash size={14} /></button>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* SET */}
      {detail.type === 'set' && detail.members && (
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-default-500">Members ({(detail.members as string[]).length}):</span>
            <button onClick={() => { setAddMode(true); setAddVal('') }} className="text-primary hover:text-primary-500 text-lg leading-none">+</button>
          </div>
          <div className="max-h-80 overflow-y-auto space-y-1">
            {addMode && (
              <div className="flex gap-1">
                <Input size="sm" placeholder="member" value={addVal} onValueChange={setAddVal} className="flex-1" />
                <Button size="sm" isIconOnly color="primary" onPress={() => { if (addVal) { doEdit('sadd', addVal); setAddMode(false) } }}><IconDeviceFloppy size={12} /></Button>
                <Button size="sm" isIconOnly variant="flat" onPress={() => setAddMode(false)}><IconX size={12} /></Button>
              </div>
            )}
            {(detail.members as string[]).map((m, i) => (
              <div key={i} className="bg-default-100 px-2 py-1.5 rounded flex items-center justify-between group">
                <span className="font-mono text-sm truncate">{m}</span>
                <button className="shrink-0 ml-2 text-default-400 hover:text-danger" onClick={() => askConfirm(`删除成员 "${m}"?`, () => doEdit('srem', m))}><IconTrash size={14} /></button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ZSET */}
      {detail.type === 'zset' && membersArr && (
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-default-500">Members ({membersArr.length}):</span>
            <button onClick={() => { setAddMode(true); setAddField(''); setAddVal('') }} className="text-primary hover:text-primary-500 text-lg leading-none">+</button>
          </div>
          <div className="max-h-80 overflow-y-auto space-y-1">
            {addMode && (
              <div className="flex gap-1">
                <Input size="sm" placeholder="member" value={addField} onValueChange={setAddField} className="flex-1" />
                <Input size="sm" placeholder="score" value={addVal} onValueChange={setAddVal} className="w-20" />
                <Button size="sm" isIconOnly color="primary" onPress={() => { if (addField && addVal) { doEdit('zset_score', addVal, addField); setAddMode(false) } }}><IconDeviceFloppy size={12} /></Button>
                <Button size="sm" isIconOnly variant="flat" onPress={() => setAddMode(false)}><IconX size={12} /></Button>
              </div>
            )}
            {membersArr.map((m, i) => (
              <div key={i} className="bg-default-100 px-2 py-1.5 rounded flex items-center group">
                {editField === m.Member ? (
                  <div className="flex gap-1 flex-1 items-center">
                    <span className="font-mono text-sm truncate w-24">{m.Member}</span>
                    <Input size="sm" value={editFieldVal} onValueChange={setEditFieldVal} className="w-20" />
                    <Button size="sm" isIconOnly color="primary" onPress={() => { doEdit('zset_score', editFieldVal, m.Member); setEditField(null) }}><IconDeviceFloppy size={12} /></Button>
                    <Button size="sm" isIconOnly variant="flat" onPress={() => setEditField(null)}><IconX size={12} /></Button>
                  </div>
                ) : (
                  <div className="flex items-center justify-between w-full">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="font-mono text-sm truncate">{m.Member}</span>
                      <span className="text-default-500 text-sm shrink-0">{m.Score}</span>
                    </div>
                    <button className="shrink-0 ml-2 text-default-400 hover:text-primary" onClick={() => { setEditField(m.Member); setEditFieldVal(String(m.Score)) }}><IconPencil size={14} /></button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      <Divider />
    </div>
  )
}

function ChipInner({ label }: { label: string }) {
  return <span className="inline-block px-2 py-0.5 rounded bg-default-200 text-default-800 text-sm">{label}</span>
}
