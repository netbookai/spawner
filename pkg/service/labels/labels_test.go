package labels

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestScopeTag(t *testing.T) {
//	config.Set(config.Config{Env: "dev"})
//
//	got := ScopeTag()
//	expected := "nb-dev"
//
//	assert.Equal(t, expected, got, "ScopeTag:")
//}
//
//func TestDefaultTag(t *testing.T) {
//	config.Set(config.Config{Env: "dev"})
//
//	got := DefaultTags()
//
//	scope := "nb-dev"
//	expected := map[string]*string{
//		constants.Scope:        &scope,
//		constants.CreatorLabel: &constants.SpawnerServiceLabel,
//	}
//	assert.Equalf(t, expected, got, "DefaultTags: ")
//
//}

func TestMergeRequestLabel(t *testing.T) {
	tags := make(map[string]*string)
	a := "hello"
	tags["a"] = &a
	labels := map[string]string{"a": "hi", "b": "hello"}
	MergeRequestLabel(tags, labels)
	assert.Equal(t, "hi", *tags["a"], "expected hi, got %s", *tags["a"])
	assert.Equal(t, 2, len(tags), "length must be 2, got %d", len(tags))

}
