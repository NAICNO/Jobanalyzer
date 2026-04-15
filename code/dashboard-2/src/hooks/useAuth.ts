import { useContext } from 'react'
import { AuthContext, type AuthContextValue } from '../contexts/AuthContext'

/**
 * Hook to access authentication functionality
 * Must be used within an AuthProvider
 *
 * @returns Auth context with login, logout, and user state management
 */
export const useAuth = (): AuthContextValue => {
  const context = useContext(AuthContext)

  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }

  return context
}
