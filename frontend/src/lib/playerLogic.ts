import type { Playback } from '@/api/schema'

export const refillThreshold = 3
export const refillCount = 5
export const excludeRecent = 50
export const restartThreshold = 3

/** The song that suggestions should follow: the last queued one, else the current one. */
export function refillSeed(current: string | undefined, queue: readonly string[]): string | undefined {
  return queue.at(-1) ?? current
}

export function needsRefill(queueLength: number): boolean {
  return queueLength < refillThreshold
}

/** Songs the server should not suggest: what is playing, queued, or played recently. */
export function refillExclude(current: string | undefined, queue: readonly string[], history: readonly string[]): string[] {
  const ids = new Set([...history.slice(-excludeRecent), ...queue])
  if (current !== undefined) {
    ids.add(current)
  }
  return [...ids]
}

export type AudioFailure = 'network' | 'unsupported' | 'aborted' | 'unknown'

/** Classifies an HTMLMediaElement error code (MediaError.MEDIA_ERR_*). */
export function classifyMediaError(code: number | undefined): AudioFailure {
  switch (code) {
    case 1:
      return 'aborted'
    case 2:
      return 'network'
    case 3:
    case 4:
      return 'unsupported'
    default:
      return 'unknown'
  }
}

export type FailureAction = 'retry-later' | 'try-mp3' | 'skip' | 'ignore'

/**
 * What to do when playback fails: network trouble keeps the song and retries
 * (offline on a phone must not burn through the queue), an undecodable original
 * falls back to the server-generated MP3 once, anything else skips the song.
 */
export function failureAction(failure: AudioFailure, triedMp3: boolean): FailureAction {
  switch (failure) {
    case 'aborted':
      return 'ignore'
    case 'network':
      return 'retry-later'
    case 'unsupported':
      return triedMp3 ? 'skip' : 'try-mp3'
    case 'unknown':
      return triedMp3 ? 'skip' : 'try-mp3'
  }
}

export function retryDelay(attempt: number): number {
  return Math.min(1000 * 2 ** attempt, 30_000)
}

/** Position within the trimmed song, from a position within the audio file. */
export function songTime(fileTime: number, playback: Playback): number {
  return Math.min(Math.max(fileTime - playback.start, 0), playback.end - playback.start)
}

export function fileTime(songTime: number, playback: Playback): number {
  return Math.min(Math.max(playback.start + songTime, playback.start), playback.end)
}

export const targetLoudness = -16

/**
 * Volume factor that brings a song to the target loudness. The audio element
 * cannot amplify, so quiet songs play at full volume.
 */
export function loudnessGain(loudness: number | undefined): number {
  if (loudness === undefined || !Number.isFinite(loudness)) {
    return 1
  }
  return Math.min(1, 10 ** ((targetLoudness - loudness) / 20))
}
