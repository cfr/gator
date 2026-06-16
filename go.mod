module github.com/cfr/gator

go 1.26.3

require internal/config v1.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
)

replace internal/config => ./internal/config
