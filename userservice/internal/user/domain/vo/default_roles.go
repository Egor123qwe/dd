package vo

type DefaultRoles string

const (
	Farmer  DefaultRoles = "Farmer"
	Manager DefaultRoles = "Manager"
	Service DefaultRoles = "Service"
	SUI     DefaultRoles = "SUI"
)

func (p DefaultRoles) Value() string {
	return string(p)
}
