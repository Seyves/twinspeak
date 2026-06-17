import { themes, type Theme } from '@/definitions/chat'
import { useAtom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'
import { createContext, useEffect } from 'react'

export const localThemeAtom = atomWithStorage<Theme>('theme', themes.system)

type ThemeProviderProps = {
    children: React.ReactNode
}

const ThemeProviderContext = createContext(null)

export function ThemeProvider({ children }: ThemeProviderProps) {
    const [theme] = useAtom(localThemeAtom)

    useEffect(() => {
        const root = window.document.documentElement

        root.classList.remove('light', 'dark')

        if (theme === 'system') {
            root.classList.add(resolveSystemTheme())
            return
        }

        root.classList.add(theme)
    }, [theme])

    return <ThemeProviderContext.Provider value={null}>{children}</ThemeProviderContext.Provider>
}

export function resolveSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? themes.dark : themes.light
}
