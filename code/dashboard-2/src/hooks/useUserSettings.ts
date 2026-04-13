import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getUserSettingsOptions, getUserSettingsQueryKey, putUserSettingsMutation } from '../client/@tanstack/react-query.gen'
import type { Client } from '../client/client/types.gen'
import type { UserSettings } from '../types/userSettings'
import { createDefaultSettings } from '../types/userSettings'

interface UseUserSettingsOptions {
  client: Client | null
  enabled?: boolean
}

export const useUserSettings = ({ client, enabled = true }: UseUserSettingsOptions) => {
  const queryClient = useQueryClient()

  const query = useQuery({
    ...getUserSettingsOptions({
      client: client || undefined,
    }),
    enabled: enabled && !!client,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    select: (data): UserSettings => {
      if (!data) return createDefaultSettings()
      // Backend returns { user, settings, time_modified } envelope
      // where settings is a JSON string
      const envelope = data as Record<string, unknown>
      const raw = envelope.settings ?? data
      let parsed: unknown = raw
      if (typeof parsed === 'string') {
        try {
          parsed = JSON.parse(parsed)
        } catch {
          return createDefaultSettings()
        }
      }
      if (!parsed || typeof parsed !== 'object') return createDefaultSettings()
      const settings = parsed as UserSettings
      if (settings.version !== 1) return createDefaultSettings()
      return settings
    },
  })

  const mutation = useMutation({
    ...putUserSettingsMutation({
      client: client || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getUserSettingsQueryKey({ client: client || undefined }),
      })
    },
  })

  const saveSettings = (settings: UserSettings) => {
    if (!client) return
    mutation.mutate({
      query: { settings: JSON.stringify(settings) },
      client,
    })
  }

  const updateSettings = (updater: (current: UserSettings) => UserSettings) => {
    const current = query.data ?? createDefaultSettings()
    saveSettings(updater(current))
  }

  return {
    settings: query.data ?? createDefaultSettings(),
    isLoading: query.isLoading,
    isError: query.isError,
    isSaving: mutation.isPending,
    saveError: mutation.error,
    saveSettings,
    updateSettings,
  }
}
