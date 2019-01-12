package tree

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Plugin struct {
	db   *gorm.DB
}

func Register(db *gorm.DB) (Plugin, error) {
	p := Plugin{db: db}
	callback := db.Callback()
	callback.Create().After("gorm:after_create").Register("loggable:create", p.addCreated)
	callback.Update().After("gorm:after_update").Register("loggable:update", p.addUpdated)
	callback.Delete().After("gorm:after_delete").Register("loggable:delete", p.addDeleted)

	return p, nil
}

type NestedSet struct {
	ParentID  uint
	TreeLeft  int
	TreeRight int
	TreeLevel int
}

func (t *NestedSet) BeforeSave(tx *gorm.DB) error {
	parent := NestedSet{}
	scope := tx.NewScope(t)
	fmt.Println(scope.PrimaryKey())
	where := fmt.Sprintf("%s = ?", scope.PrimaryKey())
	tx.First(&parent, where, t.ParentID)
	if t.ParentID == 0 {
		max := NestedSet{}
		tx.Order("tree_right desc").First(&max)
		t.TreeLeft = max.TreeRight + 1
		t.TreeRight = max.TreeRight + 2

		return nil
	}

	t.TreeLeft = parent.TreeRight
	t.TreeRight = parent.TreeRight + 1
	t.TreeLevel = parent.TreeLevel + 1

	tx.
		Exec("UPDATE taxons SET tree_right = tree_right + 2 WHERE tree_right >= ?", parent.TreeRight).
		Exec("UPDATE taxons SET tree_left = tree_left + 2 WHERE tree_left >= ?", parent.TreeRight)

	return nil
}

func (t *NestedSet) BeforeDelete(tx *gorm.DB) error {
	tt := NestedSet{}
	scope := tx.NewScope(t)
	where := fmt.Sprintf("%s = ?", scope.PrimaryKey())
	if tx.First(&tt, where, scope.PrimaryKeyValue()).RecordNotFound() {
		return nil
	}

	width := tt.TreeRight - tt.TreeLeft + 1
	tx.
		Exec("DELETE FROM taxons WHERE tree_left > ? AND tree_left < ?", tt.TreeLeft, tt.TreeRight).
		Exec("UPDATE taxons SET tree_right = tree_right - ? WHERE tree_right > ?", width, tt.TreeRight).
		Exec("UPDATE taxons SET tree_left = tree_left - ? WHERE tree_left > ?", width, tt.TreeRight)

	return nil
}
