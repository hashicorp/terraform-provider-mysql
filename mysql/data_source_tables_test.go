package mysql

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceTables(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTablesConfig_basic("mysql", "%"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mysql_tables.test", "database", "mysql"),
					resource.TestCheckResourceAttr("data.mysql_tables.test", "pattern", "%"),
					testAccTablesCount("data.mysql_tables.test", "tables.#", func(rn string, table_count int) error {
						if table_count < 1 {
							return fmt.Errorf("%s: tables not found", rn)
						}

						return nil
					}),
				),
			},
			{
				Config: testAccTablesConfig_basic("mysql", "__table_does_not_exist__"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mysql_tables.test", "database", "mysql"),
					resource.TestCheckResourceAttr("data.mysql_tables.test", "pattern", "__table_does_not_exist__"),
					testAccTablesCount("data.mysql_tables.test", "tables.#", func(rn string, table_count int) error {
						if table_count > 0 {
							return fmt.Errorf("%s: unexpected table found", rn)
						}

						return nil
					}),
				),
			},
		},
	})
}

func testAccTablesCount(rn string, key string, check func(string, int) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]

		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		value, ok := rs.Primary.Attributes[key]

		if !ok {
			return fmt.Errorf("%s: attribute '%s' not found", rn, key)
		}

		table_count, err := strconv.Atoi(value)

		if err != nil {
			return err
		}

		return check(rn, table_count)
	}
}

func testAccTablesConfig_basic(database string, pattern string) string {
	return fmt.Sprintf(`
data "mysql_tables" "test" {
		database = "%s"
		pattern = "%s"
}`, database, pattern)
}
