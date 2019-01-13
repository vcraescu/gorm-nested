package nested_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vcraescu/gorm-nested"
	"math/rand"
	"os"
	"testing"
)

var dbName = fmt.Sprintf("test_%d.db", rand.Int())

type PluginTestSuite struct {
	suite.Suite
	db *gorm.DB
}

type Taxon struct {
	nested.TreeNode

	ID       uint `gorm:"primary_key"`
	Name     string
	ParentID uint
	Parent   *Taxon `gorm:"association_autoupdate:false"`
}

func (t Taxon) GetParentID() interface{} {
	return t.ParentID
}

func (t Taxon) GetParent() nested.Interface {
	return t.Parent
}

func (suite *PluginTestSuite) SetupTest() {
	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		panic(fmt.Errorf("setup test: %s", err))
	}

	suite.db = db
	suite.db.AutoMigrate(&Taxon{})

	_, err = nested.Register(suite.db)
	if err != nil {
		panic(err)
	}
}

func (suite *PluginTestSuite) TearDownTest() {
	if err := suite.db.Close(); err != nil {
		panic(fmt.Errorf("tear down test: %s", err))
	}

	if err := os.Remove(dbName); err != nil {
		panic(fmt.Errorf("tear down test: %s", err))
	}
}

func (suite *PluginTestSuite) TestAddRoot() {
	root1 := Taxon{
		Name: "Root1",
	}
	suite.db.Create(&root1)

	root2 := Taxon{
		Name: "Root2",
	}
	suite.db.Create(&root2)

	root3 := Taxon{
		Name: "Root3",
	}
	suite.db.Create(&root3)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 3)
	assert.Equal(suite.T(), 1, taxons[0].TreeLeft)
	assert.Equal(suite.T(), 2, taxons[0].TreeRight)

	assert.Equal(suite.T(), 3, taxons[1].TreeLeft)
	assert.Equal(suite.T(), 4, taxons[1].TreeRight)

	assert.Equal(suite.T(), 5, taxons[2].TreeLeft)
	assert.Equal(suite.T(), 6, taxons[2].TreeRight)
}

func (suite *PluginTestSuite) TestInsertEntireTree() {
	node := Taxon{
		Name: "Tube",
		Parent: &Taxon{
			Name: "Television",
			Parent: &Taxon{
				Name: "Electronics",
			},
		},
	}
	suite.db.Create(&node)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 3)
	assert.Equal(suite.T(), 1, taxons[0].TreeLeft)
	assert.Equal(suite.T(), 6, taxons[0].TreeRight)

	assert.Equal(suite.T(), 2, taxons[1].TreeLeft)
	assert.Equal(suite.T(), 5, taxons[1].TreeRight)

	assert.Equal(suite.T(), 3, taxons[2].TreeLeft)
	assert.Equal(suite.T(), 4, taxons[2].TreeRight)
}

func (suite *PluginTestSuite) TestInsertNodeByNode() {
	television := Taxon{
		Name: "Television",
		Parent: &Taxon{
			Name: "Electronics",
		},
	}
	suite.db.Create(&television)

	tube := Taxon{
		Name:   "Tube",
		Parent: &television,
	}
	suite.db.Create(&tube)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 3)
	assert.Equal(suite.T(), 1, taxons[0].TreeLeft)
	assert.Equal(suite.T(), 6, taxons[0].TreeRight)
	assert.Equal(suite.T(), 0, taxons[0].TreeLevel)

	assert.Equal(suite.T(), 2, taxons[1].TreeLeft)
	assert.Equal(suite.T(), 5, taxons[1].TreeRight)
	assert.Equal(suite.T(), 1, taxons[1].TreeLevel)

	assert.Equal(suite.T(), 3, taxons[2].TreeLeft)
	assert.Equal(suite.T(), 4, taxons[2].TreeRight)
	assert.Equal(suite.T(), 2, taxons[2].TreeLevel)
}

