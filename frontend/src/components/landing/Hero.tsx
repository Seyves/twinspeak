import React, { useState, useEffect } from 'react';
import { ArrowRight, Globe } from 'lucide-react';

export default function Hero() {
	const [counter, setCounter] = useState(42000);

	useEffect(() => {
		const interval = setInterval(() => {
			setCounter(prev => prev + Math.floor(Math.random() * 10));
		}, 3000);
		return () => clearInterval(interval);
	}, []);

	return (
		<div className="relative min-h-screen bg-gradient-to-b from-black via-slate-900 to-black overflow-hidden pt-32 pb-32 flex items-center">
			{/* Animated background */}
			<div className="absolute inset-0 opacity-30">
				<div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl opacity-20"></div>
				<div className="absolute top-1/3 right-1/4 w-96 h-96 bg-cyan-500 rounded-full mix-blend-multiply filter blur-3xl opacity-20"></div>
			</div>

			<div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="text-center space-y-8">
					{/* Badge */}
					<div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-slate-800/50 border border-slate-700 hover:border-slate-600 transition">
						<Globe className="w-4 h-4 text-cyan-400" />
						<span className="text-sm text-slate-300">Connect with travelers worldwide</span>
					</div>

					{/* Main headline */}
					<h1 className="text-6xl md:text-7xl font-bold tracking-tight">
						<span className="bg-gradient-to-r from-white via-blue-200 to-cyan-300 bg-clip-text text-transparent">
							Talk to anyone,<br />anywhere.
						</span>
					</h1>

					{/* Subheading */}
					<p className="text-xl md:text-2xl text-slate-300 max-w-3xl mx-auto leading-relaxed">
						Real-time conversation with instant translation. Chat with locals while traveling—no language barriers.
					</p>

					{/* CTA Buttons */}
					<div className="flex flex-col sm:flex-row gap-4 justify-center pt-8">
						<a
							href="https://localhost:4321"
							className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-lg bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 text-white font-semibold transition-all transform hover:scale-105"
						>
							Start Chatting
							<ArrowRight className="w-5 h-5" />
						</a>
						<a
							href="#how-it-works"
							className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-lg bg-slate-800 hover:bg-slate-700 text-white font-semibold border border-slate-700 transition-all"
						>
							See How It Works
						</a>
					</div>

					{/* Social proof */}
					<div className="grid grid-cols-1 md:grid-cols-3 gap-8 pt-16 border-t border-slate-800">
						<div className="space-y-2">
							<p className="text-3xl font-bold text-cyan-400">
								{counter.toLocaleString()}
							</p>
							<p className="text-slate-400">Active Conversations</p>
						</div>
						<div className="space-y-2">
							<p className="text-3xl font-bold text-cyan-400">180+</p>
							<p className="text-slate-400">Countries</p>
						</div>
						<div className="space-y-2">
							<p className="text-3xl font-bold text-cyan-400">50+</p>
							<p className="text-slate-400">Languages</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
