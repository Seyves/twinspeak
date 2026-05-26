import React from 'react';
import { Star } from 'lucide-react';

const testimonials = [
	{
		name: 'Sarah',
		location: 'Tourist from USA',
		text: 'I met a local in Bangkok and we talked for hours. Without Twinspeak, we wouldn\'t have been able to communicate. Best experience ever!',
		rating: 5,
	},
	{
		name: 'Marco',
		location: 'Local in Rome',
		text: 'I love helping travelers understand my city. Twinspeak makes it so easy to have real conversations without struggling with translation apps.',
		rating: 5,
	},
	{
		name: 'Yuki',
		location: 'Tourist from Japan',
		text: 'Not just a translation app—it\'s a way to make genuine connections. I made friends I\'ll visit again.',
		rating: 5,
	},
	{
		name: 'Diego',
		location: 'Local in Buenos Aires',
		text: 'The real-time translation is seamless. I\'ve met travelers from 15 different countries through this app.',
		rating: 5,
	},
];

export default function Comparison() {
	return (
		<section className="relative py-24 md:py-32 bg-black">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="text-center mb-16">
					<h2 className="text-5xl md:text-6xl font-bold mb-6">
						<span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
							Real Stories from Real Users
						</span>
					</h2>
					<p className="text-xl text-slate-400 max-w-2xl mx-auto">
						See how travelers and locals are connecting
					</p>
				</div>

				<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
					{testimonials.map((testimonial, index) => (
						<div
							key={index}
							className="p-6 rounded-xl bg-gradient-to-br from-slate-900 to-slate-950 border border-slate-800 hover:border-cyan-600 transition-all"
						>
							<div className="flex gap-1 mb-4">
								{[...Array(testimonial.rating)].map((_, i) => (
									<Star key={i} className="w-5 h-5 fill-yellow-400 text-yellow-400" />
								))}
							</div>
							<p className="text-slate-300 mb-6 text-lg">"{testimonial.text}"</p>
							<div>
								<p className="font-semibold text-white">{testimonial.name}</p>
								<p className="text-slate-400 text-sm">{testimonial.location}</p>
							</div>
						</div>
					))}
				</div>

				{/* Trust section */}
				<div className="mt-20 grid grid-cols-1 md:grid-cols-3 gap-6">
					<div className="p-6 rounded-lg bg-slate-900/50 border border-slate-800 text-center">
						<h4 className="text-white font-semibold mb-2">🛡️ Safe & Verified</h4>
						<p className="text-slate-400 text-sm">All users verified. Block and report features built-in.</p>
					</div>
					<div className="p-6 rounded-lg bg-slate-900/50 border border-slate-800 text-center">
						<h4 className="text-white font-semibold mb-2">🔒 Your Privacy First</h4>
						<p className="text-slate-400 text-sm">End-to-end encrypted. Your data is yours alone.</p>
					</div>
					<div className="p-6 rounded-lg bg-slate-900/50 border border-slate-800 text-center">
						<h4 className="text-white font-semibold mb-2">⚡ Always Available</h4>
						<p className="text-slate-400 text-sm">24/7 support for travelers worldwide.</p>
					</div>
				</div>
			</div>
		</section>
	);
}
