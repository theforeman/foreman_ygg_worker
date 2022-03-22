package main

import "testing"

func TestGetSetGetRemove(t *testing.T) {
	s := newJobStorage()

	_, prs := s.Get("key")
	if prs {
		t.Errorf("Found key 'key' which should not be there")
	}

	s.Set("key", 1)
	v, prs := s.Get("key")
	if !prs {
		t.Errorf("Could not find key 'key' which should be there")
	}
	if v != 1 {
		t.Errorf("Retrieved '%v', expected '%v'", v, 1)
	}

	s.Remove("key")
	_, prs = s.Get("key")
	if prs {
		t.Errorf("Found key 'key' which should not be there")
	}
}

func TestRemoveMissingIsNoop(t *testing.T) {
	s := newJobStorage()

	_, prs := s.Get("key")
	if prs {
		t.Errorf("Found key 'key' which should not be there")
	}

	s.Remove("key")
	_, prs = s.Get("key")
	if prs {
		t.Errorf("Found key 'key' which should not be there")
	}
}
