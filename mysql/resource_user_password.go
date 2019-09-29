package mysql

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/helper/encryption"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceUserPassword() *schema.Resource {
	return &schema.Resource{
		Create: SetUserPassword,
		Read:   ReadUserPassword,
		Delete: DeleteUserPassword,
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
			"pgp_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_password": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func SetUserPassword(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration))
	if err != nil {
		return err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	password := uuid.String()
	pgpKey := d.Get("pgp_key").(string)
	encryptionKey, err := encryption.RetrieveGPGKey(pgpKey)
	if err != nil {
		return err
	}
	fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, password, "MySQL Password")
	if err != nil {
		return err
	}
	d.Set("key_fingerprint", fingerprint)
	d.Set("encrypted_password", encrypted)

	requiredVersion, _ := version.NewVersion("8.0.0")
	currentVersion, err := serverVersion(db)
	if err != nil {
		return err
	}

	passSQL := fmt.Sprintf("'%s'", password)
	if currentVersion.LessThan(requiredVersion) {
		passSQL = fmt.Sprintf("PASSWORD(%s)", passSQL)
	}

	sql := fmt.Sprintf("SET PASSWORD FOR '%s'@'%s' = %s",
		d.Get("user").(string),
		d.Get("host").(string),
		passSQL)

	_, err = db.Exec(sql)
	if err != nil {
		return err
	}
	user := fmt.Sprintf("%s@%s",
		d.Get("user").(string),
		d.Get("host").(string))
	d.SetId(user)
	return nil
}

func ReadUserPassword(d *schema.ResourceData, meta interface{}) error {
	// This is obviously not possible.
	return nil
}

func DeleteUserPassword(d *schema.ResourceData, meta interface{}) error {
	// We don't need to do anything on the MySQL side here. Just need TF
	// to remove from the state file.
	return nil
}
