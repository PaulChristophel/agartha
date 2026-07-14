package saltKeys

import "github.com/PaulChristophel/agartha/server/config"

var table string

func SetOptions(saltTables config.SaltDBTables) {
	table = saltTables.SaltKeys
	if table == "" {
		table = "salt_keys"
	}
}
