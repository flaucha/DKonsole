/**
 * DKonsole Custom Monaco Editor Theme
 * Ensures consistent styling across all application themes
 */

export const DKONSOLE_THEME = 'dkonsole-dark';

export const defineMonacoTheme = (monaco) => {
    monaco.editor.defineTheme(DKONSOLE_THEME, {
        base: 'vs-dark',
        inherit: true,
        rules: [
            // YAML/JSON Keys - Light blue
            { token: 'string.key.json', foreground: '9CDCFE' },
            { token: 'type.yaml', foreground: '9CDCFE' },

            // YAML/JSON Values - Light green
            { token: 'string.value.json', foreground: 'CE9178' },
            { token: 'string.yaml', foreground: 'CE9178' },

            // Numbers - Light green
            { token: 'number', foreground: 'B5CEA8' },

            // Booleans and null - Purple
            { token: 'keyword.json', foreground: 'C586C0' },
            { token: 'constant.language.yaml', foreground: 'C586C0' },

            // Comments - Gray
            { token: 'comment', foreground: '6A9955' },
        ],
        colors: {
            // Editor background
            'editor.background': '#1e1e1e',
            'editor.foreground': '#d4d4d4',

            // Line numbers
            'editorLineNumber.foreground': '#858585',
            'editorLineNumber.activeForeground': '#c6c6c6',

            // Current line
            'editor.lineHighlightBackground': '#2a2a2a',
            'editor.lineHighlightBorder': '#00000000',

            // Selection
            'editor.selectionBackground': '#264f78',
            'editor.inactiveSelectionBackground': '#3a3d41',

            // Cursor
            'editorCursor.foreground': '#aeafad',

            // Scrollbar
            'scrollbarSlider.background': '#79797966',
            'scrollbarSlider.hoverBackground': '#646464b3',
            'scrollbarSlider.activeBackground': '#bfbfbf66',

            // Whitespace
            'editorWhitespace.foreground': '#e3e4e229',

            // Indent guides
            'editorIndentGuide.background': '#404040',
            'editorIndentGuide.activeBackground': '#707070',
        },
    });
};
