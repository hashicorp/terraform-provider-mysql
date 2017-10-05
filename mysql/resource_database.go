package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/hashicorp/terraform/helper/schema"
)

const defaultCharacterSetKeyword = "CHARACTER SET "
const defaultCollateKeyword = "COLLATE "
const unknownDatabaseErrCode = 1049

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: CreateDatabase,
		Update: UpdateDatabase,
		Read:   ReadDatabase,
		Delete: DeleteDatabase,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"default_character_set": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "utf8",
			},

			"default_collation": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "utf8_general_ci",
			},
		},
	}
}

func CreateDatabase(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	stmtSQL := databaseConfigSQL("CREATE", d)
	log.Println("Executing statement:", stmtSQL)

	_, err := db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))

	return nil
}

func UpdateDatabase(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	stmtSQL := databaseConfigSQL("ALTER", d)
	log.Println("Executing statement:", stmtSQL)

	_, err := db.Exec(stmtSQL)
	if err != nil {
		return err
	}

	return nil
}

func ReadDatabase(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	// This is kinda flimsy-feeling, since it depends on the formatting
	// of the SHOW CREATE DATABASE output... but this data doesn't seem
	// to be available any other way, so hopefully MySQL keeps this
	// compatible in future releases.

	name := d.Id()
	stmtSQL := "SHOW CREATE DATABASE " + quoteIdentifier(name)

	log.Println("Executing query:", stmtSQL)
	var createSQL, _database string
	err := db.QueryRow(stmtSQL).Scan(&_database, &createSQL)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == unknownDatabaseErrCode {
				d.SetId("")
				return nil
			}
		}
		return fmt.Errorf("Error during show create database: %s", err)
	}

	defaultCharset := extractIdentAfter(createSQL, defaultCharacterSetKeyword)
	defaultCollation := extractIdentAfter(createSQL, defaultCollateKeyword)

	if defaultCollation == "" && defaultCharset != "" {
		// MySQL doesn't return the collation if it's the default one for
		// the charset, so if we don't have a collation we need to go
		// hunt for the default.
		stmtSQL := "SHOW COLLATION WHERE `Charset` = ? AND `Default` = 'Yes'"
		var defaultCollation string
		var empty interface{}
		err := db.QueryRow(stmtSQL, defaultCharset).Scan(&defaultCollation, &empty, &empty, &empty, &empty, &empty)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("Charset %s has no default collation", defaultCharset)
			}
			return fmt.Errorf("Error getting default charset: %s, %s", err, defaultCharset)
		}
		return err
	}

	d.Set("default_character_set", defaultCharset)
	d.Set("default_collation", defaultCollation)

	return nil
}

func DeleteDatabase(d *schema.ResourceData, meta interface{}) error {
	db := meta.(*providerConfiguration).DB

	name := d.Id()
	stmtSQL := "DROP DATABASE " + quoteIdentifier(name)
	log.Println("Executing statement:", stmtSQL)

	_, err := db.Exec(stmtSQL)
	if err == nil {
		d.SetId("")
	}
	return err
}

func databaseConfigSQL(verb string, d *schema.ResourceData) string {
	name := d.Get("name").(string)
	defaultCharset := d.Get("default_character_set").(string)
	defaultCollation := d.Get("default_collation").(string)

	var defaultCharsetClause string
	var defaultCollationClause string

	if defaultCharset != "" {
		defaultCharsetClause = defaultCharacterSetKeyword + quoteIdentifier(defaultCharset)
	}
	if defaultCollation != "" {
		defaultCollationClause = defaultCollateKeyword + quoteIdentifier(defaultCollation)
	}

	return fmt.Sprintf(
		"%s DATABASE %s %s %s",
		verb,
		quoteIdentifier(name),
		defaultCharsetClause,
		defaultCollationClause,
	)
}

func extractIdentAfter(sql string, keyword string) string {
	charsetIndex := strings.Index(sql, keyword)
	if charsetIndex != -1 {
		charsetIndex += len(keyword)
		remain := sql[charsetIndex:]
		spaceIndex := strings.IndexRune(remain, ' ')
		return remain[:spaceIndex]
	}

	return ""
}
