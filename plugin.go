package nested

import (
	"github.com/jinzhu/gorm"
)

const (
	callbackNameCreate  = "gorm-nested:create"
	callbackNameUpdate  = "gorm-nested:update"
	callbackNameDelete  = "gorm-nested:delete"
	settingIgnoreUpdate = "gorm-nested:ignore_update"
	settingIgnoreDelete = "gorm-nested:ignore_delete"
)

// Plugin gorm nested set plugin
type Plugin struct {
	db *gorm.DB
}

// Register registers nested set plugin
func Register(db *gorm.DB) (Plugin, error) {
	p := Plugin{db: db}

	p.enableCallbacks()

	return p, nil
}

func (p *Plugin) enableCallbacks() {
	callback := p.db.Callback()
	callback.Create().After("gorm:after_create").Register(callbackNameCreate, p.createCallback)
	callback.Update().After("gorm:after_update").Register(callbackNameUpdate, p.updateCallback)
	callback.Delete().After("gorm:after_delete").Register(callbackNameDelete, p.deleteCallback)
}

// Interface must be implemented by the gorm model
type Interface interface {
	GetTreeLeft() int
	GetTreeRight() int
	GetTreeLevel() int
	GetParentID() interface{}
	GetParent() Interface
	SetTreeLeft(treeLeft int)
	SetTreeRight(treeRight int)
	SetTreeLevel(level int)
}

// TreeNode mode definition, include fields TreeLeft, TreeRight, TreeLevel which could be embedded in your models
type TreeNode struct {
	TreeLeft  int
	TreeRight int
	TreeLevel int
}

// GetTreeLeft returns tree node left value
func (tn TreeNode) GetTreeLeft() int {
	return tn.TreeLeft
}

// GetTreeRight returns tree node right value
func (tn TreeNode) GetTreeRight() int {
	return tn.TreeRight
}

// GetTreeLevel returns tree node level
func (tn TreeNode) GetTreeLevel() int {
	return tn.TreeLevel
}

// SetTreeLeft sets tree node left value
func (tn *TreeNode) SetTreeLeft(treeLeft int) {
	tn.TreeLeft = treeLeft
}

// SetTreeRight sets tree node right value
func (tn *TreeNode) SetTreeRight(treeRight int) {
	tn.TreeRight = treeRight
}

// SetTreeLevel sets tree node level
func (tn *TreeNode) SetTreeLevel(level int) {
	tn.TreeLevel = level
}
