package auth

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

func IsValidRole(role string) bool {
	return role == RoleUser || role == RoleAdmin
}
