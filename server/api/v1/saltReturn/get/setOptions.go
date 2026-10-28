package saltReturn

import "github.com/PaulChristophel/sylvite/server/config"

var (
	table    string
	useJSONB bool
)

func SetOptions(saltTables config.SaltDBTables) {
	table = saltTables.SaltReturns
	useJSONB = saltTables.UseJSONB
}
