import { motion } from 'motion/react'

export default function Loader() {
    return (
        <motion.div
            initial={false}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="bg-background z-10 absolute w-full h-full flex items-center justify-center"
        >
            <div className="grid w-10 aspect-square">
                <div className="col-start-1 row-start-1 bg-accent [clip-path:polygon(0_0,50%_50%,0_100%)] animate-[l11_2s_infinite]" />
                <div
                    className="col-start-1 row-start-1 bg-primary [clip-path:polygon(0_0,50%_50%,0_100%)] animate-[l11_2s_infinite] [animation-delay:-1.5s]"
                    style={{ '--s': '90deg' } as React.CSSProperties}
                />
            </div>
        </motion.div>
    )
}
