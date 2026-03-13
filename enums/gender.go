package enums

type Gender int

const (
	GenderUnknown Gender = iota // 0
	GenderMale                  // 1
	GenderFemale                // 2
)

func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	default:
		return "unknown"
	}
}

func (g Gender) ToInt() int {
	return int(g)
}

func FromString(s string) Gender {
	switch s {
	case "male":
		return GenderMale
	case "female":
		return GenderFemale
	default:
		return GenderUnknown
	}
}
