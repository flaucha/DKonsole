package api

import (
	"context"
	"fmt"
)

// CRDService provides business logic for CRD operations
type CRDService struct {
	repo CRDRepository
}

// NewCRDService creates a new CRDService
func NewCRDService(repo CRDRepository) *CRDService {
	return &CRDService{repo: repo}
}

// CRD represents a Custom Resource Definition
type CRD struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Scope   string `json:"scope"`
}

// GetCRDsRequest represents parameters for getting CRDs
type GetCRDsRequest struct {
	Limit         int64
	ContinueToken string
}

// GetCRDsResponse represents the paginated response for CRDs
type GetCRDsResponse struct {
	CRDs     []CRD  `json:"crds"`
	Continue string `json:"continue,omitempty"`
}

// GetCRDs returns all Custom Resource Definitions
func (s *CRDService) GetCRDs(ctx context.Context, req GetCRDsRequest) (*GetCRDsResponse, error) {
	crdList, err := s.repo.ListCRDs(ctx, req.Limit, req.ContinueToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}
	var crds []CRD
	for _, item := range crdList.Items {
		spec, ok := item.Object["spec"].(map[string]interface{})
		if !ok {
			continue
		}
		group, _ := spec["group"].(string)
		names, _ := spec["names"].(map[string]interface{})
		kind, _ := names["kind"].(string)
		scope, _ := spec["scope"].(string)
		versions, _ := spec["versions"].([]interface{})
		for _, v := range versions {
			if version, ok := v.(map[string]interface{}); ok {
				versionName, _ := version["name"].(string)
				served, _ := version["served"].(bool)
				if served && group != "" && kind != "" {
					crds = append(crds, CRD{
						Name:    item.GetName(),
						Group:   group,
						Version: versionName,
						Kind:    kind,
						Scope:   scope,
					})
				}
			}
		}
	}
	return &GetCRDsResponse{
		CRDs:     crds,
		Continue: crdList.GetContinue(),
	}, nil
}
