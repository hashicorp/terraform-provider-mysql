package mysql

import (
	"github.com/hashicorp/go-version"
	"testing"
)

func TestParseServerVersionString_MySQL(t *testing.T) {
	mysql := parseServerVersionString("5.7.5")
	if mysql.vendor != MySQL {
		t.Errorf("Simple versions are assumed to be MySQL")
	}
	expected, _ := version.NewVersion("5.7.5")
	if !mysql.version.Equal(expected) {
		t.Errorf("Should not modify the simple version string")
	}
}

func TestParseServerVersionString_MariaDB(t *testing.T) {
	mariadb := parseServerVersionString("10.1.8-MariaDB")
	if mariadb.vendor != MariaDB {
		t.Errorf("MariaDB adds a postfix to their version strings")
	}
	expected, _ := version.NewVersion("10.1.8")
	if !mariadb.version.Equal(expected) {
		t.Errorf("Should get the correct part as version string")
	}
}
