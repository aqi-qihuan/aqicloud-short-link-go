package util

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSnowflakeID_NonZero(t *testing.T) {
	id := GenerateSnowflakeID()
	assert.NotZero(t, id, "雪花 ID 不应为 0")
}

func TestGenerateSnowflakeID_Unique(t *testing.T) {
	// 生成 1000 个 ID 应全部唯一
	seen := make(map[uint64]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := GenerateSnowflakeID()
		assert.NotZero(t, id)
		assert.False(t, seen[id], "雪花 ID %d 重复", id)
		seen[id] = true
	}
}

func TestGenerateSnowflakeID_MonotonicIncreasing(t *testing.T) {
	// 单线程内应严格单调递增
	prev := GenerateSnowflakeID()
	for i := 0; i < 100; i++ {
		cur := GenerateSnowflakeID()
		assert.Greater(t, cur, prev, "雪花 ID 应单调递增")
		prev = cur
	}
}

func TestGenerateSnowflakeID_ConcurrentUniqueness(t *testing.T) {
	// 多线程并发生成应无重复
	const goroutines = 10
	const perGoroutine = 100
	total := goroutines * perGoroutine

	ids := make(chan uint64, total)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				ids <- GenerateSnowflakeID()
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[uint64]bool, total)
	for id := range ids {
		assert.NotZero(t, id)
		assert.False(t, seen[id], "并发雪花 ID %d 重复", id)
		seen[id] = true
	}
	assert.Equal(t, total, len(seen))
}

func TestGenerateSnowflakeIDStr(t *testing.T) {
	idStr := GenerateSnowflakeIDStr()
	assert.NotEmpty(t, idStr)
	assert.NotEqual(t, "0", idStr, "雪花 ID 字符串不应为 '0'")

	// 验证是纯数字
	for _, c := range idStr {
		assert.True(t, c >= '0' && c <= '9',
			"雪花 ID 字符串应为纯数字，实际: %c", c)
	}
}
