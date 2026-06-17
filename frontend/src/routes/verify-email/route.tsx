import { createFileRoute, Outlet } from '@tanstack/react-router'
import { checkAuthServerFn } from '@/routes/auth'

export const Route = createFileRoute('/verify-email')({
    beforeLoad: async () => {
        const auth = await checkAuthServerFn()
        if (auth.emailVerified) {
            throw Route.redirect({ to: '/' })
        }
    },
    component: VerifyEmailLayout,
})

function VerifyEmailLayout() {
    return (
        <div className="h-full w-full bg-[#f1f2f7] dark:bg-[#060607]">
            <Outlet />
        </div>
    )
}
