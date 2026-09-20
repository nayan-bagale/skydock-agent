
import { DesktopUiProvider } from './context/DesktopUiContext'
import Header from './components/Header'
import MainContent from './components/MainContent'
import Sidebar from './components/Sidebar'
function App() {
  return (
    <DesktopUiProvider>
      <main className="grid h-screen w-screen grid-cols-[245px_minmax(0,1fr)] overflow-hidden bg-background text-foreground max-[800px]:grid-cols-1 max-[800px]:overflow-auto">
        <Sidebar />
        <section className="min-w-0">
          <Header />
          <MainContent />
        </section>
      </main>
    </DesktopUiProvider>
  )
}

export default App
