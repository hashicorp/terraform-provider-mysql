package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

type MySQLConfiguration struct {
	Config *mysql.Config
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
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_PASSWORD", nil),
			},

			"tls": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_TLS_CONFIG", "false"),
				/*
					ValidateFunc: validation.StringInSlice([]string{
						"true",
						"false",
						"skip-verify",
					}, false),
				*/
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"mysql_database": resourceDatabase(),
			"mysql_user":     resourceUser(),
			"mysql_grant":    resourceGrant(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var endpoint = d.Get("endpoint").(string)

	proto := "tcp"
	if len(endpoint) > 0 && endpoint[0] == '/' {
		proto = "unix"
	}

	conf := mysql.Config{
		User:      d.Get("username").(string),
		Passwd:    d.Get("password").(string),
		Net:       proto,
		Addr:      endpoint,
		TLSConfig: d.Get("tls").(string),
	}

	return &MySQLConfiguration{
		Config: &conf,
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

func connectToMySQL(conf *mysql.Config) (*sql.DB, error) {
	dsn := conf.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to DB (%s): %s", dsn, err)
	}

	return db, nil
}
