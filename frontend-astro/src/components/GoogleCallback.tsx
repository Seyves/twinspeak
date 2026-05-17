import { useEffect } from 'react'
import { googleProcessCallback } from '@/api/auth'

export default function GoogleCallback() {
    async function processCallback() {
        const queryString = window.location.search
        const query = new URLSearchParams(queryString)

        const code = query.get('code') as string
        const state = query.get('state') as string

        try {
            await googleProcessCallback(code, state)
            window.location.replace('/')
        } catch (error) {
            console.error('OAuth callback failed:', error)
        }
    }

    useEffect(() => {
        processCallback()
    }, [])

    return <div></div>
}
