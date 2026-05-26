import { useEffect, useState } from 'react'

export type Preferences = {
    chatFontSize: 'sm' | 'md' | 'lg'
}

const STORAGE_KEY = 'twinspeak-preferences'
const DEFAULT_PREFERENCES: Preferences = {
    chatFontSize: 'md',
}

export function usePreferences() {
    const [preferences, setPreferences] = useState<Preferences>(DEFAULT_PREFERENCES)
    const [isLoaded, setIsLoaded] = useState(false)

    useEffect(() => {
        try {
            const stored = localStorage.getItem(STORAGE_KEY)
            if (stored) {
                const parsed = JSON.parse(stored)
                setPreferences({ ...DEFAULT_PREFERENCES, ...parsed })
            }
        } catch (e) {
            console.error('Failed to load preferences:', e)
        }
        setIsLoaded(true)
    }, [])

    const updatePreferences = (updates: Partial<Preferences>) => {
        const newPreferences = { ...preferences, ...updates }
        setPreferences(newPreferences)
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(newPreferences))
        } catch (e) {
            console.error('Failed to save preferences:', e)
        }
    }

    return { preferences, updatePreferences, isLoaded }
}
