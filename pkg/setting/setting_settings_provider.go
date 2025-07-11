package setting

func (cfg *Cfg) setSettingsProviderConfig() {
	section := cfg.Raw.Section("config_provider")
	allowAdminAccess, err := section.GetKey("allow_admin_access")
	if err != nil {
		return
	}
	cfg.SettingsAllowAdminAccess = allowAdminAccess.MustBool(false)
}
