package sharding

import "github.com/aqi/aqicloud-short-link-go/internal/common/util"

// DBPrefixList contains the active database shard prefixes.
// Matches Java's ShardingDBConfig.dbPrefixList: ["0", "1", "a"]
var DBPrefixList = []string{"0", "1", "a"}

// DBDatasourceNames are the corresponding datasource names.
var DBDatasourceNames = []string{"aqicloud_link_0", "aqicloud_link_1", "aqicloud_link_a"}

// GetRandomDBPrefix returns a deterministic DB prefix based on the base62 code.
// Uses Java's String.hashCode() algorithm for compatibility.
func GetRandomDBPrefix(code string) string {
	h := util.JavaStringHashCode(code)
	if h < 0 {
		h = -h
	}
	idx := int(h) % len(DBPrefixList)
	return DBPrefixList[idx]
}

// GetDBNameByPrefix returns the datasource name for a given prefix.
func GetDBNameByPrefix(prefix string) string {
	for i, p := range DBPrefixList {
		if p == prefix {
			return DBDatasourceNames[i]
		}
	}
	return DBDatasourceNames[0]
}

// GetDBIndexByPrefix returns the datasource index (0, 1, 2) for a given prefix.
func GetDBIndexByPrefix(prefix string) int {
	for i, p := range DBPrefixList {
		if p == prefix {
			return i
		}
	}
	return 0
}
