package config

import (
	"net"
	"time"

	"github.com/bitcoin-sv/pulse/domains/logging"
)

// DBSqlite creating config for sqlite db.
const DBSqlite DbType = "sqlite"

func GetDefaultAppConfig(lf logging.LoggerFactory) *AppConfig {
	return &AppConfig{
		DbConfig:         getDbDefaults(),
		HTTPConfig:       getAuthConfigDefaults(),
		MerkleRootConfig: getMerkleRootDefaults(),
		WebsocketConfig:  getWebsocketDefaults(),
		WebhookConfig:    getWebhookDefaults(),
		P2PConfig:        getP2PDefaults(lf),
	}
}

func getDbDefaults() *DbConfig {
	return &DbConfig{
		Type:               DBSqlite,
		FilePath:           "./data/blockheaders.db",
		Dsn:                "file:./data/blockheaders.db?_foreign_keys=true&pooling=true",
		SchemaPath:         "./database/migrations",
		PreparedDb:         false,
		PreparedDbFilePath: "./data/blockheaders.csv.gz",
	}
}

func getAuthConfigDefaults() *HTTPConfig {
	return &HTTPConfig{
		ReadTimeout:  10,
		WriteTimeout: 10,
		Port:         8080,
		UrlPrefix:    "/api/v1",
		UseAuth:      true,
		AuthToken:    "mQZQ6WmxURxWz5ch",
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

func getP2PDefaults(lf logging.LoggerFactory) *P2PConfig {
	return &P2PConfig{
		LogLevel:                  "info",
		MaxPeers:                  125,
		MaxPeersPerIP:             5,
		BanDuration:               time.Hour * 24,
		MinSyncPeerNetworkSpeed:   51200,
		ExcessiveBlockSize:        128000000,
		TrickleInterval:           50 * time.Millisecond,
		BlocksForForkConfirmation: 10,
		DefaultConnectTimeout:     30 * time.Second,
		DisableCheckpoints:        false,
		Dial:                      net.DialTimeout,
		Lookup:                    net.LookupIP,
		TimeSource:                NewMedianTime(lf),
		Checkpoints:               ActiveNetParams.Checkpoints,
	}
}
