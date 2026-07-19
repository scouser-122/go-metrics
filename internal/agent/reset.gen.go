// Generated code; DO NOT EDIT.

package agent

// Reset reset fields AgentConfig to their initial values.
func (a *AgentConfig) Reset() {
	a.RuntimeMetricNames = a.RuntimeMetricNames[:0]
	a.ServerAddress = ""
	a.ReportInterval = 0
	a.PollInterval = 0
	a.LogLevel = ""
	a.Environment = ""
	a.HMACKey = ""
	a.RequestRateLimit = 0
	a.CollectChannelSize = 0
}

// Reset reset fields CollectedData to their initial values.
func (c *CollectedData) Reset() {
	c.metrics = c.metrics[:0]
}
