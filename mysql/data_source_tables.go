package mysql

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceTables() *schema.Resource {
	return &schema.Resource{
		Read: ShowTables,
		Schema: map[string]*schema.Schema{
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func ShowTables(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration))

	if err != nil {
		return err
	}

	database := d.Get("database").(string)
	pattern := d.Get("pattern").(string)

	sql := fmt.Sprintf("SHOW TABLES FROM %s", quoteIdentifier(database))

	if pattern != "" {
		sql += fmt.Sprintf(" LIKE '%s'", pattern)
	}

	log.Printf("[DEBUG] SQL: %s", sql)

	rows, err := db.Query(sql)

	if err != nil {
		return err
	}

	defer rows.Close()

	var tables []string

	for rows.Next() {
		var table string

		err := rows.Scan(&table)

		if err != nil {
			return err
		}

		tables = append(tables, table)
	}

	err = d.Set("tables", tables)

	if err != nil {
		return err
	}

	d.SetId(resource.UniqueId())

	return nil
}
