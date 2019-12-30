package example

import "time"

// gen:ddd {"table": "units"}
type Unit struct {
	Id       *int       `db:"field=id,primary"`
	UnitType string     `db:"field=type,searchable"`
	Name     string     `db:"field=name,searchGroup=NameAndColor"`
	Color    string     `db:"field=color,searchGroup=NameAndColor"`
	Deleted  *time.Time `db:"field=deleted"`

	Details []*Detail `db:"foreign=detailId"`
}

// gen:ddd {"active":"\"units\".\"deleted\" is null"}
func (a Unit) IsActive() bool {
	return a.Deleted == nil
}

// gen:ddd {"table": "details", "nested": true}
type Detail struct {
	Id          *int   `db:"field=detailId,primary"`
	DetailType  string `db:"field=type,searchable"`
	Description string `db:"field=description"`
}
