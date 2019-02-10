package nested

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
	"strings"
)

func (p *Plugin) createCallback(scope *gorm.Scope) {
	p.initColumnNames(scope.Value)

	value := doubleToSingleIndirect(scope.Value)
	if !isTreeNode(value) {
		return
	}

	node := value.(Interface)
	defer refreshNode(node, scope)

	if isRoot(node) {
		p.updateInsertRootNode(node, scope)

		return
	}

	err := p.updateTreeAfterInsertChildNode(node, scope)
	if err != nil {
		panic(err)
	}
}

func (p *Plugin) updateCallback(scope *gorm.Scope) {
	p.initColumnNames(scope.Value)

	value := doubleToSingleIndirect(scope.Value)
	if isUpdateIgnored(scope) || !isTreeNode(value) {
		return
	}

	node := value.(Interface)
	defer refreshNode(node, scope)

	width := nodeWidth(node)
	db := scope.DB().Set(settingIgnoreUpdate, true)

	parent := node.GetParent()
	if isRoot(node) {
		max := newNodePtrFromValue(node)
		db.Order(p.expr(":tree_right desc")).First(max)

		treeOffset := (getTreeRight(max) - width) + 1 - getTreeLeft(node)
		if treeOffset == 0 {
			return
		}

		levelOffset := 0 - getTreeLevel(node)
		db.
			Table(scope.TableName()).
			Where(p.expr(":tree_left >= ? and :tree_right <= ?"), getTreeLeft(node), getTreeRight(node)).
			Updates(map[string]interface{}{
				p.treeLeftName:  gorm.Expr(p.expr("-1 * (:tree_left + ?)"), treeOffset),
				p.treeRightName: gorm.Expr(p.expr("-1 * (:tree_right + ?)"), treeOffset),
				p.treeLevelName: gorm.Expr(p.expr(":tree_level + ?"), levelOffset),
			})

		p.shiftTreeFromRightOf(scope, node, width)

		db.
			Table(scope.TableName()).
			Where(p.expr(":tree_right < 0"), getTreeLeft(node), getTreeRight(node)).
			Updates(map[string]interface{}{
				p.treeLeftName:  gorm.Expr(p.expr("-1 * :tree_left")),
				p.treeRightName: gorm.Expr(p.expr("-1 * tree_right")),
			})

		return
	}

	// update current node subtreee and remove it
	treeOffset := getTreeRight(parent) - getTreeLeft(node)
	levelOffset := getTreeLevel(parent) + 1 - getTreeLevel(node)

	db.
		Table(scope.TableName()).
		Where(p.expr(":tree_left >= ? and :tree_right <= ?"), getTreeLeft(node), getTreeRight(node)).
		Updates(map[string]interface{}{
			p.treeLeftName:  gorm.Expr(p.expr("0 - (:tree_left + ?)"), treeOffset),
			p.treeRightName: gorm.Expr(p.expr("0 - (:tree_right + ?)"), treeOffset),
			p.treeLevelName: gorm.Expr(p.expr(":tree_level + ?"), levelOffset),
		})

	// shift nodes from the right of moving node to the left
	p.shiftTreeFromRightOf(scope, node, width)

	// reload parent because it might be update after the query from above
	db.First(parent)

	// shift nodes from the right of parent node to the right
	p.shiftTreeFromRightOf(scope, parent, -1*width)

	updateCurrentNode(parent, map[string]interface{}{
		p.treeRightName: getTreeRight(parent) + width,
	}, scope)

	// put back current tree
	db.
		Table(scope.TableName()).
		Where(p.expr(":tree_right < 0")).
		Updates(map[string]interface{}{
			p.treeLeftName:  gorm.Expr(p.expr("-1 * :tree_left")),
			p.treeRightName: gorm.Expr(p.expr("-1 * :tree_right")),
		})
}

func (p *Plugin) deleteCallback(scope *gorm.Scope) {
	p.initColumnNames(scope.Value)

	value := doubleToSingleIndirect(scope.Value)
	if isDeletionIgnored(scope) || !isTreeNode(scope.Value) {
		return
	}

	node := value.(Interface)

	defer refreshNode(node, scope)

	p.deleteTree(node, scope)
	width := nodeWidth(node)
	p.shiftTreeFromRightOf(scope, node, width)
}

func (p *Plugin) shiftTreeFromRightOf(scope *gorm.Scope, node Interface, offset int) {
	db := scope.DB().Set(settingIgnoreUpdate, true)
	treeRight := getTreeRight(node)
	db.
		Table(scope.TableName()).
		Where(p.expr(":tree_right > ?"), treeRight).
		Update(p.treeRightName, gorm.Expr(p.expr(":tree_right - ?"), offset))
	db.
		Table(scope.TableName()).
		Where(p.expr(":tree_left > ?"), treeRight).
		Update(p.treeLeftName, gorm.Expr(p.expr(":tree_left - ?"), offset))
}

func findParent(node Interface, scope *gorm.Scope) (Interface, bool) {
	db := scope.NewDB()
	parent := newNodePtrFromValue(node)
	where := fmt.Sprintf("%s = ?", scope.PrimaryKey())
	db.First(parent, where, node.GetParentID())

	return parent, !isZeroValue(node.GetParentID())
}

func (p *Plugin) deleteTree(node Interface, scope *gorm.Scope) {
	db := scope.DB().Set(settingIgnoreDelete, true)
	db.Delete(
		newNodePtrFromValue(scope.Value),
		p.expr(":tree_left > ? AND :tree_left < ?"),
		getTreeLeft(node),
		getTreeRight(node),
	)
}

