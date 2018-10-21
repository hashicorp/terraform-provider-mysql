package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform/helper/resource"
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
			"mysql_role":     resourceRole(),
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
	var versionString string
	err := db.QueryRow("SELECT @@GLOBAL.innodb_version").Scan(&versionString)
	if err != nil {
		return nil, err
	}

	return version.NewVersion(versionString)
}

func connectToMySQL(conf *mysql.Config) (*sql.DB, error) {
	dsn := conf.FormatDSN()
	var db *sql.DB
	var err error

	// When provisioning a database server there can often be a lag between
	// when Terraform thinks it's available and when it is actually available.
	// This is particularly acute when provisioning a server and then immediately
	// trying to provision a database on it.
	retryError := resource.Retry(5*time.Minute, func() *resource.RetryError {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return resource.RetryableError(err)
		}

		// The Go SDK for MySQL doesn't actually connect until a query is ran.
		// This forces a simple query to run, which runs a connect, which lets
		// the retry logic do its thing.
		_, err = serverVersion(db)
		if err != nil {
			return resource.RetryableError(err)
		}

		return nil
	})

	if retryError != nil {
		return nil, fmt.Errorf("Could not connect to server: %s", retryError)
	}

	return db, nil
}
