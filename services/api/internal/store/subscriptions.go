package store

// Notify levels for channel subscriptions: all videos, personalized
// highlights, or muted. Stored per subscription; fan-out honors the level.
const (
	NotifyAll    = "all"
	NotifyCustom = "custom"
	NotifyNone   = "none"
)

// ValidNotifyLevel reports whether a subscription notify level is known.
func ValidNotifyLevel(level string) bool {
	return level == NotifyAll || level == NotifyCustom || level == NotifyNone
}

// SubNotifyLevel returns the stored notify level, defaulting to "all" for
// subscriptions without an explicit preference, "" when not subscribed.
func (s *MemoryStore) SubNotifyLevel(userID, channelID string) string {
	subs, ok := s.Subscriptions[userID]
	if !ok || !subs[channelID] {
		return ""
	}
	if notify, ok := s.SubNotify[userID]; ok {
		if level, ok := notify[channelID]; ok && ValidNotifyLevel(level) {
			return level
		}
	}
	return NotifyAll
}

// SetSubNotify stores a notify level for an active subscription.
func (s *MemoryStore) SetSubNotify(userID, channelID, level string) {
	if s.SubNotify == nil {
		s.SubNotify = make(map[string]map[string]string)
	}
	if s.SubNotify[userID] == nil {
		s.SubNotify[userID] = make(map[string]string)
	}
	s.SubNotify[userID][channelID] = level
}

// ClearSubNotify drops the stored preference (called on unsubscribe).
func (s *MemoryStore) ClearSubNotify(userID, channelID string) {
	if notify, ok := s.SubNotify[userID]; ok {
		delete(notify, channelID)
	}
}
