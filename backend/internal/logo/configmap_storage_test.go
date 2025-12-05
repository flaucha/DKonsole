package logo

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestConfigMapLogoStorage_EnsureDataDirCreatesConfigMap(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	storage := NewConfigMapLogoStorage(client, "dkonsole")

	ctx := context.Background()
	if err := storage.EnsureDataDir(ctx); err != nil {
		t.Fatalf("EnsureDataDir error: %v", err)
	}

	if _, err := client.CoreV1().ConfigMaps("dkonsole").Get(ctx, ConfigMapName, metav1.GetOptions{}); err != nil {
		t.Fatalf("expected ConfigMap to be created: %v", err)
	}
}

func TestConfigMapLogoStorage_SaveAndGet(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	storage := NewConfigMapLogoStorage(client, "dkonsole")
	ctx := context.Background()

	content := []byte{0x1, 0x2, 0x3}
	if err := storage.Save(ctx, "normal", ".png", bytes.NewReader(content)); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	encoded, err := storage.Get(ctx, "normal", ".png")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if encoded != base64.StdEncoding.EncodeToString(content) {
		t.Fatalf("unexpected stored content: %s", encoded)
	}

	decoded, err := storage.GetLogoContent(ctx, "normal", ".png")
	if err != nil {
		t.Fatalf("GetLogoContent error: %v", err)
	}
	if !bytes.Equal(decoded, content) {
		t.Fatalf("decoded content mismatch: %v", decoded)
	}
}

func TestConfigMapLogoStorage_SaveUpdatesExisting(t *testing.T) {
	data := map[string]string{"logo.png": base64.StdEncoding.EncodeToString([]byte("old"))}
	client := k8sfake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: "dkonsole",
		},
		Data: data,
	})
	storage := NewConfigMapLogoStorage(client, "dkonsole")
	ctx := context.Background()

	if err := storage.Save(ctx, "normal", ".png", bytes.NewReader([]byte("new"))); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	cm, _ := client.CoreV1().ConfigMaps("dkonsole").Get(ctx, ConfigMapName, metav1.GetOptions{})
	if cm.Data["logo.png"] != base64.StdEncoding.EncodeToString([]byte("new")) {
		t.Fatalf("expected updated content, got %s", cm.Data["logo.png"])
	}
}

func TestConfigMapLogoStorage_RemoveAll(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: "dkonsole",
		},
		Data: map[string]string{
			"logo.png":       "a",
			"logo.svg":       "b",
			"logo-light.png": "c",
			"other":          "d",
		},
	})
	storage := NewConfigMapLogoStorage(client, "dkonsole")
	ctx := context.Background()

	if err := storage.RemoveAll(ctx, "light"); err != nil {
		t.Fatalf("RemoveAll error: %v", err)
	}

	cm, _ := client.CoreV1().ConfigMaps("dkonsole").Get(ctx, ConfigMapName, metav1.GetOptions{})
	if _, ok := cm.Data["logo-light.png"]; ok {
		t.Fatalf("expected light logo key removed")
	}
	if _, ok := cm.Data["logo-light.svg"]; ok {
		t.Fatalf("expected light logo svg key removed")
	}
	if cm.Data["logo.png"] == "" || cm.Data["logo.svg"] == "" {
		t.Fatalf("non-light logos should remain")
	}
	if cm.Data["other"] == "" {
		t.Fatalf("non logo keys should remain")
	}
}

func TestConfigMapLogoStorage_ListKeysAndByKey(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: "dkonsole",
		},
		Data: map[string]string{
			"logo.png":       base64.StdEncoding.EncodeToString([]byte("png")),
			"logo-light.svg": base64.StdEncoding.EncodeToString([]byte("svg")),
			"skip":           "noop",
		},
	})
	storage := NewConfigMapLogoStorage(client, "dkonsole")
	ctx := context.Background()

	keys, err := storage.ListLogoKeys(ctx)
	if err != nil {
		t.Fatalf("ListLogoKeys error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}

	content, err := storage.GetLogoContentByKey(ctx, "logo-light.svg")
	if err != nil {
		t.Fatalf("GetLogoContentByKey error: %v", err)
	}
	if string(content) != "svg" {
		t.Fatalf("unexpected content: %s", string(content))
	}
}
