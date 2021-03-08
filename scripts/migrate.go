package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/cycloidio/terracost/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:terracost@tcp(172.44.0.2:3306)/terracost_test?multiStatements=true")
	if err != nil {
		log.Fatal(err)
	}

	if err := mysql.Migrate(context.Background(), db, "_migrations"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Migrated successfully!")
}
