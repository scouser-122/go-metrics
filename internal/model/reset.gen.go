// Generated code; DO NOT EDIT.

package models

// Reset reset fields Metrics to their initial values.
func (m *Metrics) Reset() {
	m.ID = ""
	m.MType = ""
	if m.Delta != nil {
		(*m.Delta) = 0
	}
	if m.Value != nil {
		(*m.Value) = 0
	}
	m.Hash = ""
}

// Reset reset fields MetricsReceivedEvent to their initial values.
func (m *MetricsReceivedEvent) Reset() {
	m.TS = 0
	m.Metrics = m.Metrics[:0]
	m.IPAddress = ""
}

// Reset reset fields ResponsePayload to their initial values.
func (r *ResponsePayload) Reset() {
	r.Status = ""
	r.Message = ""
}
