import React from 'react';
import { Smartphone, Zap, Users, Globe } from 'lucide-react';

const features = [
	{
		icon: Smartphone,
		title: 'Mobile First',
		description: 'Built for travelers. Everything you need right in your pocket—simple, fast, and intuitive.',
	},
	{
		icon: Zap,
		title: 'Instant Translation',
		description: "See each other's messages translated in real-time. No delays, no confusion.",
	},
	{
		icon: Users,
		title: 'Match with Locals',
		description: 'Connect with people who want to help travelers. Make friends while exploring.',
	},
	{
		icon: Globe,
		title: '50+ Languages',
		description: 'From English to Mandarin, Thai to Portuguese—we support the languages you need.',
	},
];

export default function Features() {
	return (
		<section className="relative py-24 md:py-32 bg-black">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="text-center mb-16">
					<h2 className="text-5xl md:text-6xl font-bold mb-6">
						<span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
							Made for Travelers
						</span>
					</h2>
					<p className="text-xl text-slate-400 max-w-2xl mx-auto">
						Everything you need for authentic local conversations, no matter where you are.
					</p>
				</div>

				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
					{features.map((feature, index) => {
						const Icon = feature.icon;
						return (
							<div
								key={index}
								className="group relative p-6 rounded-xl bg-gradient-to-br from-slate-900 to-slate-950 border border-slate-800 hover:border-cyan-600 transition-all duration-300 hover:shadow-lg hover:shadow-cyan-500/20"
							>
								<div className="w-12 h-12 rounded-lg bg-gradient-to-br from-blue-600 to-cyan-600 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
									<Icon className="w-6 h-6 text-white" />
								</div>
								<h3 className="text-lg font-semibold text-white mb-2">{feature.title}</h3>
								<p className="text-slate-400 text-sm leading-relaxed">{feature.description}</p>
							</div>
						);
					})}
				</div>
			</div>
		</section>
	);
}
