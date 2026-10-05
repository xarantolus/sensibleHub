import { describe, expect, it } from 'vitest'

import { isInteractiveTarget, isTypingTarget, shortcutDestination } from './shortcutKeys'

describe('shortcutDestination', () => {
  it('maps the legacy navigation keys', () => {
    expect(shortcutDestination('n')).toBe('/add')
    expect(shortcutDestination('e')).toBe('/edits')
    expect(shortcutDestination('u')).toBe('/unsynced')
  })

  it('ignores other keys', () => {
    expect(shortcutDestination('q')).toBeUndefined()
    expect(shortcutDestination('toString')).toBeUndefined()
  })
})

describe('target classification', () => {
  it('treats text fields as typing targets', () => {
    expect(isTypingTarget(document.createElement('input'))).toBe(true)
    expect(isTypingTarget(document.createElement('textarea'))).toBe(true)
    expect(isTypingTarget(document.createElement('div'))).toBe(false)
    expect(isTypingTarget(null)).toBe(false)
  })

  it('treats buttons and links as interactive', () => {
    expect(isInteractiveTarget(document.createElement('button'))).toBe(true)
    expect(isInteractiveTarget(document.createElement('p'))).toBe(false)
  })
})
