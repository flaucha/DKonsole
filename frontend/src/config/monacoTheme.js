export const defineMonacoTheme = (monaco) => {
    monaco.editor.defineTheme('dkonsole-dark', {
        base: 'vs-dark',
        inherit: true,
        rules: [
            { token: '', background: '1f2937' }, // gray-800
            { token: 'comment', foreground: '6b7280' }, // gray-500
            { token: 'keyword', foreground: '60a5fa' }, // blue-400
            { token: 'string', foreground: '34d399' }, // green-400
            { token: 'number', foreground: 'f87171' }, // red-400
            { token: 'type', foreground: 'a78bfa' }, // purple-400
        ],
        colors: {
            'editor.background': '#1f2937', // gray-800
            'editor.foreground': '#e5e7eb', // gray-200
            'editor.lineHighlightBackground': '#374151', // gray-700
            'editorCursor.foreground': '#60a5fa', // blue-400
            'editorWhitespace.foreground': '#4b5563', // gray-600
            'editorIndentGuide.background': '#374151', // gray-700
            'editorIndentGuide.activeBackground': '#6b7280', // gray-500
        }
    });
};
