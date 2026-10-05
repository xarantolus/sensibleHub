const destinations: ReadonlyMap<string, string> = new Map([
  ['n', '/add'],
  ['s', '/songs'],
  ['a', '/artists'],
  ['y', '/years'],
  ['i', '/incomplete'],
  ['u', '/unsynced'],
  ['e', '/edits'],
])

export function shortcutDestination(key: string): string | undefined {
  return destinations.get(key)
}

export function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false
  }
  return target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)
}

export function isInteractiveTarget(target: EventTarget | null): boolean {
  return target instanceof HTMLElement && ['BUTTON', 'A', 'AUDIO', 'VIDEO', 'SUMMARY'].includes(target.tagName)
}
