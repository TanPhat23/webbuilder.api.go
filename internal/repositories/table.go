package repositories

type TableName string

// Table names as for the database
const (
	TableProject TableName = `public."Project"`

	TableElement TableName = `public."Element"`

	TableSetting TableName = `public."Setting"`

	TablePage TableName = `public."Page"`
)

//Convert TableName to string for gorm
func (t TableName) String() string {
	return string(t)
}
