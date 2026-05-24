package generator

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateVerificationToken produces a short unique token suitable for
// email verification. The format is:
//
//	<userID>-<unixNano>-<randomHex>
//
// Example:
//
//	6f1e2d3c-...-abcd1234-1616161616161616-7f3a9b...
//
// The token embeds the user ID and a timestamp so you can validate
// and optionally check expiration without requiring an extra DB field.
// Still, it's common to store the token in the database so it can be
// invalidated after use.
func GenerateVerificationToken(userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("userID is required")
	}

	// 16 bytes of randomness (32 hex chars)
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	randHex := hex.EncodeToString(b)

	ts := strconv.FormatInt(time.Now().UnixNano(), 10)

	// join with '-' which is safe for URLs and emails
	token := fmt.Sprintf("%s-%s-%s", userID.String(), ts, randHex)
	return token, nil
}

// ParseVerificationToken splits a token created by GenerateVerificationToken
// and returns the userID, the timestamp (as time.Time) and the random hex part.
func ParseVerificationToken(token string) (uuid.UUID, time.Time, string, error) {
	parts := strings.Split(token, "-")
	if len(parts) < 3 {
		return uuid.Nil, time.Time{}, "", errors.New("invalid token format")
	}

	// userID may itself contain dashes; the timestamp is the second-to-last
	// and the random hex is the last element. Re-join the remaining parts for userID.
	randHex := parts[len(parts)-1]
	tsStr := parts[len(parts)-2]
	userIDParts := parts[:len(parts)-2]
	userIDStr := strings.Join(userIDParts, "-")

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, time.Time{}, "", fmt.Errorf("parse user id: %w", err)
	}

	tsInt, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return uuid.Nil, time.Time{}, "", fmt.Errorf("parse timestamp: %w", err)
	}

	return uid, time.Unix(0, tsInt), randHex, nil
}

// TokenExpired returns true if the token is older than the provided expiry duration.
// It returns an error if the token format is invalid.
func TokenExpired(token string, expiry time.Duration) (bool, error) {
	_, ts, _, err := ParseVerificationToken(token)
	if err != nil {
		return false, err
	}
	return time.Since(ts) > expiry, nil
}
