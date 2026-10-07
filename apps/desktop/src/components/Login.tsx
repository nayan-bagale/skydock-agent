import { useState } from 'react'

import { useAuth } from '../context/AuthContext'
import Button from './ui/Button'

const Login = () => {
  const { login } = useAuth()
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const handleLogin = async () => {
    setError('')
    setBusy(true)
    try {
      await login()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not start sign-in')
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className='flex bg-accent flex-col items-center justify-center h-screen gap-3'>
      <img src="/skydock-logo-resized.png" alt='SkyDock' className='w-40 h-40' />
      {error ? <p className="text-sm text-destructive px-6 text-center">{error}</p> : null}
      <Button
        intent='primary'
        size='medium'
        className='rounded-full'
        disabled={busy}
        onClick={() => void handleLogin()}
      >
        {busy ? 'Opening browser…' : 'Login'}
      </Button>
    </section>
  )
}

export default Login