func nodeWidth(node Interface) int {
	return getTreeRight(node) - getTreeLeft(node) + 1
}

func isRoot(node Interface) bool {
	return isZeroValue(node.GetParentID())
}

func (p *Plugin) updateInsertRootNode(node Interface, scope *gorm.Scope) {
	db := scope.DB().Set(settingIgnoreUpdate, true)
	max := newNodePtrFromValue(node)
	db.Order(p.expr(":tree_right desc")).First(max)
	treeRight := getTreeRight(max)
	updateCurrentNode(node, map[string]interface{}{
		p.treeLeftName:  treeRight + 1,
		p.treeRightName: treeRight + 2,
	}, scope)
}

func (p *Plugin) updateTreeAfterInsertChildNode(node Interface, scope *gorm.Scope) error {
	parent, ok := findParent(node, scope)
	if !ok {
		panic(fmt.Errorf("parent not found: %s", node.GetParentID()))
	}

	db := scope.DB().Set(settingIgnoreUpdate, true)
	treeRight := getTreeRight(parent)
	db.
		Table(scope.TableName()).
		Where(p.expr(":tree_right >= ?"), treeRight).
		Update(p.treeRightName, gorm.Expr(p.expr(":tree_right + 2")))
	db.
		Table(scope.TableName()).
		Where(p.expr(":tree_left >= ?"), treeRight).
		Update(p.treeLeftName, gorm.Expr(p.expr(":tree_left + 2")))

	updateCurrentNode(node, map[string]interface{}{
		p.treeLeftName:  treeRight,
		p.treeRightName: treeRight + 1,
		p.treeLevelName: getTreeLevel(parent) + 1,
	}, scope)

	return nil
}

func (p *Plugin) expr(expr string) string {
	expr = strings.Replace(expr, ":tree_left", p.treeLeftName, -1)
	expr = strings.Replace(expr, ":tree_right", p.treeRightName, -1)
	expr = strings.Replace(expr, ":tree_level", p.treeLevelName, -1)

	return expr
}

func (p *Plugin) initColumnNames(node interface{}) {
	var dbf *gorm.Field

	if p.treeRightName != "" || p.treeLeftName != "" || p.treeLevelName != "" {
		return
	}

	if f, ok := getFieldByTagValue(node, "left"); ok {
		if dbf, ok = p.db.NewScope(node).FieldByName(f.Name); ok {
			p.treeLeftName = dbf.DBName
		}
	}

	if f, ok := getFieldByTagValue(node, "right"); ok {
		if dbf, ok = p.db.NewScope(node).FieldByName(f.Name); ok {
			p.treeRightName = dbf.DBName
		}
	}

	if f, ok := getFieldByTagValue(node, "level"); ok {
		if dbf, ok = p.db.NewScope(node).FieldByName(f.Name); ok {
			p.treeLevelName = dbf.DBName
		}
	}
}

func updateCurrentNode(node Interface, updates map[string]interface{}, scope *gorm.Scope) {
	scope = scope.New(node)
	db := scope.DB().Set(settingIgnoreUpdate, true)
	db.
		Table(scope.TableName()).
		Where(fmt.Sprintf("%s = ?", scope.PrimaryKey()), scope.PrimaryKeyValue()).
		Updates(updates)
}

func refreshNode(node Interface, scope *gorm.Scope) {
	parent := node.GetParent()
	for !isNilInterface(parent) {
		scope.DB().First(parent)
		parent = parent.GetParent()
	}

	scope.DB().First(node)
}

func getFieldByTagValue(node interface{}, tagValue string) (*reflect.StructField, bool) {
	v := reflect.Indirect(reflect.Indirect(reflect.ValueOf(node))).Interface()
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tv, ok := f.Tag.Lookup(tagName)
		if !ok || tv != tagValue {
			continue
		}

		return &f, true
	}

	return nil, false
}

func getTreeLeft(node Interface) int {
	f, ok := getFieldByTagValue(node, "left")
	if !ok {
		return 0
	}

	v := reflect.Indirect(reflect.ValueOf(node))
	return int(v.FieldByName(f.Name).Int())
}

func getTreeRight(node Interface) int {
	f, ok := getFieldByTagValue(node, "right")
	if !ok {
		return 0
	}

	v := reflect.Indirect(reflect.ValueOf(node))
	return int(v.FieldByName(f.Name).Int())
}

func getTreeLevel(node Interface) int {
	f, ok := getFieldByTagValue(node, "level")
	if !ok {
		return 0
	}

	v := reflect.Indirect(reflect.ValueOf(node))
	return int(v.FieldByName(f.Name).Int())
}

func isTreeNode(v interface{}) bool {
	node, ok := v.(Interface)
	if !ok {
		return false
	}

	return isValidNode(node)
}

func isValidNode(node Interface) bool {
	_, ok := getFieldByTagValue(node, "left")
	if !ok {
		return false
	}

	_, ok = getFieldByTagValue(node, "right")
	if !ok {
		return false
	}

	_, ok = getFieldByTagValue(node, "level")

	return ok
}

func isUpdateIgnored(scope *gorm.Scope) bool {
	v, ok := scope.Get(settingIgnoreUpdate)
	if !ok {
		return false
	}

	vv, _ := v.(bool)

	return vv
}

func isDeletionIgnored(scope *gorm.Scope) bool {
	v, ok := scope.Get(settingIgnoreDelete)
	if !ok {
		return false
	}

	vv, _ := v.(bool)

	return vv
}
