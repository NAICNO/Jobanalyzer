export interface SidebarSubItem {
  text: string
  path: string
  matches: string
}

export interface SidebarItem {
  type: 'link' | 'separator'
  path?: string
  matches?: string
  text?: string
  icon?: any,
  subItems?: SidebarSubItem[]
}
