/** Headless dropdown keyboard + collision helpers (framework-free). */
export type DropdownPlacement = 'bottom-end' | 'bottom-start' | 'top-end' | 'top-start'

export function nextFocusIndex(key: string, current: number, count: number): number {
  if (count <= 0) return -1
  if (key === 'ArrowDown') return (current + 1 + count) % count
  if (key === 'ArrowUp') return (current - 1 + count) % count
  if (key === 'Home') return 0
  if (key === 'End') return count - 1
  return current
}

export function shouldCloseOnKey(key: string): boolean {
  return key === 'Escape' || key === 'Tab'
}

/** Flip vertical placement when the panel would clip the viewport. */
export function resolvePlacement(
  preferred: DropdownPlacement,
  triggerRect: { top: number; bottom: number },
  panelHeight: number,
  viewportHeight: number,
): DropdownPlacement {
  const [v, h] = preferred.split('-') as ['top' | 'bottom', 'start' | 'end']
  const spaceBelow = viewportHeight - triggerRect.bottom
  const spaceAbove = triggerRect.top
  if (v === 'bottom' && spaceBelow < panelHeight && spaceAbove > spaceBelow) {
    return `top-${h}` as DropdownPlacement
  }
  if (v === 'top' && spaceAbove < panelHeight && spaceBelow > spaceAbove) {
    return `bottom-${h}` as DropdownPlacement
  }
  return preferred
}

export function placementClass(p: DropdownPlacement): string {
  switch (p) {
    case 'bottom-end': return 'right-0 top-full mt-2 origin-top-right'
    case 'bottom-start': return 'left-0 top-full mt-2 origin-top-left'
    case 'top-end': return 'bottom-full right-0 mb-2 origin-bottom-right'
    case 'top-start': return 'bottom-full left-0 mb-2 origin-bottom-left'
  }
}
