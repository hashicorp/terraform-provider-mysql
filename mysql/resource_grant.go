package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-version"
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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"role"},
			},

			"role": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user", "host"},
			},

			"host": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Default:       "localhost",
				ConflictsWith: []string{"role"},
			},

			"database": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"table": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "*",
			},

			"privileges": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"roles": &schema.Schema{
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"privileges"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
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

func flattenList(list []interface{}, template string) string {
	var result []string
	for _, v := range list {
		result = append(result, fmt.Sprintf(template, v.(string)))
	}

	return strings.Join(result, ", ")
}

func formatDatabaseName(database string) string {
	if strings.Compare(database, "*") != 0 && !strings.HasSuffix(database, "`") {
		return fmt.Sprintf("`%s`", database)
	}

	return database
}

func userOrRole(user string, host string, role string, hasRoles bool) (string, bool, error) {
	if len(user) > 0 && len(host) > 0 {
		return fmt.Sprintf("'%s'@'%s'", user, host), false, nil
	} else if len(role) > 0 {
		if !hasRoles {
			return "", false, fmt.Errorf("Roles are only supported on MySQL 8 and above")
		}

		return fmt.Sprintf("'%s'", role), true, nil
	} else {
		return "", false, fmt.Errorf("user with host or a role is required")
	}
}

func supportsRoles(db *sql.DB) (bool, error) {
	currentVersion, err := serverVersion(db)
	if err != nil {
		return false, err
	}

	requiredVersion, _ := version.NewVersion("8.0.0")
	hasRoles := currentVersion.GreaterThan(requiredVersion)
	return hasRoles, nil
}

func CreateGrant(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	hasRoles, err := supportsRoles(db)
	if err != nil {
		return err
	}

	var (
		privilegesOrRoles string
		grantOn           string
	)

	hasPrivs := false
	rolesGranted := 0
	if attr, ok := d.GetOk("privileges"); ok {
		privilegesOrRoles = flattenList(attr.(*schema.Set).List(), "%s")
		hasPrivs = true
	} else if attr, ok := d.GetOk("roles"); ok {
		if !hasRoles {
			return fmt.Errorf("Roles are only supported on MySQL 8 and above")
		}
		listOfRoles := attr.(*schema.Set).List()
		rolesGranted = len(listOfRoles)
		privilegesOrRoles = flattenList(listOfRoles, "'%s'")
	} else {
		return fmt.Errorf("One of privileges or roles is required")
	}

	user := d.Get("user").(string)
	host := d.Get("host").(string)
	role := d.Get("role").(string)

	userOrRole, isRole, err := userOrRole(user, host, role, hasRoles)
	if err != nil {
		return err
	}

	database := formatDatabaseName(d.Get("database").(string))

	if (!isRole || hasPrivs) && rolesGranted == 0 {
		grantOn = fmt.Sprintf(" ON %s.%s", database, d.Get("table").(string))
	}

	stmtSQL := fmt.Sprintf("GRANT %s%s TO %s",
		privilegesOrRoles,
		grantOn,
		userOrRole)

	// MySQL 8+ doesn't allow REQUIRE on a GRANT statement.
	if !hasRoles {
		stmtSQL += fmt.Sprintf(" REQUIRE %s", d.Get("tls_option").(string))
	}

	if !hasRoles && !isRole && d.Get("grant").(bool) {
		stmtSQL += " WITH GRANT OPTION"
	}

	log.Println("Executing statement:", stmtSQL)
	_, err = db.Exec(stmtSQL)
	if err != nil {
		return fmt.Errorf("Error running SQL (%s): %s", stmtSQL, err)
	}

	id := fmt.Sprintf("%s@%s:%s", user, host, database)
	if isRole {
		id = fmt.Sprintf("%s:%s", role, database)
	}

	d.SetId(id)

	return ReadGrant(d, meta)
}

func ReadGrant(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	hasRoles, err := supportsRoles(db)
	if err != nil {
		return err
	}

	userOrRole, _, err := userOrRole(
		d.Get("user").(string),
		d.Get("host").(string),
		d.Get("role").(string),
		hasRoles)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("SHOW GRANTS FOR %s", userOrRole)

	log.Println("[DEBUG] SQL:", sql)

	_, err = db.Exec(sql)
	if err != nil {
		log.Printf("[WARN] GRANT not found for %s - removing from state", userOrRole)
		d.SetId("")
	}

	return nil
}

func DeleteGrant(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration).Config)
	if err != nil {
		return err
	}

	database := formatDatabaseName(d.Get("database").(string))

	hasRoles, err := supportsRoles(db)
	if err != nil {
		return err
	}

	userOrRole, isRole, err := userOrRole(
		d.Get("user").(string),
		d.Get("host").(string),
		d.Get("role").(string),
		hasRoles)
	if err != nil {
		return err
	}

	roles := d.Get("roles").(*schema.Set)

	var sql string
	if !isRole && len(roles.List()) == 0 {
		sql = fmt.Sprintf("REVOKE GRANT OPTION ON %s.%s FROM %s",
			database,
			d.Get("table").(string),
			userOrRole)

		log.Printf("[DEBUG] SQL: %s", sql)
		_, err = db.Exec(sql)
		if err != nil {
			return fmt.Errorf("error revoking GRANT (%s): %s", sql, err)
		}
	}

	whatToRevoke := fmt.Sprintf("ALL ON %s.%s", database, d.Get("table").(string))
	if len(roles.List()) > 0 {
		whatToRevoke = flattenList(roles.List(), "'%s'")
	}

	sql = fmt.Sprintf("REVOKE %s FROM %s", whatToRevoke, userOrRole)
	log.Printf("[DEBUG] SQL: %s", sql)
	_, err = db.Exec(sql)
	if err != nil {
		return fmt.Errorf("error revoking ALL (%s): %s", sql, err)
	}

	return nil
}
