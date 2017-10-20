package mysql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDatabase(t *testing.T) {
	var dbName string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDatabaseCheckDestroy(dbName),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDatabaseConfig_basic,
				Check: testAccDatabaseCheck(
					"mysql_database.test", &dbName,
				),
			},
		},
	})
}

func testAccDatabaseCheck(rn string, name *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("database id not set")
		}

		db := testAccProvider.Meta().(*providerConfiguration).DB
		rows, err := db.Query("SHOW CREATE DATABASE terraform_acceptance_test")
		if err != nil {
			return fmt.Errorf("error reading database: %s", err)
		}
		defer rows.Close()

		rows.Next()
		var _name, createSQL string
		err = rows.Scan(&_name, &createSQL)
		if err != nil {
			return fmt.Errorf("error scanning create statement: %s", err)
		}

		if strings.Index(createSQL, "CHARACTER SET utf8") == -1 {
			return fmt.Errorf("database default charset isn't utf8")
		}
		if strings.Index(createSQL, "COLLATE utf8_bin") == -1 {
			return fmt.Errorf("database default collation isn't utf8_bin")
		}

		if rows.Next() {
			return fmt.Errorf("expected 1 row reading database, but got more")
		}

		*name = rs.Primary.ID

		return nil
	}
}

func testAccDatabaseCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*providerConfiguration).DB

		var name, createSQL string
		err := db.QueryRow("SHOW CREATE DATABASE terraform_acceptance_test").Scan(&name, &createSQL)
		if err == nil {
			return fmt.Errorf("database still exists after destroy")
		}

		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == unknownDatabaseErrCode {
				return nil
			}
		}

		return fmt.Errorf("got unexpected error: %s", err)
	}
}

const testAccDatabaseConfig_basic = `
resource "mysql_database" "test" {
    name = "terraform_acceptance_test"
    default_character_set = "utf8"
    default_collation = "utf8_bin"
}
`
