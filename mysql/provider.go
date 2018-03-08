package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

type providerConfiguration struct {
	DB            *sql.DB
	ServerVersion *version.Version
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_ENDPOINT", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("Endpoint must not be an empty string"))
					}

					return
				},
			},

			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_USERNAME", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("Username must not be an empty string"))
					}

					return
				},
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_PASSWORD", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"mysql_database":      resourceDatabase(),
			"mysql_user":          resourceUser(),
			"mysql_user_password": resourceUserPassword(),
			"mysql_grant":         resourceGrant(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var username = d.Get("username").(string)
	var password = d.Get("password").(string)
	var endpoint = d.Get("endpoint").(string)

	proto := "tcp"
	if len(endpoint) > 0 && endpoint[0] == '/' {
		proto = "unix"
	}

	// database/sql is the thread-safe by default, so we can
	// safely re-use the same handle between multiple parallel
	// operations.

	dataSourceName := fmt.Sprintf("%s:%s@%s(%s)/", username, password, proto, endpoint)
	db, err := sql.Open("mysql", dataSourceName)

	ver, err := serverVersion(db)
	if err != nil {
		return nil, err
	}

	return &providerConfiguration{
		DB:            db,
		ServerVersion: ver,
	}, nil
}

var identQuoteReplacer = strings.NewReplacer("`", "``")

func quoteIdentifier(in string) string {
	return fmt.Sprintf("`%s`", identQuoteReplacer.Replace(in))
}

func serverVersion(db *sql.DB) (*version.Version, error) {
	rows, err := db.Query("SELECT VERSION()")
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("SELECT VERSION() returned an empty set")
	}

	var versionString string
	rows.Scan(&versionString)
	return version.NewVersion(versionString)
}
