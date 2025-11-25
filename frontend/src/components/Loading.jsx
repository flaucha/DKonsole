import React from 'react';
import { Loader2 } from 'lucide-react';

const Loading = ({ message = 'Loading...' }) => {
    return (
        <div className="min-h-screen bg-gray-900 flex items-center justify-center text-white">
            <div className="flex flex-col items-center space-y-4">
                <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
                <p className="text-gray-400">{message}</p>
            </div>
        </div>
    );
};

export default Loading;
