package types

type Bet struct {
	Id       int
	Duration int
	Name     string
	Pot      int64
	Active   bool
	Result   bool
}

type Rank struct {
	User   string
	Points int64
}

type DBProperties struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}
