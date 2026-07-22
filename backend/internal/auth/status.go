package auth

// Account approval statuses. Self-registered accounts start as pending and
// cannot log in until an admin approves them; admin-created and seeded accounts
// are approved on creation.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
)
