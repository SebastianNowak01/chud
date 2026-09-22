import { useState, type FormEvent } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { DrawablyButton, DrawablyCard, DrawablyInput } from 'drawably/react'
import { AuthApi } from '@/features/auth/auth-api'
import { storeToken } from '@/features/auth/lib/token'
import { buttonState } from '@/lib/button-state'

export function LoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  const login = useMutation({
    mutationFn: AuthApi.login,
    onSuccess: ({ token }) => {
      storeToken(token)
      void navigate({ to: '/' })
    },
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    login.mutate({ username, password })
  }

  return (
    <div className="login">
      <DrawablyCard className="card">
        <form className="stack" onSubmit={submit}>
          <h1>chud</h1>
          <div className="field">
            <label htmlFor="username">Username</label>
            <DrawablyInput
              id="username"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>
          <div className="field">
            <label htmlFor="password">Password</label>
            <DrawablyInput
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          {login.error && <p className="error">{login.error.message}</p>}
          <DrawablyButton type="submit" variant="solid" state={buttonState(login.status)}>
            Log in
          </DrawablyButton>
        </form>
      </DrawablyCard>
    </div>
  )
}
