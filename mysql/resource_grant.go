package mysql

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

const nonexistingGrantErrCode = 1141

func resourceGrant() *schema.Resource {
	return &schema.Resource{
		Create: CreateGrant,
		Update: nil,
		Read:   ReadGrant,
		Delete: DeleteGrant,

		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"host": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "localhost",
			},

			"database": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"privileges": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"grant": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"tls_option": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "NONE",
			},
		},
	}
}

func CreateGrant(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	// create a comma-delimited string of privileges
	var privileges string
	var privilegesList []string
	vL := d.Get("privileges").(*schema.Set).List()
	for _, v := range vL {
		privilegesList = append(privilegesList, v.(string))
	}
	privileges = strings.Join(privilegesList, ",")

	stmtSQL := fmt.Sprintf("GRANT %s on %s.* TO '%s'@'%s'",
		privileges,
		d.Get("database").(string),
		d.Get("user").(string),
		d.Get("host").(string))

	stmtSQL += fmt.Sprintf(" REQUIRE %s", d.Get("tls_option").(string))

	if d.Get("grant").(bool) {
		stmtSQL += " WITH GRANT OPTION"
	}

	log.Println("Executing statement:", stmtSQL)
	_, err := db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	user := fmt.Sprintf("%s@%s:%s", d.Get("user").(string), d.Get("host").(string), d.Get("database"))
	d.SetId(user)

	return ReadGrant(d, meta)
}

func ReadGrant(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	stmtSQL := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'",
		d.Get("user").(string),
		d.Get("host").(string))

	log.Println("Executing statement:", stmtSQL)

	rows, err := db.Query(stmtSQL)
	if err != nil {
		d.SetId("")
	} else {
		rows.Close()
	}
	return nil
}

func DeleteGrant(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	stmtSQL := fmt.Sprintf("REVOKE GRANT OPTION ON %s.* FROM '%s'@'%s'",
		d.Get("database").(string),
		d.Get("user").(string),
		d.Get("host").(string))

	log.Println("Executing statement:", stmtSQL)
	_, err := db.Query(stmtSQL)
	if err != nil {
		return err
	}

	stmtSQL = fmt.Sprintf("REVOKE ALL ON %s.* FROM '%s'@'%s'",
		d.Get("database").(string),
		d.Get("user").(string),
		d.Get("host").(string))

	log.Println("Executing statement:", stmtSQL)
	_, err = db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	return nil
}
