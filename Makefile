GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOFMT = $(GOCMD) fmt

TARGET = distrox

build:
	$(GOBUILD) -o $(TARGET) -v

clean:
	$(GOCLEAN)
	rm -f $(TARGET)

format:
	$(GOFMT) ./...
