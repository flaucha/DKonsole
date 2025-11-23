
export const parseCpu = (value) => {
    if (!value) return 0;
    if (value.endsWith('m')) {
        return parseInt(value.slice(0, -1), 10);
    }
    return parseFloat(value) * 1000;
};

export const parseMemory = (value) => {
    if (!value) return 0;
    const units = {
        Ki: 1024,
        Mi: 1024 * 1024,
        Gi: 1024 * 1024 * 1024,
        Ti: 1024 * 1024 * 1024 * 1024,
    };

    for (const [unit, multiplier] of Object.entries(units)) {
        if (value.endsWith(unit)) {
            return parseFloat(value.slice(0, -unit.length)) * multiplier;
        }
    }
    return parseInt(value, 10); // Assumes bytes if no unit
};

export const calculatePercentage = (used, hard) => {
    if (!used || !hard) return 0;

    // Determine type based on hard value format (heuristic)
    const isCpu = hard.endsWith('m') || !isNaN(hard); // Simple check, can be improved
    const isMemory = ['Ki', 'Mi', 'Gi', 'Ti'].some(u => hard.endsWith(u));

    let usedVal = 0;
    let hardVal = 0;

    if (isMemory) {
        usedVal = parseMemory(used);
        hardVal = parseMemory(hard);
    } else {
        // Default to CPU/count parsing
        usedVal = parseCpu(used);
        hardVal = parseCpu(hard);
    }

    if (hardVal === 0) return 0;
    return Math.min(Math.round((usedVal / hardVal) * 100), 100);
};

export const formatResourceValue = (value) => {
    if (!value) return '-';
    return value;
};
