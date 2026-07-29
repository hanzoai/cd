package util

import (
	"fmt"
	"net/url"
	"os"

	"github.com/hanzoai/cd/gitops-engine/pkg/utils/kube"

	appv1alpha1 "github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	"github.com/hanzoai/cd/util/config"
)

func ConstructApplicationSet(fileURL string) ([]*appv1alpha1.ApplicationSet, error) {
	if fileURL != "" {
		return constructAppsetFromFileURL(fileURL)
	}
	return nil, nil
}

func constructAppsetFromFileURL(fileURL string) ([]*appv1alpha1.ApplicationSet, error) {
	appset := make([]*appv1alpha1.ApplicationSet, 0)
	// read uri
	err := readAppsetFromURI(fileURL, &appset)
	if err != nil {
		return nil, fmt.Errorf("error reading applicationset from file %s: %w", fileURL, err)
	}

	return appset, nil
}

func readAppsetFromURI(fileURL string, appset *[]*appv1alpha1.ApplicationSet) error {
	readFilePayload := func() ([]byte, error) {
		parsedURL, err := url.ParseRequestURI(fileURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return os.ReadFile(fileURL)
		}
		return config.ReadRemoteFile(fileURL)
	}

	yml, err := readFilePayload()
	if err != nil {
		return fmt.Errorf("error reading file payload: %w", err)
	}

	return readAppset(yml, appset)
}

func readAppset(yml []byte, appsets *[]*appv1alpha1.ApplicationSet) error {
	yamls, err := kube.SplitYAMLToString(yml)
	if err != nil {
		return fmt.Errorf("error splitting YAML to string: %w", err)
	}

	for _, yml := range yamls {
		var appset appv1alpha1.ApplicationSet
		err = config.Unmarshal([]byte(yml), &appset)
		if err != nil {
			return fmt.Errorf("error unmarshalling appset: %w", err)
		}
		*appsets = append(*appsets, &appset)
	}
	// we reach here if there is no error found while reading the Application Set
	return nil
}
