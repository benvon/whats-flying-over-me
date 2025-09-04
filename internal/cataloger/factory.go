package cataloger

import (
	"fmt"
)

// NewCataloger creates a new cataloger based on the configuration
func NewCataloger(config ElasticSearchConfig) (Cataloger, error) {
	if !config.Enabled {
		return &NoOpCataloger{}, nil
	}

	// Currently only ElasticSearch is supported
	return NewElasticSearchCataloger(config)
}

// NewCatalogerFromConfig creates a cataloger from a configuration struct
func NewCatalogerFromConfig(config ElasticSearchConfig) (Cataloger, error) {
	if !config.Enabled {
		return &NoOpCataloger{}, nil
	}

	if config.URL == "" {
		return nil, fmt.Errorf("ElasticSearch URL is required when cataloging is enabled")
	}

	return NewElasticSearchCataloger(config)
}