func (suite *PluginTestSuite) TestDeleteNode() {
	suite.createTree()

	television := Taxon{}
	suite.db.First(&television, "name = 'Television'")
	suite.db.Delete(&television)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 7)
	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.TreeLeft)
			assert.Equal(suite.T(), 14, taxon.TreeRight)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 2, taxon.TreeLeft)
			assert.Equal(suite.T(), 3, taxon.TreeRight)
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 4, taxon.TreeLeft)
			assert.Equal(suite.T(), 13, taxon.TreeRight)
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 5, taxon.TreeLeft)
			assert.Equal(suite.T(), 8, taxon.TreeRight)
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 6, taxon.TreeLeft)
			assert.Equal(suite.T(), 7, taxon.TreeRight)
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 9, taxon.TreeLeft)
			assert.Equal(suite.T(), 10, taxon.TreeRight)
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 11, taxon.TreeLeft)
			assert.Equal(suite.T(), 12, taxon.TreeRight)
			count++
			break
		}
	}

	var portableElectronics Taxon
	suite.db.First(&portableElectronics, "name = 'Portable Electronics'")
	suite.db.Delete(&portableElectronics)

	taxons = []Taxon{}
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 2)
	count = 0
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.TreeLeft)
			assert.Equal(suite.T(), 4, taxon.TreeRight)
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 2, taxon.TreeLeft)
			assert.Equal(suite.T(), 3, taxon.TreeRight)
			count++
			break
		}
	}

	var gameConsoles Taxon
	suite.db.First(&gameConsoles, "name = 'Game Consoles'")
	suite.db.Delete(&gameConsoles)
	taxons = []Taxon{}
	suite.db.Find(&taxons)

	assert.Equal(suite.T(), 1, taxons[0].TreeLeft)
	assert.Equal(suite.T(), 2, taxons[0].TreeRight)
}

