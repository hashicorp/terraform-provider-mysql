package mysql

import (
	"database/sql"
	"github.com/hashicorp/go-version"
	"strings"
)

type SQLVendor string

const (
	MySQL   SQLVendor = "MySQL"
	MariaDB SQLVendor = "MariaDB"
)

type ServerVersion struct {
	vendor  SQLVendor
	version *version.Version
}

func serverVersion(db *sql.DB) (ServerVersion, error) {
	versionString, err := serverVersionString(db)
	if err != nil {
		return ServerVersion{}, err
	}

	return parseServerVersionString(versionString), nil
}

func (serverVersion ServerVersion) supportsRoles() bool {
	var requiredVersion *version.Version
	if serverVersion.vendor == MariaDB {
		requiredVersion, _ = version.NewVersion("10.0.5")
	} else {
		requiredVersion, _ = version.NewVersion("8.0.0")
	}

	return serverVersion.version.GreaterThanOrEqual(requiredVersion)
}

// ALTER USER syntax introduced in MySQL 5.7.6 deprecates SET PASSWORD (GH-8230)
// We assume that MariaDB will eventually follow suit
func (serverVersion ServerVersion) deprecatedSetPassword() bool {
	var requiredVersion *version.Version
	if serverVersion.vendor == MariaDB {
		requiredVersion, _ = version.NewVersion("10.2.0")
	} else {
		requiredVersion, _ = version.NewVersion("5.7.6")
	}

	return serverVersion.version.GreaterThanOrEqual(requiredVersion)
}

func (serverVersion ServerVersion) supportsTlsOption() bool {
	var requiredVersion *version.Version
	if serverVersion.vendor == MariaDB {
		// TODO since when exactly has MariaDB had this option?
		// The docs are unclear: https://mariadb.com/kb/en/create-user/#tls-options
		// requiredVersion, _ = version.NewVersion("5.7.0")
		return true
	} else {
		requiredVersion, _ = version.NewVersion("5.7.0")
	}

	return serverVersion.version.GreaterThanOrEqual(requiredVersion)
}

func (serverVersion ServerVersion) requiresExplicitPassword() bool {
	notRequiredVersion, _ := version.NewVersion("8.0.0")
	if serverVersion.vendor == MySQL && serverVersion.version.GreaterThanOrEqual(notRequiredVersion) {
		return false
	}

	return true
}

// MySQL 8 returns more data in a single row when issuing
// a SHOW COLLATION statement
func (serverVersion ServerVersion) extraColumnInShowCollation() bool {
	requiredMySQLVersion, _ := version.NewVersion("8.0.0")
	if serverVersion.vendor == MySQL && serverVersion.version.GreaterThanOrEqual(requiredMySQLVersion) {
		return true
	}

	return false
}

func parseServerVersionString(versionString string) ServerVersion {
	parts := strings.Split(versionString, "-")
	newVersion, _ := version.NewVersion(parts[0])

	if len(parts) == 1 {
		return ServerVersion{
			vendor:  MySQL,
			version: newVersion,
		}
	} else {
		return ServerVersion{
			vendor:  MariaDB,
			version: newVersion,
		}
	}
}

func serverVersionString(db *sql.DB) (string, error) {
	var versionString string
	err := db.QueryRow("SELECT @@GLOBAL.version").Scan(&versionString)
	if err != nil {
		return "", err
	}

	return versionString, nil
}
