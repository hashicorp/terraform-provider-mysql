package mysql

import (
	"fmt"
	"log"
	"strings"

	"errors"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/helper/schema"
)

type mySQLUser struct {
	User        string
	Host        string
	SSLType     string
	SSLCipher   string
	X509Issuer  string
	X509Subject string
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: CreateUser,
		Update: UpdateUser,
		Read:   ReadUser,
		Delete: DeleteUser,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"host": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "localhost",
			},

			"plaintext_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				StateFunc: hashSum,
			},

			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"plaintext_password"},
				Sensitive:     true,
				Deprecated:    "Please use plaintext_password instead",
			},

			"auth_plugin": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"plaintext_password", "password"},
			},

			"tls_option": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "NONE",
			},
		},
	}
}

func CreateUser(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
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

	requiredVersion, _ := version.NewVersion("5.7.0")
	currentVersion, err := serverVersion(db)
	if err != nil {
		return err
	}

	if currentVersion.GreaterThan(requiredVersion) {
		stmtSQL += fmt.Sprintf(" REQUIRE %s", d.Get("tls_option").(string))
	}

	log.Println("Executing statement:", stmtSQL)
	_, err = db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	user := fmt.Sprintf("%s@%s", d.Get("user").(string), d.Get("host").(string))
	d.SetId(user)

	return nil
}

func UpdateUser(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
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
		serverVersion, err := serverVersion(db)
		if err != nil {
			return fmt.Errorf("Could not determine server version: %s", err)
		}

		ver, _ := version.NewVersion("5.7.6")
		if serverVersion.LessThan(ver) {
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

	requiredVersion, _ := version.NewVersion("5.7.0")
	currentVersion, err := serverVersion(db)
	if err != nil {
		return err
	}

	if d.HasChange("tls_option") && currentVersion.GreaterThan(requiredVersion) {
		var stmtSQL string

		stmtSQL = fmt.Sprintf("ALTER USER '%s'@'%s'  REQUIRE %s",
			d.Get("user").(string),
			d.Get("host").(string),
			fmt.Sprintf(" REQUIRE %s", d.Get("tls_option").(string)))

		log.Println("Executing query:", stmtSQL)
		_, err := db.Exec(stmtSQL)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTLSOption(user mySQLUser) string {
	var tlsOption string
	switch user.SSLType {
	case "":
		tlsOption = "NONE"
	case "ANY":
		tlsOption = "SSL"
	case "X509":
		tlsOption = "X509"
	case "SPECIFIED":
		var params []string
		if user.X509Subject != "" {
			params = append(params, " SUBJECT '"+user.X509Subject+"'")
		}
		if user.X509Issuer != "" {
			params = append(params, " ISSUER '"+user.X509Issuer+"'")
		}
		if user.SSLCipher != "" {
			params = append(params, " CIPHER '"+user.SSLCipher+"'")
		}
		tlsOption = strings.Join(params, " AND")
	}
	log.Println("tls_option: ", tlsOption)
	return tlsOption
}

func ReadUser(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	id := strings.Split(d.Id(), "@")
	user, host := id[0], id[1]
	stmtSQL := "SELECT user,host,ssl_type,ssl_cipher,x509_issuer,x509_subject FROM mysql.user WHERE USER=? and HOST=?"
	log.Println("Executing statement:", stmtSQL)

	rows, err := db.Query(stmtSQL, user, host)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() && rows.Err() == nil {
		d.SetId("")
	} else {
		/*		var (
					db_user      string
					db_host      string
					ssl_type     string
					ssl_cipher   string
					x509_issuer  string
					x509_subject string
				)
		*/
		var user mySQLUser
		err = rows.Scan(&user.User, &user.Host, &user.SSLType, &user.SSLCipher, &user.X509Issuer, &user.X509Subject)
		d.Set("host", user.Host)
		d.Set("user", user.User)
		tlsOption := getTLSOption(user)
		d.Set("tls_option", tlsOption)

		/* ssl_type can be ANY, X509 or SPECIFIED */

	}

	return rows.Err()
}

func DeleteUser(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	stmtSQL := fmt.Sprintf("DROP USER '%s'@'%s'",
		d.Get("user").(string),
		d.Get("host").(string))

	log.Println("Executing statement:", stmtSQL)

	_, err = db.Exec(stmtSQL)
	if err == nil {
		d.SetId("")
	}
	return err
}
