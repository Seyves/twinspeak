import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/settings/')({
    beforeLoad: () => {
        throw Route.redirect({
            to: "/settings/preferences",
        })
    },
})
