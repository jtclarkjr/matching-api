package match

import (
	"math"
	"sort"
	"time"

	"matching-api/internal/models"
)

// findPotentialMatches finds potential matches based on user preferences and compatibility
func (h *Handler) findPotentialMatches(user *models.User, prefs *models.UserPrefs, limit int) []models.User {
	// Query the database with complex filters based on user preferences
	// TODO: Implement sophisticated database query with preference filtering
	// This should include age range, distance, gender preferences, etc.
	potentialUsers := h.getPotentialUsers(user, prefs)

	var matches []models.User

	for _, candidate := range potentialUsers {
		// Skip if already swiped
		if h.hasAlreadySwiped(user.ID, candidate.ID) {
			continue
		}

		// Apply preference filters
		if !h.matchesPreferences(candidate, *prefs, *user) {
			continue
		}

		matches = append(matches, candidate)
	}

	// Sort by compatibility score
	sort.Slice(matches, func(i, j int) bool {
		scoreI := h.calculateCompatibilityScore(*user, matches[i])
		scoreJ := h.calculateCompatibilityScore(*user, matches[j])
		return scoreI > scoreJ
	})

	// Limit results
	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
}

// matchesPreferences checks if a candidate matches user preferences
func (h *Handler) matchesPreferences(candidate models.User, prefs models.UserPrefs, user models.User) bool {
	// Age filter
	if candidate.Age < prefs.AgeMin || candidate.Age > prefs.AgeMax {
		return false
	}

	// Gender filter
	if len(prefs.InterestedIn) > 0 {
		genderMatches := false
		for _, interestedGender := range prefs.InterestedIn {
			if candidate.Gender == interestedGender {
				genderMatches = true
				break
			}
		}
		if !genderMatches {
			return false
		}
	}

	// Distance filter
	if user.Location != nil && candidate.Location != nil {
		distance := h.calculateDistance(*user.Location, *candidate.Location)
		if distance > float64(prefs.MaxDistance) {
			return false
		}
	}

	// Verification filter
	if prefs.OnlyVerified {
		// Check if user has verified profile (phone, email, photo verification)
		// TODO: Implement verification check against user verification status
		// SELECT is_verified FROM users WHERE id = ?
		if !candidate.IsVerified {
			return false
		}
	}

	return true
}

// calculateCompatibilityScore calculates compatibility score between two users
func (h *Handler) calculateCompatibilityScore(user1, user2 models.User) float64 {
	score := 0.0

	// Base score
	score = 50.0

	// Age similarity (closer ages get higher scores)
	ageDiff := math.Abs(float64(user1.Age - user2.Age))
	if ageDiff <= 2 {
		score += 20
	} else if ageDiff <= 5 {
		score += 10
	} else if ageDiff <= 10 {
		score += 5
	}

	// Location proximity (if both have locations)
	if user1.Location != nil && user2.Location != nil {
		distance := h.calculateDistance(*user1.Location, *user2.Location)
		if distance <= 5 {
			score += 15
		} else if distance <= 15 {
			score += 10
		} else if distance <= 25 {
			score += 5
		}
	}

	// Activity bonus (recently active users)
	if user2.LastSeen != nil && time.Since(*user2.LastSeen) < 24*time.Hour {
		score += 10
	}

	// Photo bonus (users with more photos)
	if len(user2.Photos) >= 3 {
		score += 5
	}

	// Bio bonus (users with bios)
	if len(user2.Bio) > 20 {
		score += 5
	}

	return score
}

// calculateDistance calculates distance between two locations using Haversine formula
func (h *Handler) calculateDistance(loc1, loc2 models.Location) float64 {
	// Haversine formula for distance calculation
	const R = 6371 // Earth's radius in kilometers

	lat1Rad := loc1.Latitude * math.Pi / 180
	lat2Rad := loc2.Latitude * math.Pi / 180
	deltaLat := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
	deltaLng := (loc2.Longitude - loc1.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}