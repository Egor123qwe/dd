package permission

import "fmt"

var (
	ErrInvalidPermission = fmt.Errorf("invalid permission")
)

type Permission uint8
type EncodedPermission uint64

func EncodePermissions(perms ...Permission) (EncodedPermission, error) {
	var encoded EncodedPermission

	for _, perm := range perms {
		if !perm.IsValid() {
			return 0, fmt.Errorf("%w: %s", ErrInvalidPermission, perm)
		}

		encoded |= EncodedPermission(1 << perm)
	}

	return encoded, nil
}

func DecodePermissions(encoded EncodedPermission) []Permission {
	var permissions []Permission

	for perm := Permission(0); perm < permEOF; perm++ {
		isFind := (encoded & EncodedPermission(1<<perm)) != 0

		if isFind {
			permissions = append(permissions, perm)
		}
	}

	return permissions
}

func SumPermissions(perms ...EncodedPermission) EncodedPermission {
	var sum EncodedPermission

	for _, perm := range perms {
		sum |= perm
	}

	return sum
}

func (p EncodedPermission) HasPermission(perm Permission) bool {
	if !perm.IsValid() {
		return false
	}

	bitMask := EncodedPermission(1 << perm)

	return (p & bitMask) != 0
}

func (p EncodedPermission) HasAllPermissions(perms ...Permission) bool {
	for _, perm := range perms {
		if !p.HasPermission(perm) {
			return false
		}
	}

	return true
}

func (p EncodedPermission) HasLessOrEqualPermissions(target EncodedPermission) bool {
	return (target & ^p) == 0
}
