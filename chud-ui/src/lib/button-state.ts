type MutationStatus = 'idle' | 'pending' | 'error' | 'success'
type ButtonState = 'idle' | 'loading' | 'error' | 'success'

// Maps a react-query mutation status to a DrawablyButton state.
export const buttonState = (status: MutationStatus): ButtonState =>
  status === 'pending' ? 'loading' : status
