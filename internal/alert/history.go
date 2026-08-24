package alert

// History returns the alerts raised for one device.
func (s *Service) History(deviceID string) []*Alert {
	var history []*Alert
	for _, a := range s.All() {
		if a.DeviceID == deviceID {
			history = append(history, a)
		}
	}
	return history
}
