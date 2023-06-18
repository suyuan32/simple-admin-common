package config

type CROSConf struct {
	Address string `json:",default=*,env=CROS_ADDRESS"`
}
