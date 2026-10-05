import { useNotificationProgrammatic } from '@oruga-ui/oruga-next'

import { ApiError } from '@/api/client'

type Variant = 'success' | 'danger' | 'warning' | 'info'

export function notify(message: string, variant: Variant = 'info'): void {
  useNotificationProgrammatic().open({
    message,
    variant,
    position: 'bottom-right',
    duration: variant === 'danger' ? 8000 : 3000,
    closable: true,
  })
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    return err.message
  }
  if (err instanceof Error) {
    return err.message
  }
  return 'Something went wrong'
}

/** Shows a failed action to the user; use as a mutation's onError. */
export function notifyError(err: unknown, action?: string): void {
  const msg = errorMessage(err)
  notify(action === undefined ? msg : `${action}: ${msg}`, 'danger')
}
