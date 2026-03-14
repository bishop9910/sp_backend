package enums

type Permission int

const (
	NormalPermission = iota // 0
	AdminPermission         // 1
)

func (p Permission) String() string {
	switch p {
	case NormalPermission:
		return "normal"
	case AdminPermission:
		return "admin"
	default:
		return "normal"
	}
}

func (p Permission) ToInt() int {
	return int(p)
}

func PermissionFromString(s string) Permission {
	switch s {
	case "normal":
		return NormalPermission
	case "admin":
		return AdminPermission
	default:
		return NormalPermission
	}
}
