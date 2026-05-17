import { useEffect, useRef } from 'react'

export function AnimatedBackground() {
    const canvasRef = useRef<HTMLCanvasElement>(null)

    useEffect(() => {
        const canvas = canvasRef.current
        if (!canvas) return

        const ctx = canvas.getContext('2d')
        if (!ctx) return

        // Set canvas size
        const resizeCanvas = () => {
            canvas.width = window.innerWidth
            canvas.height = window.innerHeight
        }
        resizeCanvas()
        window.addEventListener('resize', resizeCanvas)

        // Particle system
        interface Particle {
            x: number
            y: number
            vx: number
            vy: number
            size: number
            opacity: number
            color: string
        }

        const particles: Particle[] = []
        const particleCount = 50

        // Initialize particles
        for (let i = 0; i < particleCount; i++) {
            particles.push({
                x: Math.random() * canvas.width,
                y: Math.random() * canvas.height,
                vx: (Math.random() - 0.5) * 1,
                vy: (Math.random() - 0.5) * 1,
                size: Math.random() * 2 + 1,
                opacity: Math.random() * 0.5 + 0.2,
                color: ['oklch(0.7247 0.1424 266.1732)', 'oklch(0.567 0.1282 279.41)'][
                    Math.floor(Math.random() * 2)
                ],
            })
        }

        let animationFrameId: number
        const animate = () => {
            // Clear with gradient background
            const gradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height)
            gradient.addColorStop(0, 'oklch(0.1076 0.005 280)')
            gradient.addColorStop(1, 'oklch(0.13 0.006 280)')
            ctx.fillStyle = gradient
            ctx.fillRect(0, 0, canvas.width, canvas.height)

            // Update and draw particles
            particles.forEach((particle) => {
                particle.x += particle.vx
                particle.y += particle.vy

                // Wrap around edges
                if (particle.x < 0) particle.x = canvas.width
                if (particle.x > canvas.width) particle.x = 0
                if (particle.y < 0) particle.y = canvas.height
                if (particle.y > canvas.height) particle.y = 0

                // Draw particle with glow
                ctx.shadowColor = particle.color
                ctx.shadowBlur = 15
                ctx.fillStyle = particle.color
                ctx.globalAlpha = particle.opacity
                ctx.beginPath()
                ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2)
                ctx.fill()
                ctx.globalAlpha = 1
            })

            // Draw connecting lines between nearby particles
            ctx.strokeStyle = 'oklch(0.7247 0.1424 266.1732)'
            ctx.globalAlpha = 0.1
            ctx.lineWidth = 1

            for (let i = 0; i < particles.length; i++) {
                for (let j = i + 1; j < particles.length; j++) {
                    const dx = particles[i].x - particles[j].x
                    const dy = particles[i].y - particles[j].y
                    const distance = Math.sqrt(dx * dx + dy * dy)

                    if (distance < 150) {
                        ctx.beginPath()
                        ctx.moveTo(particles[i].x, particles[i].y)
                        ctx.lineTo(particles[j].x, particles[j].y)
                        ctx.stroke()
                    }
                }
            }

            ctx.globalAlpha = 1
            animationFrameId = requestAnimationFrame(animate)
        }

        animate()

        return () => {
            window.removeEventListener('resize', resizeCanvas)
            cancelAnimationFrame(animationFrameId)
        }
    }, [])

    return <canvas ref={canvasRef} className="fixed inset-0 -z-10 pointer-events-none" />
}
