interface SidebarItem {
  type: 'link' | 'separator'
  path?: string
  matches?: string
  text?: string
  icon?: any
}
