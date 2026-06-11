import React from 'react';
import { Menu, X } from 'lucide-react';
import { useState } from 'react';

export default function Header() {
	const [isOpen, setIsOpen] = useState(false);

	return (
		<header className="fixed top-0 w-full z-50 bg-black/80 backdrop-blur-md border-b border-slate-800/50">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="flex items-center justify-between h-16">
					{/* Logo */}
					<div className="flex-shrink-0">
						<h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
							Twinspeak
						</h1>
					</div>

					{/* Desktop Navigation */}
					<nav className="hidden md:flex items-center gap-8">
						<a href="#how-it-works" className="text-slate-300 hover:text-white transition">
							How It Works
						</a>
						<a href="#" className="text-slate-300 hover:text-white transition">
							Safety
						</a>
						<a href="#" className="text-slate-300 hover:text-white transition">
							About
						</a>
					</nav>

					{/* CTA Buttons */}
					<div className="hidden md:flex items-center gap-4">
						<a href="https://localhost:4321" className="text-slate-300 hover:text-white transition">
							Sign In
						</a>
						<a
							href="https://localhost:4321"
							className="px-6 py-2 rounded-lg bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 text-white font-semibold transition"
						>
							Download
						</a>
					</div>

					{/* Mobile menu button */}
					<button
						onClick={() => setIsOpen(!isOpen)}
						className="md:hidden inline-flex items-center justify-center p-2 rounded-md text-slate-400 hover:text-white transition"
					>
						{isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
					</button>
				</div>

				{/* Mobile Navigation */}
				{isOpen && (
					<nav className="md:hidden pb-4">
						<a href="#how-it-works" className="block py-2 text-slate-300 hover:text-white transition">
							How It Works
						</a>
						<a href="#" className="block py-2 text-slate-300 hover:text-white transition">
							Safety
						</a>
						<a href="#" className="block py-2 text-slate-300 hover:text-white transition">
							About
						</a>
						<div className="flex gap-4 mt-4">
							<a href="https://localhost:4321" className="flex-1 py-2 text-center rounded-lg bg-slate-800 text-white hover:bg-slate-700 transition">
								Sign In
							</a>
							<a
								href="https://localhost:4321"
								className="flex-1 py-2 text-center rounded-lg bg-gradient-to-r from-blue-600 to-cyan-600 text-white font-semibold transition"
							>
								Download
							</a>
						</div>
					</nav>
				)}
			</div>
		</header>
	);
}
