package internals

import "github.com/stavrogyn/pokedexcli/internals/pokecache"

// Re-export from pokecache package
type Cache = pokecache.Cache

var NewCache = pokecache.NewCache
