// Generated code; DO NOT EDIT.

package config

// Reset reset fields RetryConfig to their initial values.
func (r *RetryConfig) Reset() {
	r.MaxAttempts = 0
	r.InitialBackoff = 0
	r.BackoffMultiplier = 0
}

// Reset reset fields ServerConfig to their initial values.
func (s *ServerConfig) Reset() {
	s.RunAddr = ""
	s.LogLevel = ""
	s.Environment = ""
	s.StoreInterval = 0
	s.StorePath = ""
	s.Restore = false
	s.DBDataSourceName = ""
	s.HMACKey = ""
	s.AuditFile = ""
	s.AuditURL = ""
	s.ProfileEnabled = false
}
