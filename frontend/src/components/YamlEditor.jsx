import React, { useState } from 'react';
import { AlertTriangle, Loader2 } from 'lucide-react';
import Editor from '@monaco-editor/react';
import { defineMonacoTheme } from '../config/monacoTheme';
import { useSettings } from '../context/SettingsContext';
import { useAuth } from '../context/AuthContext';
import { useYamlEditor } from '../hooks/useYamlEditor';
import YamlToolbar from './yaml/YamlToolbar';

const YamlEditor = ({ resource, onClose, onSaved }) => {
  const { currentCluster } = useSettings();
  const { authFetch } = useAuth();
  const [copying, setCopying] = useState(false);

  const {
    content,
    setContent,
    loading,
    saving,
    error,
    setError,
    handleSave
  } = useYamlEditor(resource, currentCluster, authFetch, onSaved);

  const handleEditorWillMount = (monaco) => {
    defineMonacoTheme(monaco);
  };

  const handleEditorDidMount = (editor, monaco) => {
    const remeasure = () => {
      monaco.editor.remeasureFonts();
      editor.layout();
    };

    editor.updateOptions({
      lineHeight: 22,
      letterSpacing: 0,
    });

    // Run after fonts load and on next tick to avoid cursor drift due to stale metrics
    document.fonts?.ready.then(remeasure).catch(() => { });
    setTimeout(remeasure, 0);

    // Keep in sync on resize; clean up on dispose
    window.addEventListener('resize', remeasure);
    editor.onDidDispose(() => {
      window.removeEventListener('resize', remeasure);
    });
  };

  const handleCopy = async () => {
    try {
      setCopying(true);
      await navigator.clipboard.writeText(content);
      setTimeout(() => setCopying(false), 800);
    } catch {
      setCopying(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 w-full max-w-5xl h-[85vh] rounded-lg border border-gray-700 flex flex-col shadow-2xl overflow-hidden">
        <YamlToolbar
          kind={resource?.kind}
          name={resource?.name}
          namespace={resource?.namespace}
          isNew={resource?.isNew}
          loading={loading}
          saving={saving}
          onSave={handleSave}
          onCopy={handleCopy}
          onClose={onClose}
          copying={copying}
        />

        {error && (
          <div className="bg-red-900/30 text-red-200 px-4 py-3 flex items-start justify-between border-b border-red-800">
            <div className="flex items-start space-x-2">
              <AlertTriangle size={18} className="mt-0.5" />
              <span className="text-sm whitespace-pre-wrap">{error}</span>
            </div>
            <button
              onClick={() => setError('')}
              className="text-xs px-2 py-1 border border-red-700 rounded hover:bg-red-800/50 transition-colors"
            >
              Dismiss
            </button>
          </div>
        )}

        <div className="flex-1 flex flex-col relative monaco-editor-container" style={{ backgroundColor: '#1f2937' }}>
          {loading ? (
            <div className="flex-1 flex items-center justify-center text-gray-400">
              <Loader2 size={20} className="animate-spin mr-2" />
              Loading YAML...
            </div>
          ) : (
            <Editor
              height="100%"
              defaultLanguage="yaml"
              theme="dkonsole-dark"
              beforeMount={handleEditorWillMount}
              onMount={handleEditorDidMount}
              value={content}
              onChange={(value) => setContent(value)}
              options={{
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                fontSize: 14,
                automaticLayout: true,
              }}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default YamlEditor;
