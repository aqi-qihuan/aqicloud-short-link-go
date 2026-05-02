package sharding

import "github.com/aqi/aqicloud-short-link-go/internal/common/util"

// TableSuffixList contains the active table shard suffixes.
// Matches Java's ShardingTableConfig.tableSuffixList: ["0", "a"]
var TableSuffixList = []string{"0", "a"}

// GetRandomTableSuffix returns a deterministic table suffix based on the base62 code.
// Uses Java's String.hashCode() algorithm for compatibility.
func GetRandomTableSuffix(code string) string {
	h := util.JavaStringHashCode(code)
	if h < 0 {
		h = -h
	}
	idx := int(h) % len(TableSuffixList)
	return TableSuffixList[idx]
}
