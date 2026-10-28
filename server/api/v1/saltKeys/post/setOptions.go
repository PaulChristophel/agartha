package saltKeys

import "github.com/PaulChristophel/sylvite/server/config"

var table string

// SetOptions configures the database table used by salt_keys write handlers.
func SetOptions(saltTables config.SaltDBTables) {
	table = saltTables.SaltKeys
	if table == "" {
		table = "salt_keys"
	}
}
