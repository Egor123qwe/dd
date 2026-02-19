package permission

const (
	SystemActorID = -1
)

type Permit struct {
	actorId int
	perms   []Permission
}

func NewPermit(actorId int, perms ...Permission) Permit {
	p := Permit{
		actorId: actorId,
		perms:   perms,
	}

	return p
}

func (p Permit) GetActorId() int {
	return p.actorId
}

func (p Permit) GetPermissions() []Permission {
	return p.perms
}

func (p Permit) GetExternalPermissions() []Permission {
	allPerms := p.GetPermissions()
	resultPerms := make([]Permission, 0, len(allPerms))

	for _, perm := range allPerms {
		if perm.IsExternal() {
			resultPerms = append(resultPerms, perm)
		}
	}

	return resultPerms
}

func (p Permit) HasPermission(perm Permission) bool {
	for _, p := range p.perms {
		if p == perm {
			return true
		}
	}

	return false
}

func (p Permit) HasAllPermissions(perms ...Permission) bool {
	for _, perm := range perms {
		if !p.HasPermission(perm) {
			return false
		}
	}

	return true
}
