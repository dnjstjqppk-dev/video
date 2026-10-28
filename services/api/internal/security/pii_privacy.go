package security

import (
	"strings"
	"time"
)

// MaskEmail masks an email address for privacy display (TASK-019)
// e.g. "alexander@bookingnetwork.com" -> "a***r@bookingnetwork.com"
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	domain := parts[1]

	if len(name) <= 2 {
		return name[:1] + "***@" + domain
	}
	return string(name[0]) + "***" + string(name[len(name)-1]) + "@" + domain
}

// MaskPhone masks phone numbers (TASK-019)
// e.g. "+82-10-1234-5678" -> "+82-10-****-5678"
func MaskPhone(phone string) string {
	parts := strings.Split(phone, "-")
	if len(parts) >= 2 {
		// Mask the second-to-last group (the main number block), keeping country/area code and last 4 digits
		parts[len(parts)-2] = "****"
		return strings.Join(parts, "-")
	}
	if len(phone) < 8 {
		return "****"
	}
	return phone[:len(phone)-8] + "****" + phone[len(phone)-4:]
}

// MaskTaxID masks sensitive business / personal tax IDs (TASK-019)
func MaskTaxID(taxID string) string {
	if len(taxID) < 6 {
		return "***-**-*****"
	}
	return taxID[:3] + "-**-*****"
}

// COPPAEnforcement encapsulates COPPA child-safety restrictions (TASK-022)
type COPPAEnforcement struct {
	IsForKids              bool `json:"is_for_kids"`
	CommentsAllowed        bool `json:"comments_allowed"`
	PersonalizedAdsAllowed bool `json:"personalized_ads_allowed"`
	NotificationPrompts    bool `json:"notification_prompts_allowed"`
	LiveChatAllowed        bool `json:"live_chat_allowed"`
}

// ApplyCOPPARules enforces COPPA compliance when content is marked for children (TASK-022)
func ApplyCOPPARules(isForKids bool) COPPAEnforcement {
	if isForKids {
		return COPPAEnforcement{
			IsForKids:              true,
			CommentsAllowed:        false,
			PersonalizedAdsAllowed: false,
			NotificationPrompts:    false,
			LiveChatAllowed:        false,
		}
	}
	return COPPAEnforcement{
		IsForKids:              false,
		CommentsAllowed:        true,
		PersonalizedAdsAllowed: true,
		NotificationPrompts:    true,
		LiveChatAllowed:        true,
	}
}

// AccountDeletionRequest represents a GDPR "Right to be Forgotten" request with 30-day grace period (TASK-021)
type AccountDeletionRequest struct {
	UserID            string    `json:"user_id"`
	RequestedAt       time.Time `json:"requested_at"`
	PermanentDeleteAt time.Time `json:"permanent_delete_at"`
	IsCancelled       bool      `json:"is_cancelled"`
}

func NewAccountDeletionRequest(userID string) AccountDeletionRequest {
	now := time.Now().UTC()
	return AccountDeletionRequest{
		UserID:            userID,
		RequestedAt:       now,
		PermanentDeleteAt: now.Add(30 * 24 * time.Hour), // 30-day grace period
		IsCancelled:       false,
	}
}
