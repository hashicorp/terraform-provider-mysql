package mysql

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	mysqlc "github.com/ziutek/mymysql/mysql"
	mysqlts "github.com/ziutek/mymysql/thrsafe"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

type providerConfiguration struct {
	Conn                        mysqlc.Conn
	DefaultAuthenticationPlugin string
	ServerVersion               *version.Version
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
			"mysql_database": resourceDatabase(),
			"mysql_user":     resourceUser(),
			"mysql_grant":    resourceGrant(),
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

	// mysqlts is the thread-safe implementation of mymysql, so we can
	// safely re-use the same connection between multiple parallel
	// operations.
	conn := mysqlts.New(proto, "", endpoint, username, password)

	err := conn.Connect()
	if err != nil {
		return nil, err
	}

	ver, err := serverVersion(conn)
	if err != nil {
		return nil, err
	}

	authPlugin, err := defaultAuthenticationPlugin(conn)
	if err != nil {
		return nil, err
	}

	return &providerConfiguration{
		Conn: conn,
		DefaultAuthenticationPlugin: authPlugin,
		ServerVersion:               ver,
	}, nil
}

var identQuoteReplacer = strings.NewReplacer("`", "``")

func quoteIdentifier(in string) string {
	return fmt.Sprintf("`%s`", identQuoteReplacer.Replace(in))
}

func serverVersion(conn mysqlc.Conn) (*version.Version, error) {
	rows, _, err := conn.Query("SELECT VERSION()")
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("SELECT VERSION() returned an empty set")
	}

	return version.NewVersion(rows[0].Str(0))
}

func defaultAuthenticationPlugin(conn mysqlc.Conn) (string, error) {
	stmtSQL := "SELECT @@default_authentication_plugin"
	rows, _, err := conn.Query(stmtSQL)

	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("%s returned an empty set", stmtSQL)
	}

	return rows[0].Str(0), nil
}
