package applier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

type Applier struct {
	dynClient dynamic.Interface
	mapper    meta.RESTMapper
	logger    *slog.Logger
}

func buildRestConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{},
	).ClientConfig()
}

func New(logger *slog.Logger) (*Applier, error) {
	cfg, err := buildRestConfig()
	if err != nil {
		return nil, fmt.Errorf("rest config: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("discovery client: %w", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("dynamic client: %w", err)
	}

	return &Applier{
		dynClient: dynClient,
		mapper:    mapper,
		logger:    logger,
	}, nil
}

func (a *Applier) Apply(ctx context.Context, dirPath string) error {
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	resMap, err := k.Run(filesys.MakeFsOnDisk(), dirPath)
	if err != nil {
		return fmt.Errorf("kustomize build %s: %w", dirPath, err)
	}

	for _, res := range resMap.Resources() {
		yamlBytes, err := res.AsYAML()
		if err != nil {
			a.logger.Error("resource to yaml", "name", res.GetName(), "kind", res.GetKind(), "err", err)
			continue
		}
		if err := a.applyOne(ctx, yamlBytes); err != nil {
			a.logger.Error("apply resource", "name", res.GetName(), "kind", res.GetKind(), "err", err)
		}
	}
	return nil
}

func (a *Applier) applyOne(ctx context.Context, yamlBytes []byte) error {
	jsonBytes, err := utilyaml.ToJSON(yamlBytes)
	if err != nil {
		return fmt.Errorf("yaml to json: %w", err)
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonBytes, &obj.Object); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	gvk := obj.GroupVersionKind()
	mapping, err := a.mapper.RESTMapping(
		schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind},
		gvk.Version,
	)
	if err != nil {
		return fmt.Errorf("rest mapping for %s/%s: %w", gvk.Group, gvk.Kind, err)
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := obj.GetNamespace()
		if ns == "" {
			ns = "default"
		}
		dr = a.dynClient.Resource(mapping.Resource).Namespace(ns)
	} else {
		dr = a.dynClient.Resource(mapping.Resource)
	}

	force := true
	_, err = dr.Patch(
		ctx,
		obj.GetName(),
		apitypes.ApplyPatchType,
		jsonBytes,
		metav1.PatchOptions{
			FieldManager: "kustomize-watcher",
			Force:        &force,
		},
	)
	return err
}
