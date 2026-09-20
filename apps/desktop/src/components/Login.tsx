import { useAuth } from '../context/AuthContext'
import Button from './ui/Button'

const Login = () => {
  const { login } = useAuth()

  return (
    <section className='flex bg-accent flex-col items-center justify-center h-screen'>
      <img src="/skydock-logo-resized.png" alt='SkyDock' className='w-40 h-40' />
      <Button intent='primary' size='medium' className='rounded-full' onClick={login}>
        Login
      </Button>
    </section>
  )
}

export default Login
