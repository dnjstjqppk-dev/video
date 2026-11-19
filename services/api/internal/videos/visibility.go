package videos

import (
	"globalvideoplatform/services/api/internal/store"
)

// ViewVerdict classifies whether a viewer may access a video.
type ViewVerdict int

const (
	// ViewAllow grants access.
	ViewAllow ViewVerdict = iota
	// ViewHide conceals existence (respond 404).
	ViewHide
	// ViewDeny discloses existence but refuses playback (respond 403).
	ViewDeny
)

// IsPlatformOperator reports whether the user bypasses visibility gates.
func IsPlatformOperator(user *store.User) bool {
	return user != nil && (user.Role == "super_admin" || user.Role == "platform_admin")
}

// IsChannelOwner reports whether the user owns the video's channel.
func IsChannelOwner(user *store.User, video *store.Video, channels map[string]*store.Channel) bool {
	if user == nil || video == nil {
		return false
	}
	channel, ok := channels[video.ChannelID]
	return ok && channel != nil && channel.OwnerID == user.ID
}

// HasMembership reports whether the user holds a paid membership tier.
func HasMembership(user *store.User) bool {
	if user == nil {
		return false
	}
	if user.Role == "premium_member" {
		return true
	}
	switch user.SubscriptionTier {
	case "premium", "enterprise":
		return true
	default:
		return false
	}
}

// IsEntitled reports whether the user may view members-only content.
func IsEntitled(user *store.User, video *store.Video, channels map[string]*store.Channel) bool {
	return IsPlatformOperator(user) || IsChannelOwner(user, video, channels) || HasMembership(user)
}

// ClassifyViewAccess maps video visibility to an access verdict.
// Unknown visibility values fail closed to ViewHide.
func ClassifyViewAccess(user *store.User, video *store.Video, channels map[string]*store.Channel) ViewVerdict {
	switch video.Visibility {
	case "public", "unlisted":
		return ViewAllow
	case "private", "scheduled":
		if IsPlatformOperator(user) || IsChannelOwner(user, video, channels) {
			return ViewAllow
		}
		return ViewHide
	case "members_only":
		if IsEntitled(user, video, channels) {
			return ViewAllow
		}
		return ViewDeny
	default:
		return ViewHide
	}
}

// SearchVisible reports whether a video may appear in public search results
// and suggestions. Private and unlisted videos never appear; members-only
// videos appear only for entitled viewers.
func SearchVisible(user *store.User, video *store.Video, channels map[string]*store.Channel) bool {
	if video == nil || video.Status == "DELETED" {
		return false
	}
	switch video.Visibility {
	case "public":
		return true
	case "members_only":
		return IsEntitled(user, video, channels)
	default:
		return false
	}
}
