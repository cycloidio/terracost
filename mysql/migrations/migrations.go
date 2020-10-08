package migrations

type Migration struct {
	Name string
	SQL  string
}

var Migrations = [1]Migration{
	v0Initial,
}
