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

func TestNewCache(t *testing.T) {
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
