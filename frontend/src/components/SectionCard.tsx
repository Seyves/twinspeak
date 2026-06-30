export default function SectionCard({
    label,
    children,
}: {
    label: string
    children: React.ReactNode
}) {
    return (
        <div className="rounded-2xl border border-border/50 bg-card overflow-hidden mb-4">
            <div className="px-4 py-3 border-b border-border/50">
                <p className="text-sm text-muted-foreground">
                    {label}
                </p>
            </div>
            <div className="px-4 py-4">{children}</div>
        </div>
    )
}
