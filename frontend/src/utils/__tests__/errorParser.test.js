import { parseErrorResponse, parseError } from '../errorParser'

const makeResponse = (status, statusText, body, headers = {}) => ({
    status,
    statusText,
    headers: {
        get: (key) => headers[key.toLowerCase()],
    },
    async text() {
        return body
    },
    async json() {
        return JSON.parse(body)
    },
})

describe('parseErrorResponse', () => {
    it('returns JSON error message when available', async () => {
        const res = makeResponse(400, 'Bad Request', JSON.stringify({ error: 'invalid' }), { 'content-type': 'application/json' })
        await expect(parseErrorResponse(res)).resolves.toBe('invalid')
    })

    it('falls back to status text when body empty', async () => {
        const res = makeResponse(500, 'Server Error', '', { 'content-type': 'application/json' })
        await expect(parseErrorResponse(res)).resolves.toBe('Server Error')
    })
})

describe('parseError', () => {
    it('handles string input', () => {
        expect(parseError('boom')).toBe('boom')
    })

    it('handles Error object', () => {
        expect(parseError(new Error('failed'))).toBe('failed')
    })
})
