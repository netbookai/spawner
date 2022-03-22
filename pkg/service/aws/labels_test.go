package aws

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
