import { describe, expect, it } from 'vitest'

import type { DownloadFailure } from '@/api/schema'

import { describeFailure } from './downloadErrors'

const reasons: DownloadFailure['reason'][] = [
  'aborted',
  'tool_failed',
  'no_audio',
  'invalid_audio',
  'duplicate',
  'internal',
]

describe('describeFailure', () => {
  it('has a message for every reason', () => {
    for (const reason of reasons) {
      expect(describeFailure({ reason, message: 'x' }).message).not.toBe('')
    }
  })

  it('links the existing song for duplicates', () => {
    expect(describeFailure({ reason: 'duplicate', message: 'x', songId: 'abc' }).songId).toBe('abc')
  })

  it('omits the link when a duplicate has no song', () => {
    expect(describeFailure({ reason: 'duplicate', message: 'x' })).not.toHaveProperty('songId')
  })

  it('does not link other reasons', () => {
    expect(describeFailure({ reason: 'aborted', message: 'x', songId: 'abc' }).songId).toBeUndefined()
  })
})
