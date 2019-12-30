# Code generator
Generates interface and implementation of the repository for data structure based on described tags

### Generator rules

	Tags:
	primary - mark field as primary key, generate Find method
	searchable - generate Find method for that field
	searchGroup - generate Find method for fields that were marked the same group
	foreign - that field is foreign key in nested struct
	    * nested struct should set "nested" ddd field to true. 
	    * nested struct have no their own repository

	Types:
	*time.Time - generate Mark method that set field value to current timestamp
	
	Methods:
	Method marked such json {"active":"\"deleted\" is null"} (sql condition) is used to determine is entity alive at domain, mock and sql levels.
    Could be used various conditions with several fields. Such as "revoke", "deleted", etc

### Example

```go
package mypkg

// gen:ddd {"table": "units"}
type Unit struct {
	Id           *int       `db:"field=id,primary"`
	Type         string     `db:"field=type,searchable"`
	Name         string     `db:"field=type,searchGroup=NameAndColor"`
	Color        string     `db:"field=email,searchGroup=NameAndColor"`
	Deleted      *time.Time `db:"field=deleted"`
    
	Details []*Detail  `db:"foreign=detailId"`
}

// gen:ddd {"active":"\"units\".\"deleted\" is null"}
func (a Unit) IsActive() bool {	
	return a.Deleted == nil
}

// gen:ddd {"table": "details", "nested": true}
type Detail struct {
	Id            *int       `db:"field=detailId,primary"`
	Type        string       `db:"field=type,searchable"`
	Description string       `db:"field=name"`
}

```
### Usage

```bash
    go build -o generator
    ./generator [source path] [interface path] [implementation path] [mock implementation path]
    goimports -w [path] # format code and manage import section
```
