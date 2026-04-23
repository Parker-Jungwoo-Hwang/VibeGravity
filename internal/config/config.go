// Package config defines VibeGravity runtime configuration.
package config

// Config contains settings shared by the server, worker, and CLI.
type Config struct {
	DatabaseURL       string `yaml:"database_url"`
	MigrationPath     string `yaml:"migration_path"`
	EmbeddingEndpoint string `yaml:"embedding_endpoint"`
	EmbeddingModel    string `yaml:"embedding_model"`
	EmbeddingDims     int    `yaml:"embedding_dims"`
}

// NewDefaultConfig returns conservative defaults before environment loading.
func NewDefaultConfig() Config {
	return Config{
		MigrationPath:  "migrations",
		EmbeddingModel: "pending",
		EmbeddingDims:  0,
	}
}
