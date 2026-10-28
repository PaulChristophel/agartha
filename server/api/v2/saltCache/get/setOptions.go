package saltCache

import "github.com/PaulChristophel/sylvite/server/config"

var table string

func SetOptions(saltTables config.SaltDBTables) {
	table = saltTables.SaltCache
}
