/**
 * Parses error responses from the API and extracts a human-readable error message.
 * Handles both JSON and text responses.
 *
 * @param {Response} response - The fetch Response object
 * @returns {Promise<string>} A human-readable error message
 */
export async function parseErrorResponse(response) {
    try {
        const contentType = response.headers.get('content-type');

        // Try to parse as JSON first
        if (contentType && contentType.includes('application/json')) {
            const errorData = await response.json();

            // Handle different error response formats
            if (typeof errorData === 'string') {
                return errorData;
            }

            if (errorData.error) {
                return errorData.error;
            }

            if (errorData.message) {
                return errorData.message;
            }

            // If it's an object, try to extract meaningful information
            if (typeof errorData === 'object') {
                // Check for common error fields
                const errorFields = ['error', 'message', 'detail', 'description', 'reason'];
                for (const field of errorFields) {
                    if (errorData[field]) {
                        return String(errorData[field]);
                    }
                }

                // If no standard field found, stringify but make it readable
                const errorStr = JSON.stringify(errorData, null, 2);
                // If it's a short error, return it directly
                if (errorStr.length < 200) {
                    return errorStr;
                }
                // Otherwise, try to extract the first meaningful value
                const firstValue = Object.values(errorData)[0];
                if (firstValue) {
                    return String(firstValue);
                }
            }
        }

        // Fallback to text
        const text = await response.text();
        if (text) {
            // Try to parse as JSON if it looks like JSON
            if (text.trim().startsWith('{') || text.trim().startsWith('[')) {
                try {
                    const parsed = JSON.parse(text);
                    if (parsed.error) return parsed.error;
                    if (parsed.message) return parsed.message;
                } catch {
                    // Not valid JSON, return as text
                }
            }
            return text;
        }

        // Last resort: status text
        return response.statusText || `Error ${response.status}`;
    } catch (err) {
        // If all else fails, return a generic error
        return `Error ${response.status}: ${response.statusText || 'Unknown error'}`;
    }
}

/**
 * Parses an error object (from catch blocks) and extracts a human-readable message.
 *
 * @param {Error|string|object} error - The error object
 * @returns {string} A human-readable error message
 */
export function parseError(error) {
    if (!error) {
        return 'An unknown error occurred';
    }

    if (typeof error === 'string') {
        // Try to parse if it looks like JSON
        if (error.trim().startsWith('{') || error.trim().startsWith('[')) {
            try {
                const parsed = JSON.parse(error);
                if (parsed.error) return parsed.error;
                if (parsed.message) return parsed.message;
            } catch {
                // Not valid JSON, return as string
            }
        }
        return error;
    }

    if (error instanceof Error) {
        // Check if error message is JSON
        if (error.message && (error.message.trim().startsWith('{') || error.message.trim().startsWith('['))) {
            try {
                const parsed = JSON.parse(error.message);
                if (parsed.error) return parsed.error;
                if (parsed.message) return parsed.message;
            } catch {
                // Not valid JSON, return error message
            }
        }
        return error.message || error.toString();
    }

    if (typeof error === 'object') {
        if (error.error) return String(error.error);
        if (error.message) return String(error.message);
        if (error.detail) return String(error.detail);

        // Try to stringify but keep it short
        const errorStr = JSON.stringify(error);
        if (errorStr.length < 200) {
            return errorStr;
        }
        return 'An error occurred';
    }

    return String(error);
}
