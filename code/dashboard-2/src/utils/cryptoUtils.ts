/**
 * Generates a cryptographically secure random string for nonce/state
 * @param length - Length of the random string (in bytes, will be doubled in hex output)
 * @returns Random hex string
 */
export const generateRandomString = (length: number = 32): string => {
  const array = new Uint8Array(length)
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * Decodes JWT token to extract claims (for debugging)
 * @param token - JWT token string
 * @returns Decoded token payload or null if decoding fails
 */
export const decodeJwt = (token: string): Record<string, unknown> | null => {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(atob(base64).split('').map((c) => {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
    }).join(''))
    return JSON.parse(jsonPayload)
  } catch (e) {
    console.error('Failed to decode JWT:', e)
    return null
  }
}
