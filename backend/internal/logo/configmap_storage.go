package logo

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

const (
	// ConfigMapName is the name of the ConfigMap used to store logos
	ConfigMapName = "dkonsole-logo"
	// ConfigMapNamespace is the namespace where the ConfigMap is stored
	ConfigMapNamespace = "dkonsole"
)

// ConfigMapLogoStorage implements LogoStorage using Kubernetes ConfigMap
type ConfigMapLogoStorage struct {
	client    kubernetes.Interface
	namespace string
}

// NewConfigMapLogoStorage creates a new ConfigMapLogoStorage
func NewConfigMapLogoStorage(client kubernetes.Interface, namespace string) *ConfigMapLogoStorage {
	return &ConfigMapLogoStorage{
		client:    client,
		namespace: namespace,
	}
}

// EnsureDataDir is a no-op for ConfigMap storage (ConfigMaps don't need directories)
func (s *ConfigMapLogoStorage) EnsureDataDir(ctx context.Context) error {
	// Ensure ConfigMap exists
	_, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, ConfigMapName, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, create it
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ConfigMapName,
				Namespace: s.namespace,
			},
			Data: make(map[string]string),
		}
		_, err = s.client.CoreV1().ConfigMaps(s.namespace).Create(ctx, cm, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ConfigMap: %w", err)
		}
	}
	return nil
}

// Save saves a logo file to ConfigMap
func (s *ConfigMapLogoStorage) Save(ctx context.Context, logoType string, ext string, content io.Reader) error {
	// Read content into memory
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	// Determine key name in ConfigMap
	var key string
	if logoType == "light" {
		key = "logo-light" + ext
	} else {
		key = "logo" + ext
	}

	// Encode content as base64 for storage in ConfigMap
	// ConfigMaps can store binary data as base64 strings
	encodedContent := base64.StdEncoding.EncodeToString(contentBytes)

	// Get existing ConfigMap or create new one
	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, ConfigMapName, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, create it
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ConfigMapName,
				Namespace: s.namespace,
			},
			Data: make(map[string]string),
		}
		cm.Data[key] = encodedContent
		_, err = s.client.CoreV1().ConfigMaps(s.namespace).Create(ctx, cm, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ConfigMap: %w", err)
		}
	} else {
		// Update existing ConfigMap
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data[key] = encodedContent
		_, err = s.client.CoreV1().ConfigMaps(s.namespace).Update(ctx, cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ConfigMap: %w", err)
		}
	}

	utils.LogInfo("Saving logo to ConfigMap", map[string]interface{}{
		"type":      logoType,
		"key":       key,
		"namespace": s.namespace,
	})
	return nil
}

// Get returns the logo content from ConfigMap (as base64 string for internal use)
// Note: For ConfigMap storage, we return the base64-encoded content, not a file path
func (s *ConfigMapLogoStorage) Get(ctx context.Context, logoType string, ext string) (string, error) {
	var key string
	if logoType == "light" {
		key = "logo-light" + ext
	} else {
		key = "logo" + ext
	}

	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("logo not found: %w", err)
	}

	encodedContent, exists := cm.Data[key]
	if !exists {
		return "", fmt.Errorf("logo not found")
	}

	// Return the base64-encoded content
	// The caller will decode it when needed
	return encodedContent, nil
}

// RemoveAll removes all existing logo files of the specified type from ConfigMap
func (s *ConfigMapLogoStorage) RemoveAll(ctx context.Context, logoType string) error {
	extensions := []string{".png", ".svg"}
	var filenamePrefix string
	if logoType == "light" {
		filenamePrefix = "logo-light"
	} else {
		filenamePrefix = "logo"
	}

	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, ConfigMapName, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, nothing to remove
		return nil
	}

	// Remove keys matching the pattern
	for _, ext := range extensions {
		key := filenamePrefix + ext
		delete(cm.Data, key)
	}

	// Update ConfigMap
	_, err = s.client.CoreV1().ConfigMaps(s.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}

	return nil
}

// GetLogoContent returns the decoded logo content from ConfigMap
func (s *ConfigMapLogoStorage) GetLogoContent(ctx context.Context, logoType string, ext string) ([]byte, error) {
	encodedContent, err := s.Get(ctx, logoType, ext)
	if err != nil {
		return nil, err
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(encodedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo content: %w", err)
	}

	return content, nil
}

// GetLogoContentByKey returns the decoded logo content from ConfigMap by key
func (s *ConfigMapLogoStorage) GetLogoContentByKey(ctx context.Context, key string) ([]byte, error) {
	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("logo not found: %w", err)
	}

	encodedContent, exists := cm.Data[key]
	if !exists {
		return nil, fmt.Errorf("logo not found")
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(encodedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo content: %w", err)
	}

	return content, nil
}

// ListLogoKeys returns all logo keys in the ConfigMap
func (s *ConfigMapLogoStorage) ListLogoKeys(ctx context.Context) ([]string, error) {
	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, ConfigMapName, metav1.GetOptions{})
	if err != nil {
		return []string{}, nil // ConfigMap doesn't exist, return empty list
	}

	keys := make([]string, 0, len(cm.Data))
	for key := range cm.Data {
		if strings.HasPrefix(key, "logo") {
			keys = append(keys, key)
		}
	}

	return keys, nil
}
