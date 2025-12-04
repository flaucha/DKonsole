export const defineMonacoTheme = (monaco) => {
    monaco.editor.defineTheme('dkonsole-dark', {
        base: 'vs-dark',
        inherit: true,
        rules: [
            { token: '', foreground: 'abb2bf', background: '282c34' }, // default
            { token: 'comment', foreground: '5c6370' },
            { token: 'keyword', foreground: 'c678dd' },
            { token: 'string', foreground: '98c379' },
            { token: 'number', foreground: 'd19a66' },
            { token: 'type', foreground: '56b6c2' },
            { token: 'delimiter', foreground: 'abb2bf' },
            { token: 'key', foreground: 'e06c75' },
        ],
        colors: {
            'editor.background': '#282c34',
            'editor.foreground': '#abb2bf',
            'editor.lineHighlightBackground': '#2c313c',
            'editorCursor.foreground': '#528bff',
            'editor.selectionBackground': '#3e4451',
            'editor.selectionHighlightBackground': '#3e445180',
            'editorWhitespace.foreground': '#4b5263',
            'editorIndentGuide.background': '#3e4451',
            'editorIndentGuide.activeBackground': '#abb2bf',
            'editorLineNumber.foreground': '#5c6370',
            'editorLineNumber.activeForeground': '#abb2bf',
        }
    });
};
