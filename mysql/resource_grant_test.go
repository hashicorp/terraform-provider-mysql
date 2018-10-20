package mysql

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGrant(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccGrantCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccGrantConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccPrivilegeExists("mysql_grant.test", "SELECT"),
					resource.TestCheckResourceAttr("mysql_grant.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_grant.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_grant.test", "database", "foo"),
					resource.TestCheckResourceAttr("mysql_grant.test", "tls_option", "NONE"),
				),
			},
			resource.TestStep{
				Config: testAccGrantConfig_ssl,
				Check: resource.ComposeTestCheckFunc(
					testAccPrivilegeExists("mysql_grant.test", "SELECT"),
					resource.TestCheckResourceAttr("mysql_grant.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_grant.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_grant.test", "database", "foo"),
					resource.TestCheckResourceAttr("mysql_grant.test", "tls_option", "SSL"),
				),
			},
		},
	})
}

func testAccPrivilegeExists(rn string, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("grant id not set")
		}

		id := strings.Split(rs.Primary.ID, ":")
		userhost := strings.Split(id[0], "@")
		user := userhost[0]
		host := userhost[1]

		db, err := connectToMySQL(testAccProvider.Meta().(*MySQLConfiguration).Config)
		if err != nil {
			return err
		}

		stmtSQL := fmt.Sprintf("SHOW GRANTS for '%s'@'%s'", user, host)
		log.Println("Executing statement:", stmtSQL)
		rows, err := db.Query(stmtSQL)
		if err != nil {
			return fmt.Errorf("error reading grant: %s", err)
		}
		defer rows.Close()

		privilegeFound := false
		for rows.Next() {
			var grants string
			err = rows.Scan(&grants)
			if err != nil {
				return fmt.Errorf("failed to read grant for '%s'@'%s': %s", user, host, err)
			}
			log.Printf("Result Row: %s", grants)
			privIndex := strings.Index(grants, privilege)
			if privIndex != -1 {
				privilegeFound = true
			}
		}

		if !privilegeFound {
			return fmt.Errorf("grant no found for '%s'@'%s'", user, host)
		}

		return nil
	}
}

func testAccGrantCheckDestroy(s *terraform.State) error {
	db, err := connectToMySQL(testAccProvider.Meta().(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mysql_grant" {
			continue
		}

		id := strings.Split(rs.Primary.ID, ":")
		userhost := strings.Split(id[0], "@")
		user := userhost[0]
		host := userhost[1]

		stmtSQL := fmt.Sprintf("SHOW GRANTS for '%s'@'%s'", user, host)
		log.Println("Executing statement:", stmtSQL)
		rows, err := db.Query(stmtSQL)
		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				if mysqlErr.Number == nonexistingGrantErrCode {
					return nil
				}
			}

			return fmt.Errorf("error reading grant: %s", err)
		}
		defer rows.Close()

		if rows.Next() {
			return fmt.Errorf("grant still exists for'%s'@'%s'", user, host)
		}
	}
	return nil
}

const testAccGrantConfig_basic = `
resource "mysql_user" "test" {
        user = "jdoe"
				host = "example.com"
				password = "password"
}

resource "mysql_grant" "test" {
        user = "${mysql_user.test.user}"
        host = "${mysql_user.test.host}"
        database = "foo"
        privileges = ["UPDATE", "SELECT"]
}
`

const testAccGrantConfig_ssl = `
resource "mysql_user" "test" {
        user = "jdoe"
				host = "example.com"
				password = "password"
}

resource "mysql_grant" "test" {
        user = "${mysql_user.test.user}"
        host = "${mysql_user.test.host}"
        database = "foo"
		privileges = ["UPDATE", "SELECT"]
		tls_option = "SSL"
}
`
