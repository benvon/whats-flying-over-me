package cataloger

import (
	"testing"
	"time"
)

func TestNewCataloger(t *testing.T) {
	tests := []struct {
		name    string
		config  ElasticSearchConfig
		wantErr bool
	}{
		{
			name: "disabled cataloger",
			config: ElasticSearchConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled with URL",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "enabled without URL",
			config: ElasticSearchConfig{
				Enabled: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCataloger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCataloger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCatalogerFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ElasticSearchConfig
		wantErr bool
	}{
		{
			name: "disabled cataloger",
			config: ElasticSearchConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled with URL",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "enabled without URL",
			config: ElasticSearchConfig{
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "enabled with default values",
			config: ElasticSearchConfig{
				Enabled: true,
				URL:     "http://localhost:9200",
				Index:   "custom_index",
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCatalogerFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCatalogerFromConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
