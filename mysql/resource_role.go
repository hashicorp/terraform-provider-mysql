package mysql

import (
	"fmt"
	"log"

	"errors"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Create: CreateRole,
		Read:   ReadRole,
		Delete: DeleteRole,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func CreateRole(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	roleName := d.Get("name").(string)

	sql := fmt.Sprintf("CREATE ROLE '%s'", roleName)
	log.Printf("[DEBUG] SQL: ", sql)

	_, err := db.Exec(sql)
	if err != nil {
		return fmt.Errorf("error creating role: %s", err)
	}

	d.SetId(roleName)

	return nil
}

func ReadRole(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("SHOW GRANTS FOR '%s'", d.Id())
	log.Printf("[DEBUG] SQL: ", sql)

	_, err := db.Exec(err)
	if err != nil {
		log.Printf("[WARN] Role (%s) not found; removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func DeleteRole(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("DROP ROLE '%s'", d.Get("name").(string))
	log.Printf("[DEBUG] SQL: ", sql)

	_, err = db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	return nil
}
