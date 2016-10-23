package common

import (
	checker "gopkg.in/check.v1"
	"strconv"
	"testing"
)

type CacheSuite struct {
	cache *Cache
}

func Test(t *testing.T) { checker.TestingT(t) }

var _ = checker.Suite(&CacheSuite{})

func (s *CacheSuite) SetUpSuite(c *checker.C) {
	s.cache, _ = NewCache()
}

func (s *CacheSuite) TearDownTest(c *checker.C) {
	s.cache.Purge()
}

func (s *CacheSuite) TestAdd(c *checker.C) {
	key := "key"
	val := "val"
	s.cache.Add(key, val)
	obtainedVal, _ := s.cache.Get(key)
	c.Assert(obtainedVal, checker.Equals, val)
}
func (s *CacheSuite) TestContains(c *checker.C) {
	key := "key"
	val := "val"
	s.cache.Add(key, val)
	exist := s.cache.Contains(key)
	c.Assert(exist, checker.Equals, true)
}

func (s *CacheSuite) TestRemove(c *checker.C) {
	key := "key"
	val := "val"
	s.cache.Add(key, val)
	s.cache.Remove(key)
	exist := s.cache.Contains(key)
	c.Assert(exist, checker.Equals, false)
}

func (s *CacheSuite) TestKeys(c *checker.C) {
	keys := make([]interface{}, 10)
	for i := 0; i < 10; i += 1 {
		keys[i] = "key" + strconv.Itoa(i)
	}
	for _, k := range keys {
		s.cache.Add(k, "val")
	}
	obtainedKeys := s.cache.Keys()
	c.Assert(obtainedKeys, checker.HasLen, 10)
}

func (s *CacheSuite) TestAddStruct(c *checker.C) {
	type Person struct {
		name string
		age  int
	}
	value := Person{
		name: "name",
		age:  18,
	}
	s.cache.Add("person1", value)
	ret, _ := s.cache.Get("person1")
	trueVal := ret.(Person)
	c.Assert(trueVal, checker.DeepEquals, value)
}
