package migrations

// Migration represents a single DB migration with a unique Name and an SQL snippet to execute.
type Migration struct {
	Name string
	SQL  string
}

// Migrations is an ordered list of migrations to track and execute. It is represented by a fixed-size array
// to break the build if conflicting migrations were added concurrently.
var Migrations = [3]Migration{
	v0Initial,
	v1NameIndexes,
	v2ExtendPriceUnit,
}
