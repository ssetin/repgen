# Code generator
Generates interface and implementation of the repository for data structure based on described tags

### Generator rules

	Tags:
	primary - mark field as primary key, generate Find method
	searchable - generate Find method for that field
	searchGroup - generate Find method for fields that were marked the same group

	Types:
	*time.Time - generate Mark method that set field value to current timestamp
	
	Methods:
	Method marked such json {"active":"\"deleted\" is null"} is used to determine is entity alive at domain, mock and sql levels.
    Could be used various conditions with several fields. Such as "revoke", "deleted", etc

### Example

```go
// gen:ddd {"table": "accounts"}
type Account struct {
	AccountId    int        `db:"field=id,primary"`
	Phone        string     `db:"field=phone,searchable"`
	Email        string     `db:"field=email,searchGroup=Requisite"`
	PasswordHash string     `db:"field=passwordHash,searchGroup=Requisite"`
	Deleted      *time.Time `db:"field=deleted"`
}

// gen:ddd {"active":"\"deleted\" is null"}
func (a Account) IsActive() bool {
	if a.Deleted == nil {
		return true
	}
	return false
}
```
### Usage

```bash
    go build -o generator
    ./generator [source file] [interface file] [implementation file] [mock implementation file]
    goimports -w [implementation file] # format code and manage import section
```
