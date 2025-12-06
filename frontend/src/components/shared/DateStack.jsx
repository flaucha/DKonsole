import React from 'react';
import { formatDateParts } from '../../utils/dateUtils';

const DateStack = ({ value }) => {
    const { date, time } = formatDateParts(value);
    return (
        <div className="flex flex-col items-center leading-tight text-sm text-gray-300">
            <span>{date}</span>
            <span className="text-xs text-gray-500">{time}</span>
        </div>
    );
};

export default DateStack;
