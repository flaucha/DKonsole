import React from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

const MetricChart = ({ data, dataKey, color, title, unit, icon: Icon }) => {
    const gradientId = `color${dataKey}`;

    return (
        <div className="bg-gray-900/50 p-3 rounded-md border border-gray-700">
            <div className="flex items-center mb-2">
                {Icon && <Icon size={14} className={`mr-2 ${color === '#60A5FA' ? 'text-blue-400' : color === '#A78BFA' ? 'text-purple-400' : color === '#34D399' ? 'text-green-400' : color === '#FBBF24' ? 'text-yellow-400' : 'text-orange-400'}`} />}
                <h3 className="text-xs font-medium text-gray-300">{title} {unit ? `(${unit})` : ''}</h3>
            </div>
            <div className="h-32 w-full">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={data}>
                        <defs>
                            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor={color} stopOpacity={0.3} />
                                <stop offset="95%" stopColor={color} stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                        <XAxis
                            dataKey="time"
                            stroke="#9CA3AF"
                            fontSize={9}
                            tick={{ fill: '#9CA3AF' }}
                            interval="preserveStartEnd"
                        />
                        <YAxis stroke="#9CA3AF" fontSize={9} tick={{ fill: '#9CA3AF' }} />
                        <Tooltip
                            wrapperClassName="recharts-tooltip-wrapper"
                            contentStyle={{
                                backgroundColor: 'rgb(var(--color-gray-800))',
                                borderColor: 'rgb(var(--color-gray-700))',
                                color: 'rgb(var(--color-gray-100))',
                                fontSize: '11px'
                            }}
                            labelStyle={{ color: 'rgb(var(--color-gray-100))' }}
                            itemStyle={{ color: color }}
                        />
                        <Area
                            type="monotone"
                            dataKey={dataKey}
                            stroke={color}
                            fillOpacity={1}
                            fill={`url(#${gradientId})`}
                            isAnimationActive={false}
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default MetricChart;
