package data

// User represents a user in the system
type User struct {
	// Username is the unique identifier for the user
	Username string `json:"username"`

	// Password is the user's password (in a real system, this would be hashed)
	Password string `json:"password"`
}
