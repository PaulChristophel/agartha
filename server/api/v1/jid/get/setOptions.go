package jid

import "github.com/PaulChristophel/sylvite/server/config"

var (
	table    string
	useJSONB bool
)

func SetOptions(saltTables config.SaltDBTables) {
	table = saltTables.JIDs
	useJSONB = saltTables.UseJSONB
}
