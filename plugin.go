package nested

import (
	"github.com/jinzhu/gorm"
)

const (
	tagName             = "gorm-nested"
	callbackNameCreate  = "gorm-nested:create"
	callbackNameUpdate  = "gorm-nested:update"
	callbackNameDelete  = "gorm-nested:delete"
	settingIgnoreUpdate = "gorm-nested:ignore_update"
	settingIgnoreDelete = "gorm-nested:ignore_delete"
)

// Plugin gorm nested set plugin
type Plugin struct {
	db            *gorm.DB
	treeLeftName  string
	treeRightName string
	treeLevelName string
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
	GetParentID() interface{}
	GetParent() Interface
}
