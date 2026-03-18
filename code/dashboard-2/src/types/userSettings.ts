export interface UserSettings {
  version: 1
  jobQueryPresets?: JobQueryPreset[]
  tables?: {
    jobs?: AgGridTableState
    nodes?: AgGridTableState
    jobQuery?: AgGridTableState
  }
}

export interface JobQueryPreset {
  id: string
  name: string
  createdAt: string
  filters: JobQueryPresetFilters
}

export interface JobQueryPresetFilters {
  user?: string
  userId?: string
  jobId?: string
  states?: string
  startAfter?: string
  startBefore?: string
  endAfter?: string
  endBefore?: string
  submitAfter?: string
  submitBefore?: string
  minDuration?: string
  maxDuration?: string
}

export interface AgGridTableState {
  pageSize?: number
  columnState?: unknown[]
  filterModel?: unknown
}

export const MAX_PRESETS = 20

export const createDefaultSettings = (): UserSettings => ({
  version: 1,
  jobQueryPresets: [],
  tables: {},
})
