package mysql

import (
	"fmt"
	"github.com/satori/go.uuid"

	"github.com/hashicorp/terraform/helper/encryption"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUserPassword() *schema.Resource {
	return &schema.Resource{
		Create: SetUserPassword,
		Read:   ReadUserPassword,
		Delete: DeleteUserPassword,

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
	db := meta.(*providerConfiguration).DB

	password := uuid.NewV4().String()
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

	stmtSQL := fmt.Sprintf("SET PASSWORD FOR '%s'@'%s' = '%s'",
		d.Get("user").(string),
		d.Get("host").(string),
		password)

	_, err = db.Exec(stmtSQL)
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