func (suite *PluginTestSuite) TestMoveNodeToLeft() {
	suite.createTree()

	var portableElectronics Taxon
	var lcd Taxon

	assert.False(suite.T(), suite.db.First(&portableElectronics, "name = 'Portable Electronics'").RecordNotFound())
	assert.False(suite.T(), suite.db.First(&lcd, "name = 'LCD'").RecordNotFound())

	portableElectronics.Parent = &lcd
	portableElectronics.ParentID = lcd.ID

	suite.db.Save(&portableElectronics)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)

	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.TreeLeft)
			assert.Equal(suite.T(), 22, taxon.TreeRight)
			assert.Equal(suite.T(), 0, taxon.GetTreeLevel())
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.TreeLeft)
			assert.Equal(suite.T(), 19, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 20, taxon.TreeLeft)
			assert.Equal(suite.T(), 21, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.TreeLeft)
			assert.Equal(suite.T(), 4, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "LCD":
			assert.Equal(suite.T(), 5, taxon.TreeLeft)
			assert.Equal(suite.T(), 16, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 17, taxon.TreeLeft)
			assert.Equal(suite.T(), 18, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 6, taxon.TreeLeft)
			assert.Equal(suite.T(), 15, taxon.TreeRight)
			assert.Equal(suite.T(), 3, taxon.GetTreeLevel())
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 7, taxon.TreeLeft)
			assert.Equal(suite.T(), 10, taxon.TreeRight)
			assert.Equal(suite.T(), 4, taxon.GetTreeLevel())
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 8, taxon.TreeLeft)
			assert.Equal(suite.T(), 9, taxon.TreeRight)
			assert.Equal(suite.T(), 5, taxon.GetTreeLevel())
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 11, taxon.TreeLeft)
			assert.Equal(suite.T(), 12, taxon.TreeRight)
			assert.Equal(suite.T(), 4, taxon.GetTreeLevel())
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 13, taxon.TreeLeft)
			assert.Equal(suite.T(), 14, taxon.TreeRight)
			assert.Equal(suite.T(), 4, taxon.GetTreeLevel())
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestMoveNodeToRight() {
	suite.createTree()

	var mp3 Taxon
	var lcd Taxon

	assert.False(suite.T(), suite.db.First(&mp3, "name = 'MP3'").RecordNotFound())
	assert.False(suite.T(), suite.db.First(&lcd, "name = 'LCD'").RecordNotFound())

	lcd.Parent = &mp3
	lcd.ParentID = mp3.ID

	suite.db.Save(&lcd)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)

	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.TreeLeft)
			assert.Equal(suite.T(), 22, taxon.TreeRight)
			assert.Equal(suite.T(), 0, taxon.GetTreeLevel())
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.TreeLeft)
			assert.Equal(suite.T(), 7, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 8, taxon.TreeLeft)
			assert.Equal(suite.T(), 9, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.TreeLeft)
			assert.Equal(suite.T(), 4, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 5, taxon.TreeLeft)
			assert.Equal(suite.T(), 6, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 10, taxon.TreeLeft)
			assert.Equal(suite.T(), 21, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 11, taxon.TreeLeft)
			assert.Equal(suite.T(), 16, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 12, taxon.TreeLeft)
			assert.Equal(suite.T(), 13, taxon.TreeRight)
			assert.Equal(suite.T(), 3, taxon.GetTreeLevel())
			count++
			break
		case "LCD":
			assert.Equal(suite.T(), 16, taxon.TreeLeft)
			assert.Equal(suite.T(), 17, taxon.TreeRight)
			assert.Equal(suite.T(), 3, taxon.GetTreeLevel())
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 17, taxon.TreeLeft)
			assert.Equal(suite.T(), 18, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 19, taxon.TreeLeft)
			assert.Equal(suite.T(), 20, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestChildNodeBecomesRoot() {
	suite.createTree()

	var mp3 Taxon

	assert.False(suite.T(), suite.db.First(&mp3, "name = 'MP3'").RecordNotFound())

	mp3.Parent = nil
	mp3.ParentID = 0

	suite.db.Save(&mp3)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)

	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.TreeLeft)
			assert.Equal(suite.T(), 18, taxon.TreeRight)
			assert.Equal(suite.T(), 0, taxon.GetTreeLevel())
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.TreeLeft)
			assert.Equal(suite.T(), 9, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 10, taxon.TreeLeft)
			assert.Equal(suite.T(), 11, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.TreeLeft)
			assert.Equal(suite.T(), 4, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "LCD":
			assert.Equal(suite.T(), 5, taxon.TreeLeft)
			assert.Equal(suite.T(), 6, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 7, taxon.TreeLeft)
			assert.Equal(suite.T(), 8, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 12, taxon.TreeLeft)
			assert.Equal(suite.T(), 17, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 13, taxon.TreeLeft)
			assert.Equal(suite.T(), 14, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 15, taxon.TreeLeft)
			assert.Equal(suite.T(), 16, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 19, taxon.TreeLeft)
			assert.Equal(suite.T(), 22, taxon.TreeRight)
			assert.Equal(suite.T(), 0, taxon.GetTreeLevel())
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 20, taxon.TreeLeft)
			assert.Equal(suite.T(), 21, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func (suite *PluginTestSuite) TestAutoUpdateParentAssociation() {
	electronics := Taxon{
		Name: "Electronics",
	}

	television := Taxon{
		Name:   "Television",
		Parent: &electronics,
	}
	gameConsoles := Taxon{
		Name:   "Game Consoles",
		Parent: &electronics,
	}
	portableElectronics := Taxon{
		Name:   "Portable Electronics",
		Parent: &electronics,
	}

	tube := Taxon{
		Name:   "Tube",
		Parent: &television,
	}
	lcd := Taxon{
		Name:   "LCD",
		Parent: &television,
	}
	plasma := Taxon{
		Name:   "Plasma",
		Parent: &television,
	}

	mp3 := Taxon{
		Name:   "MP3",
		Parent: &portableElectronics,
	}

	cdPlayer := Taxon{
		Name:   "CD Player",
		Parent: &portableElectronics,
	}

	radio := Taxon{
		Name:   "Radio",
		Parent: &portableElectronics,
	}

	flash := Taxon{
		Name:   "Flash",
		Parent: &mp3,
	}

	suite.db.Save(&television)
	suite.db.Save(&gameConsoles)
	suite.db.Save(&tube)
	suite.db.Save(&lcd)
	suite.db.Save(&plasma)
	suite.db.Save(&flash)
	suite.db.Save(&cdPlayer)
	suite.db.Save(&radio)

	assert.Equal(suite.T(), 1, electronics.TreeLeft)
	assert.Equal(suite.T(), 22, electronics.TreeRight)
	assert.Equal(suite.T(), 0, electronics.GetTreeLevel())

	assert.Equal(suite.T(), 2, television.TreeLeft)
	assert.Equal(suite.T(), 9, television.TreeRight)
	assert.Equal(suite.T(), 1, television.GetTreeLevel())

	assert.Equal(suite.T(), 3, tube.TreeLeft)
	assert.Equal(suite.T(), 4, tube.TreeRight)
	assert.Equal(suite.T(), 2, tube.GetTreeLevel())

	assert.Equal(suite.T(), 5, lcd.TreeLeft)
	assert.Equal(suite.T(), 6, lcd.TreeRight)
	assert.Equal(suite.T(), 2, lcd.GetTreeLevel())

	assert.Equal(suite.T(), 7, plasma.TreeLeft)
	assert.Equal(suite.T(), 8, plasma.TreeRight)
	assert.Equal(suite.T(), 2, plasma.GetTreeLevel())

	//assert.Equal(suite.T(), 10, gameConsoles.TreeLeft)
	//assert.Equal(suite.T(), 11, gameConsoles.TreeRight)
	//assert.Equal(suite.T(), 1, gameConsoles.GetTreeLevel())

	assert.Equal(suite.T(), 12, portableElectronics.TreeLeft)
	assert.Equal(suite.T(), 21, portableElectronics.TreeRight)
	assert.Equal(suite.T(), 1, portableElectronics.GetTreeLevel())

	assert.Equal(suite.T(), 13, mp3.TreeLeft)
	assert.Equal(suite.T(), 16, mp3.TreeRight)
	assert.Equal(suite.T(), 2, mp3.GetTreeLevel())

	assert.Equal(suite.T(), 17, cdPlayer.TreeLeft)
	assert.Equal(suite.T(), 18, cdPlayer.TreeRight)
	assert.Equal(suite.T(), 2, cdPlayer.GetTreeLevel())

	assert.Equal(suite.T(), 19, radio.TreeLeft)
	assert.Equal(suite.T(), 20, radio.TreeRight)
	assert.Equal(suite.T(), 2, radio.GetTreeLevel())

	assert.Equal(suite.T(), 14, flash.TreeLeft)
	assert.Equal(suite.T(), 15, flash.TreeRight)
	assert.Equal(suite.T(), 3, flash.GetTreeLevel())
}

func (suite *PluginTestSuite) createTree() {
	electronics := Taxon{
		Name: "Electronics",
	}

	television := Taxon{
		Name:   "Television",
		Parent: &electronics,
	}
	gameConsoles := Taxon{
		Name:   "Game Consoles",
		Parent: &electronics,
	}
	portableElectronics := Taxon{
		Name:   "Portable Electronics",
		Parent: &electronics,
	}

	tube := Taxon{
		Name:   "Tube",
		Parent: &television,
	}
	lcd := Taxon{
		Name:   "LCD",
		Parent: &television,
	}
	plasma := Taxon{
		Name:   "Plasma",
		Parent: &television,
	}

	mp3 := Taxon{
		Name:   "MP3",
		Parent: &portableElectronics,
	}

	cdPlayer := Taxon{
		Name:   "CD Player",
		Parent: &portableElectronics,
	}

	radio := Taxon{
		Name:   "Radio",
		Parent: &portableElectronics,
	}

	flash := Taxon{
		Name:   "Flash",
		Parent: &mp3,
	}

	suite.db.Save(&television)
	suite.db.Save(&gameConsoles)
	suite.db.Save(&tube)
	suite.db.Save(&lcd)
	suite.db.Save(&plasma)
	suite.db.Save(&flash)
	suite.db.Save(&cdPlayer)
	suite.db.Save(&radio)

	var taxons []Taxon
	suite.db.Find(&taxons)

	assert.Len(suite.T(), taxons, 11)
	var count int
	for _, taxon := range taxons {
		switch taxon.Name {
		case "Electronics":
			assert.Equal(suite.T(), 1, taxon.TreeLeft)
			assert.Equal(suite.T(), 22, taxon.TreeRight)
			assert.Equal(suite.T(), 0, taxon.GetTreeLevel())
			count++
			break
		case "Television":
			assert.Equal(suite.T(), 2, taxon.TreeLeft)
			assert.Equal(suite.T(), 9, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Tube":
			assert.Equal(suite.T(), 3, taxon.TreeLeft)
			assert.Equal(suite.T(), 4, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
		case "LCD":
			assert.Equal(suite.T(), 5, taxon.TreeLeft)
			assert.Equal(suite.T(), 6, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Plasma":
			assert.Equal(suite.T(), 7, taxon.TreeLeft)
			assert.Equal(suite.T(), 8, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Game Consoles":
			assert.Equal(suite.T(), 10, taxon.TreeLeft)
			assert.Equal(suite.T(), 11, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "Portable Electronics":
			assert.Equal(suite.T(), 12, taxon.TreeLeft)
			assert.Equal(suite.T(), 21, taxon.TreeRight)
			assert.Equal(suite.T(), 1, taxon.GetTreeLevel())
			count++
			break
		case "MP3":
			assert.Equal(suite.T(), 13, taxon.TreeLeft)
			assert.Equal(suite.T(), 16, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "CD Player":
			assert.Equal(suite.T(), 17, taxon.TreeLeft)
			assert.Equal(suite.T(), 18, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Radio":
			assert.Equal(suite.T(), 19, taxon.TreeLeft)
			assert.Equal(suite.T(), 20, taxon.TreeRight)
			assert.Equal(suite.T(), 2, taxon.GetTreeLevel())
			count++
			break
		case "Flash":
			assert.Equal(suite.T(), 14, taxon.TreeLeft)
			assert.Equal(suite.T(), 15, taxon.TreeRight)
			assert.Equal(suite.T(), 3, taxon.GetTreeLevel())
			count++
			break
		}
	}

	assert.Equal(suite.T(), len(taxons), count)
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}
