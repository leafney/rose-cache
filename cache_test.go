/**
 * @Author:      leafney
 * @GitHub:      https://github.com/leafney
 * @Project:     rose-cache
 * @Date:        2023-05-16 22:45
 * @Description:
 */

package rcache

import (
	"testing"
	"time"
)

// 基础操作测试
func TestNewCache(t *testing.T) {
	// 测试基本创建
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer c.Close()

	// 测试自定义配置
	c2, err := NewCache(1,
		WithLifeWindow(5*time.Minute),
		WithCleanWindow(1*time.Minute),
	)
	if err != nil {
		t.Fatalf("NewCache with options failed: %v", err)
	}
	defer c2.Close()
}

// 基本存取测试
func TestBasicOperations(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer c.Close()

	// 测试 Set/Get
	t.Run("Set and Get", func(t *testing.T) {
		err := c.Set("key1", []byte("value1"))
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		val, err := c.Get("key1")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if string(val) != "value1" {
			t.Errorf("Get returned wrong value: got %s, want value1", string(val))
		}
	})

	// 测试 SetS/GetS
	t.Run("SetS and GetS", func(t *testing.T) {
		err := c.SetS("key2", "value2")
		if err != nil {
			t.Errorf("SetS failed: %v", err)
		}

		val, err := c.GetS("key2")
		if err != nil {
			t.Errorf("GetS failed: %v", err)
		}
		if val != "value2" {
			t.Errorf("GetS returned wrong value: got %s, want value2", val)
		}
	})

	// 测试 Del
	t.Run("Del", func(t *testing.T) {
		err := c.Set("key3", []byte("value3"))
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		err = c.Del("key3")
		if err != nil {
			t.Errorf("Del failed: %v", err)
		}

		_, err = c.Get("key3")
		if err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})

	// 测试 Exists
	t.Run("Exists", func(t *testing.T) {
		err := c.Set("key4", []byte("value4"))
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		if !c.Exists("key4") {
			t.Error("Exists returned false for existing key")
		}

		if c.Exists("nonexistent") {
			t.Error("Exists returned true for non-existent key")
		}
	})
}

// 扩展存取测试
func TestExtendedOperations(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer c.Close()

	// 测试 XSet/XGet
	t.Run("XSet and XGet", func(t *testing.T) {
		err := c.XSet("key1", []byte("value1"))
		if err != nil {
			t.Errorf("XSet failed: %v", err)
		}

		val, err := c.XGet("key1")
		if err != nil {
			t.Errorf("XGet failed: %v", err)
		}
		if string(val) != "value1" {
			t.Errorf("XGet returned wrong value: got %s, want value1", string(val))
		}
	})

	// 测试带过期时间的操作
	t.Run("XSetEx and TTL", func(t *testing.T) {
		err := c.XSetEx("key2", []byte("value2"), 2*time.Second)
		if err != nil {
			t.Errorf("XSetEx failed: %v", err)
		}

		// 检查 TTL
		ttl, err := c.XTTL("key2")
		if err != nil {
			t.Errorf("XTTL failed: %v", err)
		}
		if ttl <= 0 {
			t.Errorf("Expected positive TTL, got %d", ttl)
		}

		// 等待过期
		time.Sleep(3 * time.Second)
		_, err = c.XGet("key2")
		if err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound after expiration, got %v", err)
		}
	})
}

// 计数器操作测试
func TestCounterOperations(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer c.Close()

	// 测试 XIncr
	t.Run("XIncr", func(t *testing.T) {
		val, err := c.XIncr("counter1")
		if err != nil {
			t.Errorf("XIncr failed: %v", err)
		}
		if val != 1 {
			t.Errorf("XIncr first call returned %d, want 1", val)
		}

		val, err = c.XIncr("counter1")
		if err != nil {
			t.Errorf("XIncr failed: %v", err)
		}
		if val != 2 {
			t.Errorf("XIncr second call returned %d, want 2", val)
		}
	})

	// 测试 XIncrBy
	t.Run("XIncrBy", func(t *testing.T) {
		val, err := c.XIncrBy("counter2", 5)
		if err != nil {
			t.Errorf("XIncrBy failed: %v", err)
		}
		if val != 5 {
			t.Errorf("XIncrBy returned %d, want 5", val)
		}
	})

	// 测试 XDecr
	t.Run("XDecr", func(t *testing.T) {
		// 先设置一个值
		err := c.XSet("counter3", []byte("10"))
		if err != nil {
			t.Errorf("XSet failed: %v", err)
		}

		val, err := c.XDecr("counter3")
		if err != nil {
			t.Errorf("XDecr failed: %v", err)
		}
		if val != 9 {
			t.Errorf("XDecr returned %d, want 9", val)
		}
	})

	// 测试 XDecrBy
	t.Run("XDecrBy", func(t *testing.T) {
		// 先设置一个值
		err := c.XSet("counter4", []byte("20"))
		if err != nil {
			t.Errorf("XSet failed: %v", err)
		}

		val, err := c.XDecrBy("counter4", 5)
		if err != nil {
			t.Errorf("XDecrBy failed: %v", err)
		}
		if val != 15 {
			t.Errorf("XDecrBy returned %d, want 15", val)
		}
	})
}

// 过期时间测试
func TestExpirationOperations(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer c.Close()

	// 测试 XExpireAt
	t.Run("XExpireAt", func(t *testing.T) {
		err := c.XSet("key1", []byte("value1"))
		if err != nil {
			t.Errorf("XSet failed: %v", err)
		}

		expireTime := time.Now().Add(2 * time.Second)
		err = c.XExpireAt("key1", expireTime)
		if err != nil {
			t.Errorf("XExpireAt failed: %v", err)
		}

		time.Sleep(3 * time.Second)
		_, err = c.XGet("key1")
		if err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound after expiration, got %v", err)
		}
	})

	// 测试 XExpire
	t.Run("XExpire", func(t *testing.T) {
		err := c.XSet("key2", []byte("value2"))
		if err != nil {
			t.Errorf("XSet failed: %v", err)
		}

		err = c.XExpire("key2", 1*time.Second)
		if err != nil {
			t.Errorf("XExpire failed: %v", err)
		}

		time.Sleep(2 * time.Second)
		_, err = c.XGet("key2")
		if err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound after expiration, got %v", err)
		}
	})
}

// 错误情况测试
func TestErrorCases(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer c.Close()

	// 测试空键
	t.Run("Empty Key", func(t *testing.T) {
		err := c.Set("", []byte("value"))
		if err != ErrKeyEmpty {
			t.Errorf("Expected ErrKeyEmpty, got %v", err)
		}
	})

	// 测试空值
	t.Run("Empty Value", func(t *testing.T) {
		err := c.Set("key", nil)
		if err != ErrValueEmpty {
			t.Errorf("Expected ErrValueEmpty, got %v", err)
		}
	})

	// 测试不存在的键
	t.Run("Non-existent Key", func(t *testing.T) {
		_, err := c.Get("nonexistent")
		if err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})

	// 测试数值操作错误
	t.Run("Invalid Number Operation", func(t *testing.T) {
		err := c.XSet("key", []byte("not a number"))
		if err != nil {
			t.Errorf("XSet failed: %v", err)
		}

		_, err = c.XIncr("key")
		if err == nil || err.Error() != "value is not an integer" {
			t.Errorf("Expected 'value is not an integer' error, got %v", err)
		}
	})
}
