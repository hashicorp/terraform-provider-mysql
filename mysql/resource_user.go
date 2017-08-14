package mysql

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: CreateUser,
		Update: UpdateUser,
		Read:   ReadUser,
		Delete: DeleteUser,

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

			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"auth_plugin", "hash_string"},
			},

			"auth_plugin": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"hash_string": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func CreateUser(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*providerConfiguration).Conn

	stmtSQL := fmt.Sprintf("CREATE USER '%s'@'%s'",
		d.Get("user").(string),
		d.Get("host").(string))

	password := d.Get("password").(string)
	authPlugin := d.Get("auth_plugin").(string)
	hashString := d.Get("hash_string").(string)

	if password != "" {
		stmtSQL = stmtSQL + fmt.Sprintf(" IDENTIFIED BY '%s'", password)
	}

	if authPlugin != "" {
		stmtSQL = stmtSQL + fmt.Sprintf(" IDENTIFIED WITH '%s'", authPlugin)

		if hashString != "" {
			stmtSQL = stmtSQL + fmt.Sprintf(" AS '%s'", hashString)
		}
	}

	log.Println("Executing statement:", stmtSQL)
	_, _, err := conn.Query(stmtSQL)
	if err != nil {
		return err
	}

	user := fmt.Sprintf("%s@%s", d.Get("user").(string), d.Get("host").(string))
	d.SetId(user)

	return nil
}

func UpdateUser(d *schema.ResourceData, meta interface{}) error {
	conf := meta.(*providerConfiguration)

	if d.HasChange("password") || d.HasChange("auth_plugin") || d.HasChange("hash_string") {
		var stmtSQL string

		if _, hasPassword := d.GetOk("password"); hasPassword {
			_, newPassword := d.GetChange("password")
			ver, _ := version.NewVersion("5.7.6")

			/* ALTER USER syntax introduced in MySQL 5.7.6 deprecates SET PASSWORD (GH-8230) */
			if conf.ServerVersion.LessThan(ver) {
				if d.HasChange("auth_plugin") || d.HasChange("hash_string") {
					return fmt.Errorf(
						"The 'auth_plugin' and 'hash_string' parameters are " +
							"not supported on MySQL versions earlier than 5.7.6")
				}

				stmtSQL = fmt.Sprintf(
					"SET PASSWORD FOR '%s'@'%s' = PASSWORD('%s')",
					d.Get("user").(string),
					d.Get("host").(string),
					newPassword.(string))
			} else {
				stmtSQL = fmt.Sprintf(
					"ALTER USER '%s'@'%s' IDENTIFIED WITH %s BY '%s'",
					d.Get("user").(string),
					d.Get("host").(string),
					conf.DefaultAuthenticationPlugin,
					newPassword.(string))
			}
		} else {
			stmtSQL = fmt.Sprintf(
				"ALTER USER '%s'@'%s' IDENTIFIED WITH '%s' AS '%s'",
				d.Get("user").(string),
				d.Get("host").(string),
				d.Get("auth_plugin").(string),
				d.Get("hash_string").(string))
		}

		log.Println("Executing statement:", stmtSQL)
		_, _, err := conf.Conn.Query(stmtSQL)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadUser(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*providerConfiguration).Conn

	stmtSQL := fmt.Sprintf("SELECT USER FROM mysql.user WHERE USER='%s'",
		d.Get("user").(string))

	log.Println("Executing statement:", stmtSQL)

	rows, _, err := conn.Query(stmtSQL)
	log.Println("Returned rows:", len(rows))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		d.SetId("")
	}
	return nil
}

func DeleteUser(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*providerConfiguration).Conn

	stmtSQL := fmt.Sprintf("DROP USER '%s'@'%s'",
		d.Get("user").(string),
		d.Get("host").(string))

	log.Println("Executing statement:", stmtSQL)

	_, _, err := conn.Query(stmtSQL)
	if err == nil {
		d.SetId("")
	}
	return err
}
