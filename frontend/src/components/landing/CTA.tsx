import React from 'react';
import { ArrowRight, Smartphone } from 'lucide-react';

export default function CTA() {
	return (
		<section className="relative py-24 md:py-32 bg-gradient-to-b from-black to-slate-900 overflow-hidden">
			{/* Animated background */}
			<div className="absolute inset-0 opacity-20">
				<div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl"></div>
				<div className="absolute top-1/3 right-1/4 w-96 h-96 bg-cyan-500 rounded-full mix-blend-multiply filter blur-3xl opacity-30"></div>
			</div>

			<div className="relative z-10 max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
				<h2 className="text-5xl md:text-6xl font-bold mb-6">
					<span className="bg-gradient-to-r from-white via-blue-200 to-cyan-300 bg-clip-text text-transparent">
						Stop Traveling Alone
					</span>
				</h2>

				<p className="text-xl text-slate-300 mb-8 max-w-2xl mx-auto">
					Download Twinspeak and start connecting with locals in any city you visit. Your next adventure is just one conversation away.
				</p>

				<div className="flex flex-col sm:flex-row gap-4 justify-center mb-12">
					<a
						href="https://localhost:4321"
						className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-lg bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 text-white font-semibold transition-all transform hover:scale-105"
					>
						<Smartphone className="w-5 h-5" />
						Download Now
					</a>
					<a
						href="#how-it-works"
						className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-lg bg-slate-800 hover:bg-slate-700 text-white font-semibold border border-slate-700 transition-all"
					>
						Learn More
					</a>
				</div>

				{/* App store badges would go here */}
				<p className="text-slate-400 text-sm">Available on iOS and Android</p>
			</div>
		</section>
	);
}
