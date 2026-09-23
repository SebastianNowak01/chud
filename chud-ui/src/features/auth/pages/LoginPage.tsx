import { useState, type FormEvent } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { DrawablyButton, DrawablyInput } from 'drawably/react'
import { Logo } from '@/components/layout/Logo'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Field } from '@/components/ui/Field'
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
    <div className="grid min-h-svh place-items-center p-4">
      <Card className="w-full max-w-[360px]">
        <form className="flex flex-col gap-4" onSubmit={submit}>
          <div className="flex justify-center">
            <Logo size="large" />
          </div>
          <Field label="Username" htmlFor="username">
            <DrawablyInput
              id="username"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </Field>
          <Field label="Password" htmlFor="password">
            <DrawablyInput
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </Field>
          {login.error && <ErrorText>{login.error.message}</ErrorText>}
          <DrawablyButton type="submit" variant="solid" state={buttonState(login.status)}>
            Log in
          </DrawablyButton>
        </form>
      </Card>
    </div>
  )
}
