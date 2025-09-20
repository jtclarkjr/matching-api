package models

import (
	"time"
)

// User represents a user in the dating app
type User struct {
	ID          string     `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"-" db:"password"` // Never return password in JSON
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Age         int        `json:"age" db:"age"`
	Bio         string     `json:"bio" db:"bio"`
	Gender      string     `json:"gender" db:"gender"`
	Location    *Location  `json:"location" db:"location"`
	Photos      []Photo    `json:"photos,omitempty"`
	Preferences *UserPrefs `json:"preferences,omitempty"`
	IsVerified  bool       `json:"is_verified" db:"is_verified"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastSeen    *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Location represents geographical location
type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	City      string  `json:"city,omitempty" db:"city"`
	State     string  `json:"state,omitempty" db:"state"`
	Country   string  `json:"country,omitempty" db:"country"`
}

// Photo represents user photos
type Photo struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	URL       string    `json:"url" db:"url"`
	Position  int       `json:"position" db:"position"` // Order of photos (1 = primary)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserPrefs represents user matching preferences
type UserPrefs struct {
	ID              string   `json:"id" db:"id"`
	UserID          string   `json:"user_id" db:"user_id"`
	AgeMin          int      `json:"age_min" db:"age_min"`
	AgeMax          int      `json:"age_max" db:"age_max"`
	MaxDistance     int      `json:"max_distance" db:"max_distance"` // in kilometers
	InterestedIn    []string `json:"interested_in" db:"interested_in"` // genders
	ShowMe          string   `json:"show_me" db:"show_me"`
	OnlyVerified    bool     `json:"only_verified" db:"only_verified"`
	HideDistance    bool     `json:"hide_distance" db:"hide_distance"`
	HideAge         bool     `json:"hide_age" db:"hide_age"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Public returns a user without sensitive information
func (u *User) Public() *User {
	return &User{
		ID:         u.ID,
		FirstName:  u.FirstName,
		Age:        u.Age,
		Bio:        u.Bio,
		Gender:     u.Gender,
		Location:   u.Location,
		Photos:     u.Photos,
		IsVerified: u.IsVerified,
		IsActive:   u.IsActive,
		LastSeen:   u.LastSeen,
	}
}