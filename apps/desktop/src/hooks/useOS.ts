import { useState, useEffect } from 'react'

export type OSType = 'macOS' | 'Windows' | 'Linux' | 'unknown'

export function useOS() {
  const [os, setOS] = useState<OSType>('unknown')

  useEffect(() => {
    if (window.ipcRenderer && typeof window.ipcRenderer.getOS === 'function') {
      const platform = window.ipcRenderer.getOS()
      
      if (platform === 'darwin') setOS('macOS')
      else if (platform === 'win32') setOS('Windows')
      else if (platform === 'linux') setOS('Linux')
    }
  }, [])

  return {
    os,
    isMac: os === 'macOS',
    isWindows: os === 'Windows',
    isLinux: os === 'Linux',
  }
}
