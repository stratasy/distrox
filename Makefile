GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOFMT = $(GOCMD) fmt

TARGET = distrox

.PHONY: test

build:
	$(GOBUILD) -o $(TARGET) -v

check:
	$(GOBUILD) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(TARGET)

format:
	$(GOFMT) ./...
