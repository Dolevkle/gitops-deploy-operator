package k8s

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sManager Manages Kubernetes resources.
type K8sManager struct {
	client.Client
	Scheme *runtime.Scheme
}

func NewK8sManager(c client.Client, scheme *runtime.Scheme) *K8sManager {
	return &K8sManager{
		Client: c,
		Scheme: scheme,
	}
}

// ApplyManifests Applies Kubernetes manifests to the cluster.
func (k *K8sManager) ApplyManifests(ctx context.Context, dir string, logger logr.Logger) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		logger.Info("Applying manifest", "path", path)
		data, err := fs.ReadFile(os.DirFS("/"), path)
		if err != nil {
			return err
		}
		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024)
		for {
			obj := &unstructured.Unstructured{}
			if err := decoder.Decode(obj); err != nil {
				break
			}
			if err := k.applyObject(ctx, obj); err != nil {
				return err
			}
		}
		return nil
	})
}

// Creates or updates a Kubernetes resource.
func (k *K8sManager) applyObject(ctx context.Context, obj *unstructured.Unstructured) error {
	obj.SetNamespace("default")
	key := client.ObjectKeyFromObject(obj)
	existing := obj.DeepCopy()
	err := k.Get(ctx, key, existing)
	if err != nil {
		// Create the object
		return k.Create(ctx, obj)
	}
	// Update the object
	obj.SetResourceVersion(existing.GetResourceVersion())
	return k.Update(ctx, obj)
}

// DeleteManifests Deletes Kubernetes resources defined in the manifests.
func (k *K8sManager) DeleteManifests(ctx context.Context, dir string, logger logr.Logger) error {
	// Similar logic to ApplyManifests, but call Delete instead
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		logger.Info("Deleting manifest", "path", path)
		data, err := fs.ReadFile(os.DirFS("/"), path)
		if err != nil {
			return err
		}
		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024)
		for {
			obj := &unstructured.Unstructured{}
			if err := decoder.Decode(obj); err != nil {
				break
			}
			if err := k.deleteObject(ctx, obj); err != nil {
				return err
			}
		}
		return nil
	})
}

// Deletes a Kubernetes resource.
func (k *K8sManager) deleteObject(ctx context.Context, obj *unstructured.Unstructured) error {
	return k.Delete(ctx, obj)
}
