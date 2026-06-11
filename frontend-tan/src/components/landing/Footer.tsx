import React from 'react';
import { Mail, X } from 'lucide-react';

export default function Footer() {
	return (
		<footer className="relative bg-black border-t border-slate-800">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
				<div className="grid grid-cols-1 md:grid-cols-4 gap-12 mb-12">
					{/* Brand */}
					<div className="md:col-span-1">
						<h3 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent mb-4">
							Twinspeak
						</h3>
						<p className="text-slate-400 text-sm">Connect with travelers and locals around the world. Real conversations, real connections.</p>
						<div className="flex gap-4 mt-6">
							<a href="#" className="text-slate-400 hover:text-cyan-400 transition">
								<X className="w-5 h-5" />
							</a>
							<a href="#" className="text-slate-400 hover:text-cyan-400 transition">
								<Mail className="w-5 h-5" />
							</a>
						</div>
					</div>

					{/* App */}
					<div>
						<h4 className="text-white font-semibold mb-4">App</h4>
						<ul className="space-y-3">
							{['Download', 'How It Works', 'Languages', 'Safety'].map((item) => (
								<li key={item}>
									<a href="#" className="text-slate-400 hover:text-cyan-400 transition text-sm">
										{item}
									</a>
								</li>
							))}
						</ul>
					</div>

					{/* Company */}
					<div>
						<h4 className="text-white font-semibold mb-4">Company</h4>
						<ul className="space-y-3">
							{['About', 'Blog', 'Careers', 'Contact'].map((item) => (
								<li key={item}>
									<a href="#" className="text-slate-400 hover:text-cyan-400 transition text-sm">
										{item}
									</a>
								</li>
							))}
						</ul>
					</div>

					{/* Legal */}
					<div>
						<h4 className="text-white font-semibold mb-4">Legal</h4>
						<ul className="space-y-3">
							{['Privacy', 'Terms', 'Safety', 'Guidelines'].map((item) => (
								<li key={item}>
									<a href="#" className="text-slate-400 hover:text-cyan-400 transition text-sm">
										{item}
									</a>
								</li>
							))}
						</ul>
					</div>
				</div>

				{/* Bottom */}
				<div className="border-t border-slate-800 pt-8">
					<div className="flex flex-col md:flex-row justify-between items-center gap-6">
						<p className="text-slate-500 text-sm">
							© 2024 Twinspeak. All rights reserved.
						</p>
						<div className="flex gap-6">
							<a href="#" className="text-slate-500 hover:text-slate-300 text-sm transition">
								Status
							</a>
							<a href="#" className="text-slate-500 hover:text-slate-300 text-sm transition">
								Support
							</a>
							<a href="#" className="text-slate-500 hover:text-slate-300 text-sm transition">
								Community
							</a>
						</div>
					</div>
				</div>
			</div>
		</footer>
	);
}
