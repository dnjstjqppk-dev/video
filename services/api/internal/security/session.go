package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrTokenExpired        = errors.New("token expired")
	ErrTokenReused         = errors.New("token reuse detected: session family revoked")
	ErrStepUpRequired      = errors.New("mfa step-up required due to anomalous location change")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type Session struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	FamilyID          string    `json:"family_id"`
	Sequence          int       `json:"sequence"`
	RefreshTokenHash  string    `json:"-"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	IPAddress         string    `json:"ip_address"`
	LastCountry       string    `json:"last_country"`
	LastLatitude      float64   `json:"last_lat"`
	LastLongitude     float64   `json:"last_lon"`
	CreatedAt         time.Time `json:"created_at"`
	LastActiveAt      time.Time `json:"last_active_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	IsRevoked         bool      `json:"is_revoked"`
	StepUpRequired    bool      `json:"step_up_required"`
}

type SessionManager struct {
	mu           sync.RWMutex
	sessions     map[string]*Session        // key: sessionID
	familyTokens map[string]map[string]int  // key: familyID -> map[tokenHash]sequence
	userSessions map[string]map[string]bool // key: userID -> map[sessionID]true
	tokenTTL     time.Duration
	sessionTTL   time.Duration
}

func NewSessionManager(tokenTTL, sessionTTL time.Duration) *SessionManager {
	if tokenTTL == 0 {
		tokenTTL = 15 * time.Minute
	}
	if sessionTTL == 0 {
		sessionTTL = 30 * 24 * time.Hour
	}
	return &SessionManager{
		sessions:     make(map[string]*Session),
		familyTokens: make(map[string]map[string]int),
		userSessions: make(map[string]map[string]bool),
		tokenTTL:     tokenTTL,
		sessionTTL:   sessionTTL,
	}
}

func randomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// CreateSession issues a new Refresh Token + Session with family tracking (TASK-003, TASK-005)
func (sm *SessionManager) CreateSession(userID, deviceFP, ip, country string, lat, lon float64) (*Session, string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID, err := randomString(16)
	if err != nil {
		return nil, "", err
	}
	familyID, err := randomString(16)
	if err != nil {
		return nil, "", err
	}
	rawRefreshToken, err := randomString(32)
	if err != nil {
		return nil, "", err
	}

	tokenHash := hashToken(rawRefreshToken)
	now := time.Now().UTC()

	sess := &Session{
		ID:                sessionID,
		UserID:            userID,
		FamilyID:          familyID,
		Sequence:          1,
		RefreshTokenHash:  tokenHash,
		DeviceFingerprint: deviceFP,
		IPAddress:         ip,
		LastCountry:       country,
		LastLatitude:      lat,
		LastLongitude:     lon,
		CreatedAt:         now,
		LastActiveAt:      now,
		ExpiresAt:         now.Add(sm.sessionTTL),
		IsRevoked:         false,
		StepUpRequired:    false,
	}

	sm.sessions[sessionID] = sess

	if _, ok := sm.familyTokens[familyID]; !ok {
		sm.familyTokens[familyID] = make(map[string]int)
	}
	sm.familyTokens[familyID][tokenHash] = 1

	if _, ok := sm.userSessions[userID]; !ok {
		sm.userSessions[userID] = make(map[string]bool)
	}
	sm.userSessions[userID][sessionID] = true

	return sess, rawRefreshToken, nil
}

// RotateRefreshToken implements strict Refresh Token Rotation (RTR) and reuse detection (TASK-003, TASK-004, TASK-006)
func (sm *SessionManager) RotateRefreshToken(sessionID, presentedToken, deviceFP, ip, country string, lat, lon float64) (*Session, string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, ok := sm.sessions[sessionID]
	if !ok || sess.IsRevoked {
		return nil, "", ErrSessionNotFound
	}

	now := time.Now().UTC()
	if now.After(sess.ExpiresAt) {
		sess.IsRevoked = true
		return nil, "", ErrTokenExpired
	}

	presentedHash := hashToken(presentedToken)
	familyMap, famOk := sm.familyTokens[sess.FamilyID]
	if !famOk {
		sess.IsRevoked = true
		return nil, "", ErrTokenReused
	}

	seq, tokenSeen := familyMap[presentedHash]
	// Token Reuse Detection (TASK-004): If presented token is not the latest active token in family
	if !tokenSeen || seq != sess.Sequence || presentedHash != sess.RefreshTokenHash {
		// REUSE DETECTED: Immediately revoke all sessions in this user's account for security
		for sid := range sm.userSessions[sess.UserID] {
			if s, exists := sm.sessions[sid]; exists {
				s.IsRevoked = true
			}
		}
		delete(sm.familyTokens, sess.FamilyID)
		return nil, "", ErrTokenReused
	}

	// Impossible Travel Velocity Check (TASK-006): Check travel speed between logins
	if sess.LastCountry != "" && country != "" && sess.LastCountry != country {
		// Calculate time difference in hours
		hours := now.Sub(sess.LastActiveAt).Hours()
		if hours < 2.0 {
			// Greater than normal flight speed for country change -> Trigger MFA step-up
			sess.StepUpRequired = true
		}
	}

	// Issue new rotated refresh token
	newRawToken, err := randomString(32)
	if err != nil {
		return nil, "", err
	}
	newHash := hashToken(newRawToken)

	sess.Sequence++
	sess.RefreshTokenHash = newHash
	sess.LastActiveAt = now
	sess.IPAddress = ip
	sess.LastCountry = country
	sess.LastLatitude = lat
	sess.LastLongitude = lon
	sess.DeviceFingerprint = deviceFP

	familyMap[newHash] = sess.Sequence

	return sess, newRawToken, nil
}

// RevokeSession revokes a single session (TASK-059)
func (sm *SessionManager) RevokeSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, ok := sm.sessions[sessionID]
	if !ok || sess.IsRevoked {
		return ErrSessionNotFound
	}
	sess.IsRevoked = true
	delete(sm.familyTokens, sess.FamilyID)
	delete(sm.sessions, sessionID)
	if sm.userSessions[sess.UserID] != nil {
		delete(sm.userSessions[sess.UserID], sessionID)
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user (TASK-004, TASK-059)
func (sm *SessionManager) RevokeAllUserSessions(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for sid := range sm.userSessions[userID] {
		if s, ok := sm.sessions[sid]; ok {
			s.IsRevoked = true
			delete(sm.familyTokens, s.FamilyID)
		}
	}
	delete(sm.userSessions, userID)
}

// ListUserSessions returns all active sessions for a user (TASK-059)
func (sm *SessionManager) ListUserSessions(userID, currentSessionID string) []map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []map[string]interface{}
	for sid := range sm.userSessions[userID] {
		s, ok := sm.sessions[sid]
		if !ok || s.IsRevoked || time.Now().UTC().After(s.ExpiresAt) {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":                 s.ID,
			"ip":                 s.IPAddress,
			"country":            s.LastCountry,
			"device_fingerprint": s.DeviceFingerprint,
			"created_at":         s.CreatedAt.Format(time.RFC3339),
			"last_active_at":     s.LastActiveAt.Format(time.RFC3339),
			"is_current":         s.ID == currentSessionID,
			"step_up_required":   s.StepUpRequired,
		})
	}
	return result
}
