import { Button, Switch } from '@nextui-org/react'
import { IconLogout, IconSun, IconMoon, IconBrandGithub } from '@tabler/icons-react'

export default function Header({ dark, onToggleDark, authed, onLogout }: {
  dark: boolean; onToggleDark: (v: boolean) => void
  authed: boolean; onLogout: () => void
}) {
  return (
    <header className="flex items-center justify-between px-6 py-3 border-b border-divider bg-background">
      <h1 className="text-lg font-bold flex items-center gap-2">
        <img src="/favicon.png" alt="Godis" className="w-10 h-10" /> Godis Admin
      </h1>
      <div className="flex items-center gap-3">
        <a href="https://github.com/LanceHE6/godis" target="_blank" rel="noopener"
           className="text-default-500 hover:text-default-700">
          <IconBrandGithub size={20} />
        </a>
        <Switch size="sm" color="warning" isSelected={dark} onValueChange={onToggleDark}
          thumbIcon={dark ? <IconMoon size={12} /> : <IconSun size={12} />} />
        {authed && (
          <Button size="sm" variant="flat" color="danger" onPress={onLogout}
            startContent={<IconLogout size={14} />}>退出</Button>
        )}
      </div>
    </header>
  )
}
