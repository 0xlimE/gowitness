/**
 * Get a cookie value by name
 */
export function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift() || null;
  }
  return null;
}

/**
 * Set a cookie
 */
export function setCookie(name: string, value: string, maxAge: number = 360): void {
  document.cookie = `${name}=${value}; path=/; max-age=${maxAge}`;
}

/**
 * Check if user has seen the intro
 */
export function hasSeenIntro(): boolean {
  return getCookie('has_seen_intro') === 'true';
}
