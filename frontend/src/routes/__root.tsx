import { HeadContent, Scripts, createRootRouteWithContext } from '@tanstack/react-router'
import { Toaster } from '@/components/ui/sonner'
import { QueryClientAtomProvider } from 'jotai-tanstack-query/react'

import appCss from '../styles.css?url'
import { checkAuthServerFn } from './auth'
import { QueryClient } from '@tanstack/query-core'

const THEME_INIT_SCRIPT = `(function(){try{var stored=window.localStorage.getItem('theme');var mode=(stored==='light'||stored==='dark'||stored==='auto')?stored:'auto';var prefersDark=window.matchMedia('(prefers-color-scheme: dark)').matches;var resolved=mode==='auto'?(prefersDark?'dark':'light'):mode;var root=document.documentElement;root.classList.remove('light','dark');root.classList.add(resolved);if(mode==='auto'){root.removeAttribute('data-theme')}else{root.setAttribute('data-theme',mode)}root.style.colorScheme=resolved;}catch(e){}})();`

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: 0,
        },
        mutations: {
            retry: 0,
        },
    },
})

export const Route = createRootRouteWithContext()({
    head: () => ({
        meta: [
            {
                charSet: 'utf-8',
            },
            {
                name: 'viewport',
                content: 'width=device-width, initial-scale=1',
            },
            {
                title: 'TanStack Start Starter',
            },
        ],
        links: [
            {
                rel: 'stylesheet',
                href: appCss,
            },
        ],
    }),
    shellComponent: RootDocument,
    beforeLoad: async ({ location }) => {
        if (location.pathname.startsWith('/auth')) {
            return
        }
        const auth = await checkAuthServerFn()
        if (!auth.session) throw Route.redirect({ to: '/auth' })
    },
})

function RootDocument({ children }: { children: React.ReactNode }) {
    return (
        <QueryClientAtomProvider client={queryClient}>
            <html lang="en" suppressHydrationWarning>
                <head>
                    <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
                    <HeadContent />
                </head>
                <body className="h-dvh">
                    {children}
                    <Toaster />
                    <Scripts />
                </body>
            </html>
        </QueryClientAtomProvider>
    )
}
