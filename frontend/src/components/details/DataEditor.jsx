import React, { useState, useEffect } from 'react';
import { Save, X, Plus, Trash2, Loader2 } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useSettings } from '../../context/SettingsContext';
import { logger } from '../../utils/logger';

export const DataEditor = ({ resource, data, isSecret, onClose, onSaved }) => {
    const { authFetch } = useAuth();
    const { currentCluster } = useSettings();
    const [editingData, setEditingData] = useState({});
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');

    // Initialize editing data
    // Note: For Secrets, the backend already decodes the values, so we use them directly
    useEffect(() => {
        if (!data) {
            setEditingData({});
            return;
        }

        // Data comes already decoded from backend for Secrets
        setEditingData({ ...data });
    }, [data]);

    const handleKeyChange = (oldKey, newKey) => {
        if (oldKey === newKey) return;
        
        const newData = { ...editingData };
        const value = newData[oldKey];
        delete newData[oldKey];
        newData[newKey] = value;
        setEditingData(newData);
    };

    const handleValueChange = (key, value) => {
        setEditingData({ ...editingData, [key]: value });
    };

    const handleAddPair = () => {
        const newKey = `key${Object.keys(editingData).length + 1}`;
        setEditingData({ ...editingData, [newKey]: '' });
    };

    const handleRemovePair = (key) => {
        const newData = { ...editingData };
        delete newData[key];
        setEditingData(newData);
    };

    const handleSave = async () => {
        setSaving(true);
        setError('');

        try {
            // Get current resource YAML
            const yamlUrl = `/api/resource/yaml?kind=${resource.kind}&name=${resource.name}${resource.namespace ? `&namespace=${resource.namespace}` : ''}${currentCluster ? `&cluster=${currentCluster}` : ''}`;
            const yamlRes = await authFetch(yamlUrl);
            
            if (!yamlRes.ok) {
                throw new Error('Failed to load resource YAML');
            }

            const yamlText = await yamlRes.text();
            
            // Verify YAML has required fields
            if (!yamlText.includes('kind:') && !yamlText.includes('kind:')) {
                throw new Error('YAML from server is missing kind field');
            }
            
            // Parse YAML to update data field - more robust approach
            const lines = yamlText.split('\n');
            let dataStartIndex = -1;
            let dataEndIndex = -1;
            let dataIndent = 0;
            let foundData = false;

            // Find data section
            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];
                const trimmed = line.trim();
                
                if (trimmed === 'data:' || trimmed.startsWith('data:')) {
                    dataStartIndex = i;
                    dataIndent = line.search(/\S/);
                    foundData = true;
                    continue;
                }

                if (foundData && dataEndIndex === -1) {
                    const currentIndent = line.search(/\S/);
                    // Skip empty lines
                    if (trimmed === '') {
                        continue;
                    }
                    // End of data section: line with indent strictly less than dataIndent
                    // This means we've moved to a parent level (like another top-level key)
                    if (currentIndent !== -1 && currentIndent < dataIndent) {
                        dataEndIndex = i;
                        foundData = false;
                        break;
                    }
                    // Also check for lines at same indent that are clearly new top-level keys
                    // (not continuation of data values)
                    if (currentIndent === dataIndent && trimmed.match(/^[a-zA-Z][a-zA-Z0-9_-]*:/)) {
                        dataEndIndex = i;
                        foundData = false;
                        break;
                    }
                }
            }

            // If we didn't find an end, data section goes to end of file
            if (dataEndIndex === -1 && foundData) {
                dataEndIndex = lines.length;
            }

            // Build new data section
            const dataLines = [];
            for (const [key, value] of Object.entries(editingData)) {
                if (isSecret) {
                    // Encode to base64 for secrets
                    // Use UTF-8 safe encoding
                    const encoded = btoa(unescape(encodeURIComponent(value)));
                    dataLines.push(`${' '.repeat(dataIndent + 2)}${key}: ${encoded}`);
                } else {
                    // For ConfigMaps, check if value needs special formatting
                    if (value === '') {
                        dataLines.push(`${' '.repeat(dataIndent + 2)}${key}: ""`);
                    } else if (value.includes('\n')) {
                        // Multi-line value - use YAML block scalar
                        dataLines.push(`${' '.repeat(dataIndent + 2)}${key}: |`);
                        value.split('\n').forEach(line => {
                            dataLines.push(`${' '.repeat(dataIndent + 4)}${line}`);
                        });
                    } else if (value.includes(':') || value.includes('#') || value.trim() !== value || value.startsWith('|') || value.startsWith('>')) {
                        // Value needs quoting
                        const escaped = value.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
                        dataLines.push(`${' '.repeat(dataIndent + 2)}${key}: "${escaped}"`);
                    } else {
                        dataLines.push(`${' '.repeat(dataIndent + 2)}${key}: ${value}`);
                    }
                }
            }

            // Reconstruct YAML
            // Ensure we preserve the entire YAML structure, especially kind and apiVersion
            if (dataStartIndex === -1) {
                // If no data section exists, we need to add it
                // Find where to insert it (after metadata section, before any other sections)
                let insertIndex = lines.length;
                for (let i = 0; i < lines.length; i++) {
                    const line = lines[i];
                    const trimmed = line.trim();
                    // Look for common sections that come after data
                    if (trimmed.startsWith('type:') || trimmed.startsWith('immutable:') || trimmed.startsWith('binaryData:')) {
                        insertIndex = i;
                        break;
                    }
                }
                // Insert data section
                const indent = 2; // Default indent for data section
                lines.splice(insertIndex, 0, 'data:');
                dataStartIndex = insertIndex;
                dataIndent = 0; // Top level
            }

            const beforeData = lines.slice(0, dataStartIndex + 1);
            const afterData = dataEndIndex === -1 ? [] : lines.slice(dataEndIndex);
            const finalLines = [...beforeData, ...dataLines, ...afterData];

            const newYaml = finalLines.join('\n');

            // Verify the YAML has kind field (basic check)
            if (!newYaml.includes('kind:')) {
                throw new Error('YAML reconstruction failed: kind field missing. Original YAML may be corrupted.');
            }

            // Save using import endpoint (Server-Side Apply)
            const params = new URLSearchParams();
            if (currentCluster) params.append('cluster', currentCluster);

            const saveRes = await authFetch(`/api/resource/import?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-yaml' },
                body: newYaml,
            });

            if (!saveRes.ok) {
                const errorText = await saveRes.text();
                // Log the YAML for debugging if there's an error
                logger.error('YAML that failed:', newYaml.substring(0, 500));
                logger.error('Error from server:', errorText);
                throw new Error(errorText || 'Failed to save changes');
            }

            onSaved?.();
            onClose();
        } catch (err) {
            setError(err.message);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg w-full max-w-4xl max-h-[90vh] flex flex-col shadow-xl">
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-gray-700">
                    <h3 className="text-lg font-semibold text-white">
                        Edit {isSecret ? 'Secret' : 'ConfigMap'} Data - {resource.name}
                    </h3>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-200 transition-colors"
                        disabled={saving}
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Error message */}
                {error && (
                    <div className="mx-4 mt-4 p-3 bg-red-900/30 border border-red-700 rounded text-sm text-red-200">
                        {error}
                    </div>
                )}

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-4">
                    <div className="space-y-3">
                        {Object.entries(editingData).map(([key, value]) => (
                            <div key={key} className="bg-gray-800/50 p-4 rounded-md border border-gray-700/50">
                                <div className="flex gap-2 mb-2">
                                    <input
                                        type="text"
                                        value={key}
                                        onChange={(e) => handleKeyChange(key, e.target.value)}
                                        className="flex-1 px-3 py-2 bg-gray-900 border border-gray-600 rounded text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-gray-500"
                                        placeholder="Key"
                                        disabled={saving}
                                    />
                                    <button
                                        onClick={() => handleRemovePair(key)}
                                        className="px-3 py-2 bg-red-900/30 hover:bg-red-900/50 text-red-200 rounded border border-red-700/50 transition-colors"
                                        disabled={saving}
                                        title="Remove key"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                                <textarea
                                    value={value}
                                    onChange={(e) => handleValueChange(key, e.target.value)}
                                    className="w-full px-3 py-2 bg-gray-900 border border-gray-600 rounded text-sm font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-gray-500 resize-y min-h-[80px]"
                                    placeholder="Value"
                                    disabled={saving}
                                    rows={Math.max(3, value.split('\n').length)}
                                />
                            </div>
                        ))}
                    </div>

                    {Object.keys(editingData).length === 0 && (
                        <div className="text-center py-8 text-gray-500 italic">
                            No data entries. Click "Add Key-Value Pair" to add one.
                        </div>
                    )}

                    <button
                        onClick={handleAddPair}
                        className="mt-4 flex items-center gap-2 px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded border border-gray-600 transition-colors"
                        disabled={saving}
                    >
                        <Plus size={16} />
                        Add Key-Value Pair
                    </button>
                </div>

                {/* Footer */}
                <div className="flex items-center justify-end gap-3 p-4 border-t border-gray-700">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-200 rounded border border-gray-600 transition-colors"
                        disabled={saving}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        disabled={saving}
                    >
                        {saving ? (
                            <>
                                <Loader2 size={16} className="animate-spin" />
                                Saving...
                            </>
                        ) : (
                            <>
                                <Save size={16} />
                                Save Changes
                            </>
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
};

