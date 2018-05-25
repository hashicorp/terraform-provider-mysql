package mysql

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"testing"
)

func TestAccUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccUserConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_user.test", "plaintext_password", hashSum("password")),
				),
			},
			resource.TestStep{
				Config: testAccUserConfig_newPass,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_user.test", "plaintext_password", hashSum("password2")),
				),
			},
		},
	})
}

func TestAccUser_auth(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccUserConfig_auth_iam_plugin,
				Check: resource.ComposeTestCheckFunc(
					testAccUserAuthExists("mysql_user.test"),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_user.test", "auth_plugin", "mysql_no_login"),
				),
			},
		},
	})
}

func TestAccUser_deprecated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccUserConfig_deprecated,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_user.test", "password", "password"),
				),
			},
			resource.TestStep{
				Config: testAccUserConfig_deprecated_newPass,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "example.com"),
					resource.TestCheckResourceAttr("mysql_user.test", "password", "password2"),
				),
			},
		},
	})
}

func testAccUserExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user id not set")
		}

		data := testAccProvider.Meta().(*providerConfiguration).Data
		db, err := sql.Open("mysql", data)

		if err != nil {
			return nil
		}
		stmtSQL := fmt.Sprintf("SELECT count(*) from mysql.user where CONCAT(user, '@', host) = '%s'", rs.Primary.ID)
		log.Println("Executing statement:", stmtSQL)
		var count int
		err = db.QueryRow(stmtSQL).Scan(&count)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("expected 1 row reading user but got no rows")
			}
			return fmt.Errorf("error reading user: %s", err)
		}
		defer db.Close()
		return nil
	}
}

func testAccUserAuthExists(rn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user id not set")
		}

		data := testAccProvider.Meta().(*providerConfiguration).Data
		db, err := sql.Open("mysql", data)

		if err != nil {
			return nil
		}
		stmtSQL := fmt.Sprintf("SELECT count(*) from mysql.user where CONCAT(user, '@', host) = '%s' and plugin = 'mysql_no_login'", rs.Primary.ID)
		log.Println("Executing statement:", stmtSQL)
		var count int
		err = db.QueryRow(stmtSQL).Scan(&count)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("expected 1 row reading user but got no rows")
			}
			return fmt.Errorf("error reading user: %s", err)
		}
		defer db.Close()
		return nil
	}
}

func testAccUserCheckDestroy(s *terraform.State) error {
	data := testAccProvider.Meta().(*providerConfiguration).Data
	db, err := sql.Open("mysql", data)

	if err != nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mysql_user" {
			continue
		}

		stmtSQL := fmt.Sprintf("SELECT user from mysql.user where CONCAT(user, '@', host) = '%s'", rs.Primary.ID)
		log.Println("Executing statement:", stmtSQL)
		rows, err := db.Query(stmtSQL)
		if err != nil {
			return fmt.Errorf("error issuing query: %s", err)
		}
		defer rows.Close()
		if rows.Next() {
			return fmt.Errorf("user still exists after destroy")
		}
	}
	defer db.Close()
	return nil
}

const testAccUserConfig_basic = `
resource "mysql_user" "test" {
    user = "jdoe"
    host = "example.com"
    plaintext_password = "password"
}
`

const testAccUserConfig_newPass = `
resource "mysql_user" "test" {
    user = "jdoe"
    host = "example.com"
    plaintext_password = "password2"
}
`

const testAccUserConfig_deprecated = `
resource "mysql_user" "test" {
    user = "jdoe"
    host = "example.com"
    password = "password"
}
`

const testAccUserConfig_deprecated_newPass = `
resource "mysql_user" "test" {
    user = "jdoe"
    host = "example.com"
    password = "password2"
}
`

const testAccUserConfig_auth_iam_plugin = `
resource "mysql_user" "test" {
    user        = "jdoe"
    host        = "example.com"
    auth_plugin = "mysql_no_login"
}
`
