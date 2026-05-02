package component

import (
	"fmt"

	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/aqi/aqicloud-short-link-go/internal/link/sharding"
)

// CreateShortLinkCode generates a short link code compatible with the Java version.
// Algorithm:
//  1. Compute MurmurHash3_32 (Guava-compatible, UTF-16 LE input)
//  2. Base62 encode the hash
//  3. Add DB shard prefix + table shard suffix
//
// The input param is already prefixed with snowflakeId&originalUrl.
func CreateShortLinkCode(param string) string {
	murmurHash := util.MurmurHash3Guava(param)
	code := util.EncodeToBase62(murmurHash)
	dbPrefix := sharding.GetRandomDBPrefix(code)
	tableSuffix := sharding.GetRandomTableSuffix(code)
	return dbPrefix + code + tableSuffix
}

// PrepareUrlForHash prepares the URL with snowflake prefix for hashing.
// If version > 0, uses version number instead of snowflake ID.
func PrepareUrlForHash(originalUrl string, version int64) string {
	if version <= 0 {
		return util.AddUrlPrefix(originalUrl)
	}
	return fmt.Sprintf("%d&%s", version, originalUrl)
}

// IncrementUrlVersion increments the version prefix for collision retry.
func IncrementUrlVersion(url string) string {
	return util.AddUrlPrefixVersion(url)
}
