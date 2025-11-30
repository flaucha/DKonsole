export const defineAtomDarkTheme = (monaco) => {
    if (!monaco) return;

    monaco.editor.defineTheme('atom-dark-custom', {
        base: 'vs-dark',
        inherit: true,
        rules: [
            // YAML specific colors - Atom Dark theme
            { token: 'delimiter', foreground: 'e06c75' }, // Rojo/rosa para los ":" y otros delimitadores
            { token: 'delimiter.colon', foreground: 'e06c75' }, // Rojo/rosa específico para ":"
            { token: 'string', foreground: '98c379' }, // Verde para strings/valores
            { token: 'string.yaml', foreground: '98c379' }, // Verde para valores YAML
            { token: 'number', foreground: 'd19a66' }, // Naranja para números
            { token: 'number.yaml', foreground: 'd19a66' }, // Naranja para números YAML
            { token: 'keyword', foreground: 'c678dd' }, // Morado para keywords
            { token: 'keyword.yaml', foreground: 'c678dd' }, // Morado para keywords YAML
            { token: 'type', foreground: 'e5c07b' }, // Amarillo para tipos
            { token: 'comment', foreground: '5c6370', fontStyle: 'italic' }, // Gris para comentarios
            { token: 'comment.yaml', foreground: '5c6370', fontStyle: 'italic' }, // Gris para comentarios YAML
            { token: 'key', foreground: 'e06c75' }, // Rojo/rosa para claves YAML
            { token: 'key.yaml', foreground: 'e06c75' }, // Rojo/rosa para claves YAML
            { token: 'attribute.name', foreground: 'e06c75' }, // Rojo/rosa para nombres de atributos
            { token: 'attribute.value', foreground: '98c379' }, // Verde para valores de atributos
        ],
        colors: {
            'editor.background': '#272b34',
            'editor.foreground': '#abb2bf',
            'editor.lineHighlightBackground': '#2c313c',
            'editor.selectionBackground': '#3e4451',
            'editorCursor.foreground': '#528bff',
            'editorWhitespace.foreground': '#3b4048',
            'editorIndentGuide.background': '#3b4048',
            'editorIndentGuide.activeBackground': '#9d550f',
            'editor.selectionHighlightBackground': '#3e4451',
            'editorLineNumber.foreground': '#5c6370',
            'editorLineNumber.activeForeground': '#abb2bf',
        }
    });
};
