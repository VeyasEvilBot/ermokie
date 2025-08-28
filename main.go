package main

import (
	"fmt"
	"log"

	"github.com/stefanistkuhl/ermokie/pkg/db"
	// hcmmigration "github.com/stefanistkuhl/ermokie/pkg/hcmMigration"
)

func main() {
	database := db.Init()
	runs, err := db.GetAllRuns(database)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range runs {
		fmt.Println("thingy:", r.ID, r.Name)
	}
	// pl, err := hcmmigration.LoadProfiles("HitCounterManagerInit.xml")
	// if err != nil {
	// 	panic(err)
	// }
	//	if db.CheckEmpty(database) {
	//		err = hcmmigration.ImportProfiles(database, pl)
	//		if err != nil {
	//			panic(err)
	//		}
	//	} else {
	//
	//		fmt.Println("DB already has data")
	//	}
}
