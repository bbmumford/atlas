module github.com/bbmumford/atlas

go 1.25.0

require (
	github.com/ORBTR/aether v0.0.117
	github.com/bbmumford/loom v0.0.0
)

require golang.org/x/sync v0.22.0 // indirect

replace github.com/ORBTR/aether => ../aether

replace github.com/bbmumford/loom => ../loom
