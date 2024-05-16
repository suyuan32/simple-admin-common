package i18n

// Conf is the configuration structure for i18n
type Conf struct {
	Dir string `json:",env=I18N_DIR"`
	// BundleFilePaths store the paths of i18n bundles.
	BundleFilePaths []string `json:",optional,env=BUNDLE_FILE_PATHS"`
	// SupportLanguages store the languages you want to support, the language order must the same as the files' order
	// in BundleFilePaths.
	SupportLanguages []string `json:",optional,env=SUPPORT_LANGUAGES"`
}
