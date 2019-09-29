package mysql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatabase(t *testing.T) {
	dbName := "terraform_acceptance_test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDatabaseCheckDestroy(dbName),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(dbName),
				Check: testAccDatabaseCheck_basic(
					"mysql_database.test", dbName,
				),
			},
		},
	})
}

func TestAccDatabase_collationChange(t *testing.T) {
	dbName := "terraform_acceptance_test"

	charset1 := "latin1"
	charset2 := "utf8"
	collation1 := "latin1_bin"
	collation2 := "utf8_general_ci"

	resourceName := "mysql_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDatabaseCheckDestroy(dbName),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_full(dbName, charset1, collation1),
				Check: resource.ComposeTestCheckFunc(
					testAccDatabaseCheck_full("mysql_database.test", dbName, charset1, collation1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					db, err := connectToMySQL(testAccProvider.Meta().(*MySQLConfiguration))
					if err != nil {
						return
					}

					db.Exec(fmt.Sprintf("ALTER DATABASE %s CHARACTER SET %s COLLATE %s", dbName, charset2, collation2))
				},
				Config: testAccDatabaseConfig_full(dbName, charset1, collation1),
				Check: resource.ComposeTestCheckFunc(
					testAccDatabaseCheck_full(resourceName, dbName, charset1, collation1),
				),
			},
		},
	})
}

func testAccDatabaseCheck_basic(rn string, name string) resource.TestCheckFunc {
	return testAccDatabaseCheck_full(rn, name, "utf8", "utf8_bin")
}

func testAccDatabaseCheck_full(rn string, name string, charset string, collation string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("database id not set")
		}

		db, err := connectToMySQL(testAccProvider.Meta().(*MySQLConfiguration))
		if err != nil {
			return err
		}

		var _name, createSQL string
		err = db.QueryRow(fmt.Sprintf("SHOW CREATE DATABASE %s", name)).Scan(&_name, &createSQL)
		if err != nil {
			return fmt.Errorf("error reading database: %s", err)
		}

		if strings.Index(createSQL, fmt.Sprintf("CHARACTER SET %s", charset)) == -1 {
			return fmt.Errorf("database default charset isn't %s", charset)
		}
		if strings.Index(createSQL, fmt.Sprintf("COLLATE %s", collation)) == -1 {
			return fmt.Errorf("database default collation isn't %s", collation)
		}

		return nil
	}
}

func testAccDatabaseCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db, err := connectToMySQL(testAccProvider.Meta().(*MySQLConfiguration))
		if err != nil {
			return err
		}

		var _name, createSQL string
		err = db.QueryRow(fmt.Sprintf("SHOW CREATE DATABASE %s", name)).Scan(&_name, &createSQL)
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

func testAccDatabaseConfig_basic(name string) string {
	return testAccDatabaseConfig_full(name, "utf8", "utf8_bin")
}

func testAccDatabaseConfig_full(name string, charset string, collation string) string {
	return fmt.Sprintf(`
resource "mysql_database" "test" {
    name = "%s"
    default_character_set = "%s"
    default_collation = "%s"
}`, name, charset, collation)
}
