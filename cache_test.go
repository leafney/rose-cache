/**
 * @Author:      leafney
 * @GitHub:      https://github.com/leafney
 * @Project:     rose-cache
 * @Date:        2023-05-16 22:45
 * @Description:
 */

package rcache

import (
	"context"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		c, err := NewCache(1)
		if err != nil {
			t.Fatalf("NewCache error: %v", err)
		}
		defer c.Close()
	})

	t.Run("with options", func(t *testing.T) {
		c, err := NewCache(1,
			WithContext(context.Background()),
			WithLifeWindow(time.Minute),
			WithCleanWindow(2*time.Minute),
		)
		if err != nil {
			t.Fatalf("NewCache with options error: %v", err)
		}
		defer c.Close()
	})
}

func TestCache_BasicOperations(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache error: %v", err)
	}
	defer c.Close()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test_key"
		value := []byte("test_value")

		if err := c.Set(key, value); err != nil {
			t.Errorf("Set error: %v", err)
		}

		got, err := c.Get(key)
		if err != nil {
			t.Errorf("Get error: %v", err)
		}

		if string(got) != string(value) {
			t.Errorf("Get = %v, want %v", string(got), string(value))
		}
	})

	t.Run("SetString and GetString", func(t *testing.T) {
		key := "string_key"
		value := "string_value"

		if err := c.SetString(key, value); err != nil {
			t.Errorf("SetString error: %v", err)
		}

		got, err := c.GetString(key)
		if err != nil {
			t.Errorf("GetString error: %v", err)
		}

		if got != value {
			t.Errorf("GetString = %v, want %v", got, value)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete_key"
		value := "delete_value"

		if err := c.SetString(key, value); err != nil {
			t.Errorf("SetString error: %v", err)
		}

		if err := c.Delete(key); err != nil {
			t.Errorf("Delete error: %v", err)
		}

		if c.Has(key) {
			t.Error("key should not exist after deletion")
		}
	})

	t.Run("Has", func(t *testing.T) {
		key := "has_key"
		value := "has_value"

		if c.Has(key) {
			t.Error("key should not exist before setting")
		}

		if err := c.SetString(key, value); err != nil {
			t.Errorf("SetString error: %v", err)
		}

		if !c.Has(key) {
			t.Error("key should exist after setting")
		}
	})
}

func TestCache_SetEX(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache error: %v", err)
	}
	defer c.Close()

	t.Run("SetEX with expiration", func(t *testing.T) {
		key := "ex_key"
		value := []byte("ex_value")
		expiration := 2 * time.Second

		if err := c.SetEX(key, value, expiration); err != nil {
			t.Errorf("SetEX error: %v", err)
		}

		// Value should exist immediately
		got, err := c.Get(key)
		if err != nil {
			t.Errorf("Get error: %v", err)
		}
		if string(got) != string(value) {
			t.Errorf("Get = %v, want %v", string(got), string(value))
		}

		// Wait for expiration
		time.Sleep(expiration + time.Second)

		// Value should not exist after expiration
		if _, err := c.Get(key); err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})

	t.Run("SetEXString with expiration", func(t *testing.T) {
		key := "ex_string_key"
		value := "ex_string_value"
		expiration := 2 * time.Second

		if err := c.SetEXString(key, value, expiration); err != nil {
			t.Errorf("SetEXString error: %v", err)
		}

		// Value should exist immediately
		got, err := c.GetString(key)
		if err != nil {
			t.Errorf("GetString error: %v", err)
		}
		if got != value {
			t.Errorf("GetString = %v, want %v", got, value)
		}

		// Wait for expiration
		time.Sleep(expiration + time.Second)

		// Value should not exist after expiration
		if _, err := c.GetString(key); err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestCache_ErrorCases(t *testing.T) {
	c, err := NewCache(1)
	if err != nil {
		t.Fatalf("NewCache error: %v", err)
	}
	defer c.Close()

	t.Run("empty key", func(t *testing.T) {
		if err := c.Set("", []byte("value")); err != ErrKeyEmpty {
			t.Errorf("Expected ErrKeyEmpty, got %v", err)
		}

		if err := c.SetEX("", []byte("value"), time.Minute); err != ErrKeyEmpty {
			t.Errorf("Expected ErrKeyEmpty, got %v", err)
		}

		if _, err := c.Get(""); err != ErrKeyEmpty {
			t.Errorf("Expected ErrKeyEmpty, got %v", err)
		}
	})

	t.Run("empty value", func(t *testing.T) {
		if err := c.Set("key", nil); err != ErrValueEmpty {
			t.Errorf("Expected ErrValueEmpty, got %v", err)
		}

		if err := c.SetEX("key", nil, time.Minute); err != ErrValueEmpty {
			t.Errorf("Expected ErrValueEmpty, got %v", err)
		}
	})

	t.Run("get non-existent key", func(t *testing.T) {
		if _, err := c.Get("non_existent"); err != ErrKeyNotFound {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestNewCacheBase(t *testing.T) {
	c, err := NewCache(1)
	//c, err := NewCache(10, WithContext(context.Background()))
	if err != nil {
		t.Error(err)
		return
	}

	defer c.Close()

	c.SetString("aaa", "hello")

	time.Sleep(30 * time.Second)

	v1, err := c.GetString("aaa")
	if err != nil {
		t.Error(err)
	}
	t.Log(v1)

	time.Sleep(31 * time.Second)
	if c.Has("aaa") {
		t.Log("found")
	} else {
		t.Log("not found")
	}
}
