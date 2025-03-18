import { CLUSTER_INFO } from '../Constants.ts'
import moment from 'moment'
import { Cluster, Subcluster } from '../types'

export const findCluster = (clusterName?: string | null): Cluster | null => {
  return clusterName && CLUSTER_INFO[clusterName] ? CLUSTER_INFO[clusterName] : null
}

export const findSubcluster = (clusterName?: string | null, subclusterName?: string | null): {
  cluster: Cluster,
  subcluster: Subcluster
} | null => {
  const cluster = findCluster(clusterName)
  const subcluster = cluster?.subclusters.find(subcluster => subcluster.name === subclusterName)
  return cluster && subcluster ? {cluster, subcluster} : null
}

export const breakText = (text: string): string => {
  // Don't break at spaces that exist, but at commas.  Generally this has the effect of
  // keeping duration and timestamp fields together and breaking the command field apart.
  //
  // TODO: This is not ideal, b/c it breaks within node name ranges, we can fix that.
  return text.replaceAll(' ', '\xA0').replaceAll(',', ', ')
}

export const formatToUtcDateTimeString = (date: Date) => {
  return moment(date).utc().format('YYYY-MM-DDTHH:mm[Z]')
}

export const toPercentage = (value: number) => {
  return (value / 100).toFixed(1)
}

export const parseDateString = (dateString: string, format?: string) => {
  return moment.utc(dateString, format || 'YYYY-MM-DD HH:mm').toDate()
}

// Job Query Form - Custom Yup validation for date format
export const validateDateFormat = (value: string | undefined) => {
  if (!value) return true // Field is optional, so it's valid if empty

  const absoluteDateFormat = 'YYYY-MM-DD'
  const relativeDateRegex = /^\d+[dw]$/ // Matches "1w", "2d", etc.

  return !!(moment(value, absoluteDateFormat, true).isValid() || relativeDateRegex.test(value))

}

// Job Query Form - Parse relative dates to absolute dates based on current date
export const parseRelativeDate = (value: string) => {
  const relativeDateRegex = /^(\d+)([dw])$/
  const match = value.match(relativeDateRegex)
  if (match) {
    const amount = parseInt(match[1], 10)
    const unit = match[2] === 'd' ? 'days' : 'weeks'
    return moment().add(amount, unit)
  }
  return moment(value, 'YYYY-MM-DD', true)
}

export const reformatHostDescriptions = (description: string): string => {
  if (!description) return ''

  const counts: Record<string, number> = description
    .split('|||')
    .reduce((acc, desc) => {
      acc[desc] = (acc[desc] || 0) + 1
      return acc
    }, {} as Record<string, number>)

  return Object.entries(counts)
    .sort(([descA, countA], [descB, countB]) =>
      countB - countA || descA.localeCompare(descB))
    .map(([desc, count]) => `${count}x ${desc}`)
    .join('\n')
}

export const generateTolRainbowColors = (numColors: number): string[] => {
  const colors: string[] = []
  const saturation = 0.60 // High saturation for vibrant colors
  const lightness = 0.60  // Lightness adjusted for visibility on both light and dark backgrounds

  for (let i = 0; i < numColors; i++) {
    const hue = (i / numColors) * 360 // Distribute hues evenly
    colors.push(hslToHex(hue, saturation, lightness))
  }
  return colors
}

/**
 * Converts HSL color values to a hexadecimal string.
 * @param h - Hue (0-360)
 * @param s - Saturation (0-1)
 * @param l - Lightness (0-1)
 * @returns Hexadecimal color string.
 */
const hslToHex = (h: number, s: number, l: number): string => {
  l = l < 0 ? 0 : l > 1 ? 1 : l
  s = s < 0 ? 0 : s > 1 ? 1 : s
  h = h % 360
  const c = (1 - Math.abs(2 * l - 1)) * s
  const x = c * (1 - Math.abs((h / 60) % 2 - 1))
  const m = l - c / 2
  let r = 0, g = 0, b = 0

  if (h < 60) {
    r = c
    g = x
    b = 0
  } else if (h < 120) {
    r = x
    g = c
    b = 0
  } else if (h < 180) {
    r = 0
    g = c
    b = x
  } else if (h < 240) {
    r = 0
    g = x
    b = c
  } else if (h < 300) {
    r = x
    g = 0
    b = c
  } else {
    r = c
    g = 0
    b = x
  }

  const toHex = (n: number) => {
    return Math.round((n + m) * 255).toString(16).padStart(2, '0')
  }

  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}
