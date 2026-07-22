import { useState } from 'react'
import { Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, Button, Input, Textarea, Select, SelectItem, Divider } from '@nextui-org/react'
import { IconPlus, IconTrash } from '@tabler/icons-react'
import { apiPost } from '../api'

interface KeyValuePair { field: string; value: string }
interface ZSetMember { member: string; score: string }

export default function AddKeyModal({ open, onClose, onCreated, db }: {
  open: boolean; onClose: () => void; onCreated: () => void; db: number
}) {
  const [keyName, setKeyName] = useState('')
  const [keyType, setKeyType] = useState('string')
  const [ttl, setTtl] = useState('')
  const [msg, setMsg] = useState('')

  const [strVal, setStrVal] = useState('')
  const [hashFields, setHashFields] = useState<KeyValuePair[]>([{ field: '', value: '' }])
  const [items, setItems] = useState<string[]>([''])
  const [zsetMembers, setZsetMembers] = useState<ZSetMember[]>([{ member: '', score: '0' }])

  const reset = () => {
    setKeyName(''); setKeyType('string'); setTtl(''); setMsg('')
    setStrVal(''); setHashFields([{ field: '', value: '' }])
    setItems(['']); setZsetMembers([{ member: '', score: '0' }])
  }

  const handleClose = () => { reset(); onClose() }

  const create = async () => {
    const name = keyName.trim().replace(/\s/g, '')
    if (!name) { setMsg('请输入 Key 名'); return }
    try {
      let cmds: string[] = []
      switch (keyType) {
        case 'string':
          cmds = ttl ? [`SET ${name} ${strVal || ''} EX ${ttl}`] : [`SET ${name} ${strVal || ''}`]
          break
        case 'hash':
          const validHash = hashFields.filter(f => f.field.trim())
          if (validHash.length === 0) { setMsg('至少需要一个 field'); return }
          cmds = [`HSET ${name} ${validHash.flatMap(f => [f.field, f.value || '']).join(' ')}`]
          if (ttl) cmds.push(`EXPIRE ${name} ${ttl}`)
          break
        case 'list':
          const validList = items.filter(v => v.trim())
          if (validList.length === 0) { setMsg('至少需要元素'); return }
          cmds = [`LPUSH ${name} ${[...validList].reverse().join(' ')}`]
          if (ttl) cmds.push(`EXPIRE ${name} ${ttl}`)
          break
        case 'set':
          const validSet = items.filter(v => v.trim())
          if (validSet.length === 0) { setMsg('至少需要成员'); return }
          cmds = [`SADD ${name} ${validSet.join(' ')}`]
          if (ttl) cmds.push(`EXPIRE ${name} ${ttl}`)
          break
        case 'zset':
          const validZset = zsetMembers.filter(m => m.member.trim())
          if (validZset.length === 0) { setMsg('至少需要成员'); return }
          cmds = [`ZADD ${name} ${validZset.flatMap(m => [m.score || '0', m.member]).join(' ')}`]
          if (ttl) cmds.push(`EXPIRE ${name} ${ttl}`)
          break
      }
      for (const cmd of cmds) {
        const res = await apiPost<{ reply: string }>('/exec', { command: cmd, db })
        if (res.reply.startsWith('-')) {
          setMsg(res.reply.replace(/^-/, ''))
          return
        }
      }
      handleClose(); onCreated()
    } catch { setMsg('创建失败') }
  }

  const addHashField = () => setHashFields([...hashFields, { field: '', value: '' }])
  const updateHashField = (i: number, f: Partial<KeyValuePair>) => { const n = [...hashFields]; n[i] = { ...n[i], ...f }; setHashFields(n) }
  const removeHashField = (i: number) => { if (hashFields.length <= 1) return; setHashFields(hashFields.filter((_, idx) => idx !== i)) }

  const addItem = () => setItems([...items, ''])
  const updateItem = (i: number, v: string) => { const n = [...items]; n[i] = v; setItems(n) }
  const removeItem = (i: number) => { if (items.length <= 1) return; setItems(items.filter((_, idx) => idx !== i)) }

  const addZsetMember = () => setZsetMembers([...zsetMembers, { member: '', score: '0' }])
  const updateZsetMember = (i: number, m: Partial<ZSetMember>) => { const n = [...zsetMembers]; n[i] = { ...n[i], ...m }; setZsetMembers(n) }
  const removeZsetMember = (i: number) => { if (zsetMembers.length <= 1) return; setZsetMembers(zsetMembers.filter((_, idx) => idx !== i)) }

  const labelClass = "text-xs text-default-500 font-medium"

  return (
    <Modal isOpen={open} onClose={handleClose} size="lg" scrollBehavior="inside">
      <ModalContent>
        <ModalHeader>新增 Key</ModalHeader>
        <ModalBody className="space-y-4">
          <div className="flex gap-2 items-end">
            <div className="flex-1">
              <div className={labelClass}>Key 名称</div>
              <Input size="sm" placeholder="mykey" value={keyName} onValueChange={setKeyName} />
            </div>
            <div className="w-32">
              <div className={labelClass}>类型</div>
              <Select size="sm" defaultSelectedKeys={['string']} onChange={(e) => setKeyType(e.target.value)}>
                <SelectItem key="string">string</SelectItem>
                <SelectItem key="hash">hash</SelectItem>
                <SelectItem key="list">list</SelectItem>
                <SelectItem key="set">set</SelectItem>
                <SelectItem key="zset">zset</SelectItem>
              </Select>
            </div>
            <div className="w-24">
              <div className={labelClass}>TTL（秒）</div>
              <Input size="sm" placeholder="可选" value={ttl} onValueChange={setTtl} />
            </div>
          </div>

          <Divider />

          {/* string */}
          {keyType === 'string' && (
            <div>
              <div className={labelClass}>值</div>
              <Textarea size="sm" placeholder="输入字符串值" value={strVal} onValueChange={setStrVal} minRows={3} />
            </div>
          )}

          {/* hash */}
          {keyType === 'hash' && (
            <div className="space-y-2">
              {hashFields.map((h, i) => (
                <div key={i} className="flex gap-2 items-end">
                  <div className="flex-1">
                    <div className={labelClass}>Field</div>
                    <Input size="sm" placeholder="field 名" value={h.field} onValueChange={(v) => updateHashField(i, { field: v })} />
                  </div>
                  <div className="flex-1">
                    <div className={labelClass}>Value</div>
                    <Input size="sm" placeholder="值" value={h.value} onValueChange={(v) => updateHashField(i, { value: v })} />
                  </div>
                  <Button size="sm" isIconOnly color="danger" variant="flat" onPress={() => removeHashField(i)}>
                    <IconTrash size={14} />
                  </Button>
                </div>
              ))}
              <Button size="sm" variant="flat" onPress={addHashField}><IconPlus size={14} /> 添加 Field</Button>
            </div>
          )}

          {/* list */}
          {keyType === 'list' && (
            <div className="space-y-2">
              {items.map((v, i) => (
                <div key={i} className="flex gap-2 items-end">
                  <div className="flex-1">
                    <div className={labelClass}>元素 [{i}]</div>
                    <Input size="sm" placeholder="元素值" value={v} onValueChange={(val) => updateItem(i, val)} />
                  </div>
                  <Button size="sm" isIconOnly color="danger" variant="flat" onPress={() => removeItem(i)}>
                    <IconTrash size={14} />
                  </Button>
                </div>
              ))}
              <Button size="sm" variant="flat" onPress={addItem}><IconPlus size={14} /> 添加元素</Button>
            </div>
          )}

          {/* set */}
          {keyType === 'set' && (
            <div className="space-y-2">
              {items.map((v, i) => (
                <div key={i} className="flex gap-2 items-end">
                  <div className="flex-1">
                    <div className={labelClass}>成员 {i + 1}</div>
                    <Input size="sm" placeholder="成员值" value={v} onValueChange={(val) => updateItem(i, val)} />
                  </div>
                  <Button size="sm" isIconOnly color="danger" variant="flat" onPress={() => removeItem(i)}>
                    <IconTrash size={14} />
                  </Button>
                </div>
              ))}
              <Button size="sm" variant="flat" onPress={addItem}><IconPlus size={14} /> 添加成员</Button>
            </div>
          )}

          {/* zset */}
          {keyType === 'zset' && (
            <div className="space-y-2">
              {zsetMembers.map((m, i) => (
                <div key={i} className="flex gap-2 items-end">
                  <div className="flex-1">
                    <div className={labelClass}>Member</div>
                    <Input size="sm" placeholder="成员" value={m.member} onValueChange={(v) => updateZsetMember(i, { member: v })} />
                  </div>
                  <div className="w-24">
                    <div className={labelClass}>Score</div>
                    <Input size="sm" placeholder="分数" value={m.score} onValueChange={(v) => updateZsetMember(i, { score: v })} />
                  </div>
                  <Button size="sm" isIconOnly color="danger" variant="flat" onPress={() => removeZsetMember(i)}>
                    <IconTrash size={14} />
                  </Button>
                </div>
              ))}
              <Button size="sm" variant="flat" onPress={addZsetMember}><IconPlus size={14} /> 添加成员</Button>
            </div>
          )}

          {msg && <p className="text-sm text-danger">{msg}</p>}
        </ModalBody>
        <ModalFooter>
          <Button variant="flat" onPress={handleClose}>取消</Button>
          <Button color="primary" onPress={create}>创建</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
