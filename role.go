package babynames

import "fmt"

// Role identifies who the identity accessing the system is.
type Role int

const (
	// MomRole is the mom
	MomRole Role = iota

	// DadRole is the dad
	DadRole
)

// InverseRole gets the inverse of the specified role (mom turns in to dad, and vice versa)
func InverseRole(role Role) Role {
	switch role {
	case MomRole:
		return DadRole
	case DadRole:
		return MomRole
	default:
		panic(fmt.Errorf("Unable to inverse role '%+v'", role))
	}
}

// RoleName returns the display name of the role.
func RoleName(role Role) string {
	switch role {
	case MomRole:
		return "Mom"
	case DadRole:
		return "Dad"
	default:
		panic(fmt.Errorf("Unable to get name for role '%+v'", role))
	}
}
