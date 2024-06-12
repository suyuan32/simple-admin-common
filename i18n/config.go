package i18n

// Conf is the configuration structure for i18n
type Conf struct {
	Dir string `json:",env=I18N_DIR"`
}
