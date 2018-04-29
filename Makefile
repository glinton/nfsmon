default: test

test: 
	@go test -coverprofile=cover.prof -race

view: 
	@go tool cover -html=cover.prof

.PHONY: test view
