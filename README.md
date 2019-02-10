# GORM Nested Set Tree [![Go Report Card](https://goreportcard.com/badge/github.com/vcraescu/gorm-nested)](https://goreportcard.com/report/github.com/vcraescu/gorm-nested) [![Build Status](https://travis-ci.com/vcraescu/gorm-nested.svg?branch=master)](https://travis-ci.com/vcraescu/gorm-nested) [![Coverage Status](https://coveralls.io/repos/github/vcraescu/gorm-nested/badge.svg?branch=master)](https://coveralls.io/github/vcraescu/gorm-nested?branch=master)

#### Do not use in production. This is an experiment!!!


#### Install

`go get github.com/vcraescu/gorm-nested`


#### How to use

```go
type Taxon struct {
	ID        uint `gorm:"primary_key"`
	Name      string
	ParentID  uint
	Parent    *Taxon `gorm:"association_autoupdate:false"`
	TreeLeft  int    `gorm-nested:"left"`
	TreeRight int    `gorm-nested:"right"`
	TreeLevel int    `gorm-nested:"level"`
}

func (t Taxon) GetParentID() interface{} {
	return t.ParentID
}

func (t Taxon) GetParent() nested.Interface {
	return t.Parent
}

func main() {
	db, err := gorm.Open("sqlite3", "my_dbname")
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Taxon{})

	_, err = nested.Register(db)
}
```
