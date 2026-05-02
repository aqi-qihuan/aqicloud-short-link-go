package sharding

// RouteShortLink returns the DB prefix and table suffix for a given short link code.
// DB shard: first character of code.
// Table shard: last character of code.
func RouteShortLink(code string) (dbPrefix string, tableSuffix string) {
	if len(code) == 0 {
		return "0", "0"
	}
	dbPrefix = string(code[0])
	tableSuffix = string(code[len(code)-1])
	return
}

// RouteGroupCodeMapping returns the DB index and table index for group_code_mapping.
// DB shard: account_no % 2 (0 or 1)
// Table shard: group_id % 2 (0 or 1)
func RouteGroupCodeMapping(accountNo int64, groupId int64) (dbIndex int, tableIndex int) {
	dbIndex = int(accountNo%2) & 0x7FFFFFFF // ensure non-negative
	tableIndex = int(groupId%2) & 0x7FFFFFFF
	return
}

// RouteLinkGroup returns the DB index for link_group.
// DB shard: account_no % 2
func RouteLinkGroup(accountNo int64) int {
	return int(accountNo%2) & 0x7FFFFFFF
}

// GetTableName returns the physical table name by appending the suffix.
func GetTableName(logicName string, suffix string) string {
	return logicName + "_" + suffix
}
