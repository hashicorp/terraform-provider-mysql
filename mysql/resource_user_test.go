package mysql

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUser(t *testing.T) {
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
					resource.TestCheckResourceAttr("mysql_user.test", "password", "password"),
				),
			},
			resource.TestStep{
				Config: testAccUserConfig_newPass,
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

func TestAccUser_plugin(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccUserCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccUserConfig_plugin,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					testAccUserAuthenticationPlugin("mysql_user.test", "auth_socket", ""),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "localhost"),
					resource.TestCheckResourceAttr("mysql_user.test", "auth_plugin", "auth_socket"),
				),
			},
			resource.TestStep{
				Config: testAccUserConfig_plugin2,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					testAccUserAuthenticationPlugin("mysql_user.test", "mysql_native_password", "*2470C0C06DEE42FD1618BB99005ADCA2EC9D1E19"),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "localhost"),
					resource.TestCheckResourceAttr("mysql_user.test", "password", "password"),
				),
			},
			resource.TestStep{
				Config: testAccUserConfig_plugin3,
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists("mysql_user.test"),
					testAccUserAuthenticationPlugin("mysql_user.test", "auth_socket", ""),
					resource.TestCheckResourceAttr("mysql_user.test", "user", "jdoe"),
					resource.TestCheckResourceAttr("mysql_user.test", "host", "localhost"),
					resource.TestCheckResourceAttr("mysql_user.test", "auth_plugin", "auth_socket"),
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

		conn := testAccProvider.Meta().(*providerConfiguration).Conn
		stmtSQL := fmt.Sprintf("SELECT count(*) from mysql.user where CONCAT(user, '@', host) = '%s'", rs.Primary.ID)
		log.Println("Executing statement:", stmtSQL)
		rows, _, err := conn.Query(stmtSQL)
		if err != nil {
			return fmt.Errorf("error reading user: %s", err)
		}
		if len(rows) != 1 {
			return fmt.Errorf("expected 1 row reading user but got %d", len(rows))
		}

		return nil
	}
}

func testAccUserAuthenticationPlugin(rn string, authPlugin string, authString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("Resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("User ID not set")
		}

		conn := testAccProvider.Meta().(*providerConfiguration).Conn
		stmtSQL := fmt.Sprintf(
			"SELECT plugin, authentication_string FROM mysql.user "+
				"WHERE CONCAT(user, '@', host) = '%s'", rs.Primary.ID)
		log.Println("Executing statement:", stmtSQL)

		rows, _, err := conn.Query(stmtSQL)
		if err != nil {
			return fmt.Errorf("Error reading user: %s", err)
		}
		if len(rows) != 1 {
			return fmt.Errorf("Expected 1 row reading user but got %d rows", len(rows))
		}

		row := rows[0]
		if authPlugin != row.Str(0) {
			return fmt.Errorf("Expected plugin %s but got %s", authPlugin, row.Str(0))
		}
		if authString != row.Str(1) {
			return fmt.Errorf("Expected authentication_string %s but got %s", authString, row.Str(1))
		}

		return nil
	}
}

func testAccUserCheckDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*providerConfiguration).Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mysql_user" {
			continue
		}

		stmtSQL := fmt.Sprintf("SELECT user from mysql.user where CONCAT(user, '@', host) = '%s'", rs.Primary.ID)
		log.Println("Executing statement:", stmtSQL)
		rows, _, err := conn.Query(stmtSQL)
		if err != nil {
			return fmt.Errorf("error issuing query: %s", err)
		}
		if len(rows) != 0 {
			return fmt.Errorf("user still exists after destroy")
		}
	}
	return nil
}

const testAccUserConfig_basic = `
resource "mysql_user" "test" {
        user = "jdoe"
        host = "example.com"
        password = "password"
}
`

const testAccUserConfig_newPass = `
resource "mysql_user" "test" {
        user = "jdoe"
        host = "example.com"
        password = "password2"
}
`

const testAccUserConfig_plugin = `
resource "mysql_user" "test" {
    user        = "jdoe"
    host        = "localhost"
    auth_plugin = "auth_socket"
}
`

const testAccUserConfig_plugin2 = `
resource "mysql_user" "test" {
    user     = "jdoe"
    host     = "localhost"
    password = "password"
}
`

const testAccUserConfig_plugin3 = `
resource "mysql_user" "test" {
    user        = "jdoe"
    host        = "localhost"
    auth_plugin = "auth_socket"
}
`
