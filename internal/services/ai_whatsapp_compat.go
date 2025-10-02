package services

import (
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"
)

// BalasHelper provides compatibility for balas field (int vs string)
type BalasHelper struct{}

// GetBalasAsString converts balas int to string timestamp
func (b *BalasHelper) GetBalasAsString(balas int) string {
	if balas == 0 {
		return ""
	}
	// Assume balas stores Unix timestamp if > 0
	return time.Unix(int64(balas), 0).Format("2006-01-02 15:04:05")
}

// SetBalasAsInt converts timestamp string to int (Unix timestamp)
func (b *BalasHelper) SetBalasAsInt() int {
	// Return current Unix timestamp as int
	return int(time.Now().Unix())
}

// CheckTimeThrottleCompat checks throttling with sql.NullString balas field
func (s *aiWhatsappService) CheckTimeThrottleWithNullString(balasNullString sql.NullString, thresholdSeconds int) bool {
	if thresholdSeconds <= 0 {
		thresholdSeconds = 4
	}

	if !balasNullString.Valid || balasNullString.String == "" {
		return true // No previous response, allow
	}

	// Parse the timestamp string
	lastTime, err := time.Parse("2006-01-02 15:04:05", balasNullString.String)
	if err != nil {
		return true // Can't parse, allow request
	}

	currentTime := time.Now()
	timeDifference := currentTime.Sub(lastTime).Seconds()

	if timeDifference < float64(thresholdSeconds) {
		logrus.WithFields(logrus.Fields{
			"time_diff": timeDifference,
			"threshold": thresholdSeconds,
		}).Debug("⏱️ THROTTLE: Request throttled")
		return false
	}

	return true
}
