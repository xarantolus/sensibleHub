import type { DownloadFailure } from '@/api/schema'

export interface FailureView {
  message: string
  songId?: string
}

export function describeFailure(failure: DownloadFailure): FailureView {
  switch (failure.reason) {
    case 'aborted':
      return { message: 'The download was aborted.' }
    case 'tool_failed':
      return { message: 'The downloader failed. The link may be unsupported, private or unavailable.' }
    case 'no_audio':
      return { message: 'No audio could be found at that link.' }
    case 'invalid_audio':
      return { message: 'The downloaded file is not valid audio.' }
    case 'duplicate':
      return failure.songId === undefined
        ? { message: 'This song has already been downloaded.' }
        : { message: 'This song has already been downloaded.', songId: failure.songId }
    case 'internal':
      return { message: 'Something went wrong on the server while processing the download.' }
  }
}
