package repositories

type TableName string

// Table names as for the database
const (
	TableProject TableName = `public."Project"`

	TableElement TableName = `public."Element"`

	TableSetting TableName = `public."Setting"`

	TablePage TableName = `public."Page"`

	TableGroup TableName = `public."Group"`

	TableSnapshot TableName = `public."Snapshot"`
)

//Convert TableName to string for gorm
func (t TableName) String() string {
	return string(t)
}
