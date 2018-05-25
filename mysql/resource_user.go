package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
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

			"plaintext_password": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				StateFunc: hashSum,
			},

			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"plaintext_password"},
				Sensitive:     true,
				Deprecated:    "Please use plaintext_password instead",
			},

			"auth_plugin": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"plaintext_password", "password"},
			},
		},
	}
}

func CreateUser(d *schema.ResourceData, meta interface{}) error {
	data := meta.(*providerConfiguration).Data
	db, err := sql.Open("mysql", data)

	if err != nil {
		return err
	}

	var authStm string
	var auth string
	if v, ok := d.GetOk("auth_plugin"); ok {
		auth = v.(string)
	}

	if len(auth) > 0 {
		switch auth {
		case "AWSAuthenticationPlugin":
			authStm = " IDENTIFIED WITH AWSAuthenticationPlugin as 'RDS'"
		case "mysql_no_login":
			authStm = " IDENTIFIED WITH mysql_no_login"
		}
	}

	stmtSQL := fmt.Sprintf("CREATE USER '%s'@'%s'",
		d.Get("user").(string),
		d.Get("host").(string))

	var password string
	if v, ok := d.GetOk("plaintext_password"); ok {
		password = v.(string)
	} else {
		password = d.Get("password").(string)
	}

	if auth == "AWSAuthenticationPlugin" && d.Get("host").(string) == "localhost" {
		return errors.New("cannot use IAM auth against localhost")
	}

	if authStm != "" {
		stmtSQL = stmtSQL + authStm
	} else {
		stmtSQL = stmtSQL + fmt.Sprintf(" IDENTIFIED BY '%s'", password)
	}

	log.Println("Executing statement:", stmtSQL)
	_, err = db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	user := fmt.Sprintf("%s@%s", d.Get("user").(string), d.Get("host").(string))
	d.SetId(user)
	defer db.Close()
	return nil
}

func UpdateUser(d *schema.ResourceData, meta interface{}) error {
	data := meta.(*providerConfiguration).Data
	db, err := sql.Open("mysql", data)

	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT VERSION()")
	if err != nil {
		return err
	}
	if !rows.Next() {
		return nil
	}

	var versionString string
	rows.Scan(&versionString)
	dbversion, err := version.NewVersion(versionString)

	if err != nil {
		return err
	}
	var auth string
	if v, ok := d.GetOk("auth_plugin"); ok {
		auth = v.(string)
	}

	if len(auth) > 0 {
		// nothing to change, return
		return nil
	}

	var newpw interface{}
	if d.HasChange("plaintext_password") {
		_, newpw = d.GetChange("plaintext_password")
	} else if d.HasChange("password") {
		_, newpw = d.GetChange("password")
	} else {
		newpw = nil
	}

	if newpw != nil {
		var stmtSQL string

		/* ALTER USER syntax introduced in MySQL 5.7.6 deprecates SET PASSWORD (GH-8230) */
		ver, _ := version.NewVersion("5.7.6")
		if dbversion.LessThan(ver) {
			stmtSQL = fmt.Sprintf("SET PASSWORD FOR '%s'@'%s' = PASSWORD('%s')",
				d.Get("user").(string),
				d.Get("host").(string),
				newpw.(string))
		} else {
			stmtSQL = fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'",
				d.Get("user").(string),
				d.Get("host").(string),
				newpw.(string))
		}

		log.Println("Executing query:", stmtSQL)
		_, err = db.Exec(stmtSQL)
		if err != nil {
			return err
		}
	}
	defer db.Close()
	return nil
}

func ReadUser(d *schema.ResourceData, meta interface{}) error {
	data := meta.(*providerConfiguration).Data
	db, err := sql.Open("mysql", data)

	if err != nil {
		return err
	}

	stmtSQL := fmt.Sprintf("SELECT USER FROM mysql.user WHERE USER='%s'",
		d.Get("user").(string))

	log.Println("Executing statement:", stmtSQL)

	rows, err := db.Query(stmtSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() && rows.Err() == nil {
		d.SetId("")
	}
	defer db.Close()
	return rows.Err()
}

func DeleteUser(d *schema.ResourceData, meta interface{}) error {
	data := meta.(*providerConfiguration).Data
	db, err := sql.Open("mysql", data)

	if err != nil {
		log.Fatal(err)
	}

	stmtSQL := fmt.Sprintf("DROP USER '%s'@'%s'",
		d.Get("user").(string),
		d.Get("host").(string))

	log.Println("Executing statement:", stmtSQL)

	_, err = db.Exec(stmtSQL)
	if err == nil {
		d.SetId("")
	}
	defer db.Close()
	return err
}
