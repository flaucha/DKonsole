import React from 'react';
import { Github, Mail, Coffee, Code } from 'lucide-react';

const AcercaDe = () => {
    const version = '1.1.5';
    const githubUrl = 'https://github.com/flaucha/DKonsole';
    const email = 'flaucha@gmail.com';
    const buymeacoffeeUrl = 'https://buymeacoffee.com/flaucha';

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <div className="bg-gray-800 p-8 rounded-lg border border-gray-700 shadow-lg">
                <div className="flex items-center mb-6">
                    <Code size={32} className="mr-3 text-blue-400" />
                    <h1 className="text-3xl font-bold text-white">DKonsole</h1>
                </div>
                
                <p className="text-gray-300 mb-6 text-lg">
                    A modern, lightweight Kubernetes dashboard built entirely with <strong>Artificial Intelligence</strong>.
                </p>

                <div className="space-y-4 mb-8">
                    <div className="flex items-center text-gray-300">
                        <span className="font-semibold mr-2">Version:</span>
                        <span className="px-2 py-1 bg-blue-900/30 text-blue-300 rounded text-sm font-mono">{version}</span>
                    </div>

                    <div className="flex items-center space-x-4">
                        <a
                            href={githubUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-md transition-colors"
                        >
                            <Github size={18} className="mr-2" />
                            GitHub
                        </a>

                        <a
                            href={`mailto:${email}`}
                            className="flex items-center px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-md transition-colors"
                        >
                            <Mail size={18} className="mr-2" />
                            Email
                        </a>

                        <a
                            href={buymeacoffeeUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-md transition-colors"
                        >
                            <Coffee size={18} className="mr-2" />
                            Buy me a coffee
                        </a>
                    </div>
                </div>

                <div className="border-t border-gray-700 pt-6">
                    <h2 className="text-lg font-semibold text-white mb-4">Features</h2>
                    <ul className="space-y-2 text-gray-300">
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>Resource Management: View and manage Deployments, Pods, Services, ConfigMaps, Secrets, and more</span>
                        </li>
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>Prometheus Integration: Historical metrics for Pods with customizable time ranges</span>
                        </li>
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>Live Logs: Stream logs from containers in real-time</span>
                        </li>
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>Terminal Access: Execute commands directly in pod containers</span>
                        </li>
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>YAML Editor: Edit resources with a built-in YAML editor</span>
                        </li>
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>Secure Authentication: Argon2 password hashing and JWT-based sessions</span>
                        </li>
                        <li className="flex items-start">
                            <span className="text-blue-400 mr-2">•</span>
                            <span>Multi-Cluster Support: Manage multiple Kubernetes clusters from a single interface</span>
                        </li>
                    </ul>
                </div>

                <div className="border-t border-gray-700 pt-6 mt-6">
                    <p className="text-sm text-gray-500">
                        DKonsole is licensed under the MIT License.
                    </p>
                </div>
            </div>
        </div>
    );
};

export default AcercaDe;

