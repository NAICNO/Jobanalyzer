import type { IconType } from 'react-icons'
import { LuServer, LuCpu, LuCloud, LuHardDrive, LuMonitor, LuNetwork, LuBookOpen, LuGraduationCap } from 'react-icons/lu'
import { GrNodes } from 'react-icons/gr'
import { GiFox } from 'react-icons/gi'

const ICON_REGISTRY: Record<string, IconType> = {
  LuServer,
  LuCpu,
  LuCloud,
  LuHardDrive,
  LuMonitor,
  LuNetwork,
  GrNodes,
  GiFox,
  LuBookOpen,
  LuGraduationCap
}

export const DEFAULT_ICON: IconType = LuServer

export const resolveIcon = (iconKey?: string): IconType => {
  if (!iconKey) return DEFAULT_ICON
  return ICON_REGISTRY[iconKey] ?? DEFAULT_ICON
}
