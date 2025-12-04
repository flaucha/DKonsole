/**
 * Parses error responses from the API and extracts a human-readable error message.
 * Handles JSON bodies, plain text and empty responses without throwing.
 *
 * @param {Response} response - The fetch Response object
 * @returns {Promise<string>} A human-readable error message
 */
export async function parseErrorResponse(response) {
    const status = response?.status;
    const statusText = response?.statusText || '';
    const contentType = response?.headers?.get?.('content-type') || '';

    try {
        const raw = (await response?.text?.()) || '';
        const text = raw.trim();
        const looksJson = contentType.includes('application/json') || text.startsWith('{') || text.startsWith('[');

        if (looksJson && text) {
            try {
                const errorData = JSON.parse(text);

                if (typeof errorData === 'string') {
                    return errorData;
                }

                const knownFields = ['error', 'message', 'detail', 'description', 'reason'];
                for (const field of knownFields) {
                    if (errorData?.[field]) {
                        return String(errorData[field]);
                    }
                }

                if (typeof errorData === 'object' && errorData !== null) {
                    const stringified = JSON.stringify(errorData, null, 2);
                    if (stringified.length < 200) {
                        return stringified;
                    }
                    const firstValue = Object.values(errorData)[0];
                    if (firstValue) {
                        return String(firstValue);
                    }
                }
            } catch {
                // Parsing failed, fall back to the raw text below.
            }
        }

        if (text) {
            return text;
        }

        return statusText || (status ? `Error ${status}` : 'Unknown error');
    } catch {
        return status ? `Error ${status}: ${statusText || 'Unknown error'}` : 'Unknown error';
    }
}

// Backwards compatibility for older imports.
export const parseResponseError = parseErrorResponse;

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
