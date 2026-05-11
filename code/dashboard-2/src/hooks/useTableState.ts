import { useCallback, useRef, useEffect } from 'react'
import type { GridApi } from 'ag-grid-community'
import type { AgGridTableState, UserSettings } from '../types/userSettings'

type TableKey = 'jobs' | 'nodes' | 'jobQuery'

interface UseTableStateOptions {
  tableKey: TableKey
  settings: UserSettings
  updateSettings: (updater: (current: UserSettings) => UserSettings) => void
  enabled?: boolean
}

const DEBOUNCE_MS = 500

export const useTableState = ({
  tableKey,
  settings,
  updateSettings,
  enabled = true,
}: UseTableStateOptions) => {
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const apiRef = useRef<GridApi | null>(null)
  const hasRestoredRef = useRef(false)

  const savedState = settings.tables?.[tableKey]

  const save = useCallback(
    (api: GridApi) => {
      if (!enabled) return

      if (debounceRef.current) clearTimeout(debounceRef.current)

      debounceRef.current = setTimeout(() => {
        const state: AgGridTableState = {
          columnState: api.getColumnState(),
          filterModel: api.getFilterModel(),
          pageSize: api.paginationGetPageSize(),
        }

        updateSettings((current) => ({
          ...current,
          tables: {
            ...current.tables,
            [tableKey]: state,
          },
        }))
      }, DEBOUNCE_MS)
    },
    [enabled, tableKey, updateSettings],
  )

  // Cleanup debounce on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [])

  const onGridReady = useCallback(
    (api: GridApi) => {
      apiRef.current = api

      // Restore saved state on first load
      if (savedState && !hasRestoredRef.current && enabled) {
        hasRestoredRef.current = true

        if (savedState.columnState) {
          api.applyColumnState({ state: savedState.columnState, applyOrder: true })
        }
        if (savedState.filterModel) {
          api.setFilterModel(savedState.filterModel)
        }
        if (savedState.pageSize) {
          api.setGridOption('paginationPageSize', savedState.pageSize)
        }
      }
    },
    [savedState, enabled],
  )

  const onStateChanged = useCallback(
    (api: GridApi) => {
      save(api)
    },
    [save],
  )

  return {
    onGridReady,
    onStateChanged,
    savedPageSize: savedState?.pageSize,
  }
}
