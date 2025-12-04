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
            { token: 'regexp', foreground: '56b6c2' },
            { token: 'type', foreground: 'e5c07b' },
            { token: 'delimiter', foreground: 'abb2bf' },
            { token: 'operator', foreground: '56b6c2' },
            { token: 'variable', foreground: 'e06c75' },
            { token: 'identifier', foreground: 'abb2bf' },
            { token: 'key', foreground: 'e06c75' },
            { token: 'string.key.yaml', foreground: 'e5c07b' },
            { token: 'attribute.name', foreground: '61afef' },
        ],
        colors: {
            'editor.background': '#282c34',
            'editor.foreground': '#abb2bf',
            'editor.lineHighlightBackground': '#2c313c',
            'editorCursor.foreground': '#528bff',
            'editor.selectionBackground': '#3e4451',
            'editor.selectionHighlightBackground': '#3e445180',
            'editorWhitespace.foreground': '#4b5263',
            'editorIndentGuide.background': '#3b4048',
            'editorIndentGuide.activeBackground': '#7f848e',
            'editorLineNumber.foreground': '#636d83',
            'editorLineNumber.activeForeground': '#abb2bf',
            'editor.inactiveSelectionBackground': '#3e445199',
        }
    });
};
