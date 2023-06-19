package config

// CROSConf stores the configuration for cross domain
type CROSConf struct {
	Address string `json:",env=CROS_ADDRESS"`
}
