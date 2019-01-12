package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/vcraescu/gorm-tree/tree"
)

type Taxon struct {
	tree.NestedSet
	ID   uint `gorm:"primary_key"`
	Name string
}

func main() {
	db := getDB()
	defer db.Close()

	db.AutoMigrate(&Taxon{})

	createRecords(db)
}

func createRecords(db *gorm.DB) {
	root := Taxon{
		Name: "Electronics",
	}
	db.Create(&root)

	//rootA := Taxon{
	//	Name: "A",
	//}
	//db.Create(&rootA)

	// level 1
	television := Taxon{
		Name:     "Television",
		NestedSet: tree.NestedSet{
			ParentID: root.ID,
		},
		//Parent:   &root,
	}
	db.Create(&television)

	gameConsoles := Taxon{
		Name:     "Game Consoles",
		NestedSet: tree.NestedSet{
			ParentID: root.ID,
			//Parent:   &root,
		},
	}
	db.Create(&gameConsoles)

	//portElectronics := Taxon{
	//	Name:     "Portable Electronics",
	//	ParentID: root.ID,
	//	Parent:   &root,
	//}
	//db.Create(&portElectronics)
	//
	//// level 2
	//tube := Taxon{
	//	Name:     "Tube",
	//	ParentID: television.ID,
	//	Parent:   &television,
	//}
	//db.Create(&tube)
	//
	//lcd := Taxon{
	//	Name:     "LCD",
	//	ParentID: television.ID,
	//	Parent:   &television,
	//}
	//db.Create(&lcd)
	//
	//plasma := Taxon{
	//	Name:     "Plasma",
	//	ParentID: television.ID,
	//	Parent:   &television,
	//}
	//db.Create(&plasma)
	//
	//mp3Player := Taxon{
	//	Name:     "MP3 Player",
	//	ParentID: portElectronics.ID,
	//	Parent:   &portElectronics,
	//}
	//db.Create(&mp3Player)
	//
	//// level 3
	//flash := Taxon{
	//	Name:     "Flash",
	//	ParentID: mp3Player.ID,
	//	Parent:   &mp3Player,
	//}
	//db.Create(&flash)
	//
	//flash.Parent = &gameConsoles
	//flash.ParentID = gameConsoles.ID
	//
	//db.Save(&flash)

	//b := Taxon{
	//	Name: "B",
	//	ParentID: rootA.ID,
	//	Parent:   &rootA,
	//}
	//db.Create(&b)

	//| ELECTRONICS           |
	//|  TELEVISIONS          |
	//|   TUBE                |
	//|   LCD                 |
	//|   PLASMA              |
	//|  GAME CONSOLES        |
	//|  PORTABLE ELECTRONICS |
	//|   MP3 PLAYERS         |
	//|    FLASH              |
	//|   CD PLAYERS          |
	//|   2 WAY RADIOS
}

func getDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(true)

	return db
}
