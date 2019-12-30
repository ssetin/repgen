example:
	./generator example example/sql example/mock
	goimports -w example/sql/example.gen.go
	goimports -w example/mock/example.gen.go
	goimports -w example/example.gen.go

build:
	go build -o generator

.PHONY: example