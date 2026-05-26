import React from 'react';
import { Smartphone, Users, MessageCircle, Smile } from 'lucide-react';

const steps = [
	{
		icon: Smartphone,
		title: 'Download & Sign Up',
		description: 'Open the app and create your profile in 30 seconds. Tell us what languages you speak.',
	},
	{
		icon: Users,
		title: 'Find Someone to Chat',
		description: 'Browse profiles of travelers or locals in your city. Swipe and match with people.',
	},
	{
		icon: MessageCircle,
		title: 'Start a Conversation',
		description: 'Send messages, voice notes, or go live. Everything translates in real-time.',
	},
	{
		icon: Smile,
		title: 'Make a Connection',
		description: 'Learn about their culture. Share your stories. Build real friendships.',
	},
];

export default function HowItWorks() {
	return (
		<section id="how-it-works" className="relative py-24 md:py-32 bg-gradient-to-b from-black to-slate-900">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="text-center mb-20">
					<h2 className="text-5xl md:text-6xl font-bold mb-6">
						<span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
							Simple as 1, 2, 3, 4
						</span>
					</h2>
					<p className="text-xl text-slate-400 max-w-2xl mx-auto">
						Start chatting with locals in minutes
					</p>
				</div>

				<div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-12">
					{steps.map((step, index) => {
						const Icon = step.icon;
						return (
							<React.Fragment key={index}>
								<div className="relative">
									<div className="relative z-10 p-6 rounded-xl bg-gradient-to-br from-slate-800 to-slate-900 border border-slate-700 h-full">
										<div className="absolute -top-3 -left-3 w-8 h-8 rounded-full bg-gradient-to-r from-blue-600 to-cyan-600 flex items-center justify-center text-white font-bold text-sm">
											{index + 1}
										</div>
										<div className="mt-4 mb-3">
											<div className="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-600 to-cyan-600 flex items-center justify-center">
												<Icon className="w-5 h-5 text-white" />
											</div>
										</div>
										<h3 className="text-lg font-semibold text-white mb-2">{step.title}</h3>
										<p className="text-slate-400 text-sm">{step.description}</p>
									</div>
									{index < steps.length - 1 && (
										<div className="hidden md:flex absolute -right-2 top-1/2 transform -translate-y-1/2 z-0">
											<div className="w-4 h-0.5 bg-gradient-to-r from-cyan-600 to-transparent opacity-50"></div>
										</div>
									)}
								</div>
							</React.Fragment>
						);
					})}
				</div>

				{/* Highlight features */}
				<div className="grid grid-cols-1 md:grid-cols-2 gap-8 mt-20">
					<div className="p-8 rounded-xl bg-slate-900/50 border border-slate-800">
						<h3 className="text-2xl font-bold text-white mb-4">📍 Location-Based</h3>
						<ul className="space-y-3">
							{['Find people near you', 'See who\'s active right now', 'Plan meetups or video calls'].map((item, i) => (
								<li key={i} className="flex items-center gap-3 text-slate-300">
									<span className="w-1.5 h-1.5 rounded-full bg-cyan-500"></span>
									{item}
								</li>
							))}
						</ul>
					</div>

					<div className="p-8 rounded-xl bg-slate-900/50 border border-slate-800">
						<h3 className="text-2xl font-bold text-white mb-4">💬 Real-Time Chat</h3>
						<ul className="space-y-3">
							{['Text with instant translation', 'Send voice messages', 'Go live and talk face-to-face'].map((item, i) => (
								<li key={i} className="flex items-center gap-3 text-slate-300">
									<span className="w-1.5 h-1.5 rounded-full bg-cyan-500"></span>
									{item}
								</li>
							))}
						</ul>
					</div>
				</div>
			</div>
		</section>
	);
}
