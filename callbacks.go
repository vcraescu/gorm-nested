package nested

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

func (p *Plugin) createCallback(scope *gorm.Scope) {
	value := doubleToSingleIndirect(scope.Value)
	if !isTreeNode(value) {
		return
	}

	node := value.(Interface)
	defer refreshNode(node, scope)

	if isRoot(node) {
		updateInsertRootNode(node, scope)

		return
	}

	err := updateTreeAfterInsertChildNode(node, scope)
	if err != nil {
		panic(err)
	}
}

func (p *Plugin) updateCallback(scope *gorm.Scope) {
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
		db.Order("tree_right desc").First(max)

		treeOffset := (max.GetTreeRight() - width) + 1 - node.GetTreeLeft()
		if treeOffset == 0 {
			return
		}

		levelOffset := 0 - node.GetTreeLevel()
		db.
			Table(scope.TableName()).
			Where("tree_left >= ? and tree_right <= ?", node.GetTreeLeft(), node.GetTreeRight()).
			Updates(map[string]interface{}{
				"tree_left":  gorm.Expr("-1 * (tree_left + ?)", treeOffset),
				"tree_right": gorm.Expr("-1 * (tree_right + ?)", treeOffset),
				"tree_level": gorm.Expr("tree_level + ?", levelOffset),
			})

		shiftTreeFromRightOf(scope, node, width)

		db.
			Table(scope.TableName()).
			Where("tree_right < 0", node.GetTreeLeft(), node.GetTreeRight()).
			Updates(map[string]interface{}{
				"tree_left":  gorm.Expr("-1 * tree_left"),
				"tree_right": gorm.Expr("-1 * tree_right"),
			})

		return
	}

	// update current node subtreee and remove it
	treeOffset := parent.GetTreeRight() - node.GetTreeLeft()
	levelOffset := parent.GetTreeLevel() + 1 - node.GetTreeLevel()

	db.
		Table(scope.TableName()).
		Where("tree_left >= ? and tree_right <= ?", node.GetTreeLeft(), node.GetTreeRight()).
		Updates(map[string]interface{}{
			"tree_left":  gorm.Expr("0 - (tree_left + ?)", treeOffset),
			"tree_right": gorm.Expr("0 - (tree_right + ?)", treeOffset),
			"tree_level": gorm.Expr("tree_level + ?", levelOffset),
		})

	// shift nodes from the right of moving node to the left
	shiftTreeFromRightOf(scope, node, width)

	// reload parent because it might be update after the query from above
	db.First(parent)

	// shift nodes from the right of parent node to the right
	shiftTreeFromRightOf(scope, parent, -1*width)

	updateCurrentNode(parent, map[string]interface{}{
		"tree_right": parent.GetTreeRight() + width,
	}, scope)

	// put back current tree
	db.
		Table(scope.TableName()).
		Where("tree_right < 0").
		Updates(map[string]interface{}{
			"tree_left":  gorm.Expr("-1 * tree_left"),
			"tree_right": gorm.Expr("-1 * tree_right"),
		})
}

func (p *Plugin) deleteCallback(scope *gorm.Scope) {
	value := doubleToSingleIndirect(scope.Value)
	if isDeletionIgnored(scope) || !isTreeNode(scope.Value) {
		return
	}

	node := value.(Interface)

	defer refreshNode(node, scope)

	deleteTree(node, scope)
	width := nodeWidth(node)
	shiftTreeFromRightOf(scope, node, width)
}

func shiftTreeFromRightOf(scope *gorm.Scope, node Interface, offset int) {
	db := scope.DB().Set(settingIgnoreUpdate, true)
	db.
		Table(scope.TableName()).
		Where("tree_right > ?", node.GetTreeRight()).
		Update("tree_right", gorm.Expr("tree_right - ?", offset))
	db.
		Table(scope.TableName()).
		Where("tree_left > ?", node.GetTreeRight()).
		Update("tree_left", gorm.Expr("tree_left - ?", offset))
}

func getParent(node Interface, scope *gorm.Scope) (Interface, bool) {
	db := scope.NewDB()
	parent := newNodePtrFromValue(node)
	where := fmt.Sprintf("%s = ?", scope.PrimaryKey())
	db.First(parent, where, node.GetParentID())

	return parent, !isZeroValue(node.GetParentID())
}

func deleteTree(node Interface, scope *gorm.Scope) {
	db := scope.DB().Set(settingIgnoreDelete, true)
	db.Delete(
		newNodePtrFromValue(scope.Value),
		"tree_left > ? AND tree_left < ?",
		node.GetTreeLeft(),
		node.GetTreeRight(),
	)
}

func nodeWidth(node Interface) int {
	return node.GetTreeRight() - node.GetTreeLeft() + 1
}

func isRoot(node Interface) bool {
	return isZeroValue(node.GetParentID())
}

func updateInsertRootNode(node Interface, scope *gorm.Scope) {
	db := scope.DB().Set(settingIgnoreUpdate, true)
	max := newNodePtrFromValue(node)
	db.Order("tree_right desc").First(max)
	updateCurrentNode(node, map[string]interface{}{
		"tree_left":  max.GetTreeRight() + 1,
		"tree_right": max.GetTreeRight() + 2,
	}, scope)
}

func updateTreeAfterInsertChildNode(node Interface, scope *gorm.Scope) error {
	parent, ok := getParent(node, scope)
	if !ok {
		panic(fmt.Errorf("parent not found: %s", node.GetParentID()))
	}

	db := scope.DB().Set(settingIgnoreUpdate, true)
	db.
		Table(scope.TableName()).
		Where("tree_right >= ?", parent.GetTreeRight()).
		Update("tree_right", gorm.Expr("tree_right + 2"))
	db.
		Table(scope.TableName()).
		Where("tree_left >= ?", parent.GetTreeRight()).
		Update("tree_left", gorm.Expr("tree_left + 2"))

	updateCurrentNode(node, map[string]interface{}{
		"tree_left":  parent.GetTreeRight(),
		"tree_right": parent.GetTreeRight() + 1,
		"tree_level": parent.GetTreeLevel() + 1,
	}, scope)

	return nil
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
