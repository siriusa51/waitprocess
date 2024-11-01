package waitprocess

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func Test_orderMap(t *testing.T) {
	m := newOrderMap[string, string]()
	getted := m.get("key")
	assert.Equal(t, "", getted)

	exist, getted := m.load("key")
	assert.False(t, exist)
	assert.Equal(t, "", getted)

	exist = m.contains("key")
	assert.False(t, exist)

	ok := m.delete("key")
	assert.False(t, ok)

	size := m.size()
	assert.Equal(t, 0, size)

	// setted
	m.set("key", "value")
	getted = m.get("key")
	assert.Equal(t, "value", getted)

	exist, getted = m.load("key")
	assert.True(t, exist)
	assert.Equal(t, "value", getted)

	exist = m.contains("key")
	assert.True(t, exist)

	size = m.size()
	assert.Equal(t, 1, size)

	// deleted
	ok = m.delete("key")
	assert.True(t, ok)

	getted = m.get("key")
	assert.Equal(t, "", getted)

	exist, getted = m.load("key")
	assert.False(t, exist)
	assert.Equal(t, "", getted)

	exist = m.contains("key")
	assert.False(t, exist)

	ok = m.delete("key")
	assert.False(t, ok)

	size = m.size()
	assert.Equal(t, 0, size)
}

func Test_orderMap_set(t *testing.T) {
	// set same key
	m := newOrderMap[string, string]()
	m.set("key1", "value1")
	m.set("key2", "value2")
	m.set("key3", "value3")

	m.set("key1", "key1")
	getted := m.get("key1")
	assert.Equal(t, "key1", getted)

	m.rangeFunc(func(index int, key string, value string) bool {
		if key == "key1" {
			assert.Equal(t, "key1", value)
		}
		return true
	})
}

func Test_orderMap_rangeFunc(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			datas := rand.Perm(100)
			m := newOrderMap[int, int]()
			for _, v := range datas {
				m.set(v, v)
			}

			m.rangeFunc(func(index int, key int, value int) bool {
				assert.Equal(t, datas[index], key)
				return true
			})
		}
	})

	t.Run("break", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		m := newOrderMap[int, int]()

		for _, v := range data {
			m.set(v, v)
		}

		count := 0
		m.rangeFunc(func(index int, key int, value int) bool {
			if key == 3 {
				return false
			}
			count++
			return true
		})

		assert.Equal(t, 2, count)
	})
}
