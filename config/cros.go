package config

type CROSConf struct {
	Address string `json:",env=CROS_ADDRESS"`
}
