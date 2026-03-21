package enums

type CreditScoreLevel int

const (
	LevelDanger CreditScoreLevel = iota // 0
	LevelMidium                         // 1
	LevelGood                           // 2
)

func (g CreditScoreLevel) String() string {
	switch g {
	case LevelDanger:
		return "danger"
	case LevelMidium:
		return "midium"
	default:
		return "good"
	}
}

func (g CreditScoreLevel) ToInt() int {
	return int(g)
}
