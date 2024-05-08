package config

import (
	"time"
)

// #nosec G101
const DefaultAppToken = "mQZQ6WmxURxWz5ch"

func GetDefaultAppConfig() *AppConfig {
	return &AppConfig{
		Db:         getDbDefaults(),
		HTTP:       getHttpConfigDefaults(),
		MerkleRoot: getMerkleRootDefaults(),
		Websocket:  getWebsocketDefaults(),
		Webhook:    getWebhookDefaults(),
		P2P:        getP2PDefaults(),
		Logging:    getLoggingDefaults(),
		Metrics:    getMetricsDefaults(),
	}
}

func getDbDefaults() *DbConfig {
	return &DbConfig{
		Engine:             DBSqlite,
		SchemaPath:         "./database/migrations",
		PreparedDb:         false,
		PreparedDbFilePath: "./data/blockheaders.csv.gz",
		Sqlite: SqliteConfig{
			FilePath: "./data/blockheaders.db",
		},
		Postgres: getPostgresDefaults(),
	}
}

func getHttpConfigDefaults() *HTTPConfig {
	return &HTTPConfig{
		ReadTimeout:               10,
		WriteTimeout:              10,
		Port:                      8080,
		UseAuth:                   true,
		AuthToken:                 DefaultAppToken,
		ProfilingEndpointsEnabled: true,
	}
}

func getMerkleRootDefaults() *MerkleRootConfig {
	return &MerkleRootConfig{
		MaxBlockHeightExcess: 6,
	}
}

func getWebsocketDefaults() *WebsocketConfig {
	return &WebsocketConfig{
		HistoryMax: 300,
		HistoryTTL: 10,
	}
}

func getWebhookDefaults() *WebhookConfig {
	return &WebhookConfig{
		MaxTries: 10,
	}
}

func getP2PDefaults() *P2PConfig {
	return &P2PConfig{
		BanDuration:               time.Hour * 24,
		BlocksForForkConfirmation: 10,
		DefaultConnectTimeout:     30 * time.Second,
		DisableCheckpoints:        false,
		UserAgentName:             ApplicationName,
		UserAgentVersion:          Version(),
		Experimental:              false,
		MaxOutboundConnections:    8,
		MaxInboundConnections:     8,
		AcceptLocalPeers:          false,
	}
}

func getLoggingDefaults() *LoggingConfig {
	return &LoggingConfig{
		Level:        "debug",
		Format:       "console",
		InstanceName: ApplicationName,
		LogOrigin:    true,
	}
}

func getMetricsDefaults() *MetricsConfig {
	return &MetricsConfig{
		Enabled: false,
	}
}

func getPostgresDefaults() PostgreSqlConfig {
	return PostgreSqlConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DbName:   "bhs",
		Sslmode:  "disable",
	}
}
