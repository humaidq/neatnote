GC = go
BINDATA = go-bindata

TARGET = neatnote

all: clean bindata format test $(TARGET)
dev: clean bindata-dev format test $(TARGET)

$(TARGET): main.go
	$(GC) build

bindata:
	$(BINDATA) -prefix templates -o templates/bindata.go -pkg templates templates/...
	$(BINDATA) -prefix public -o public/bindata.go -pkg public public/...

bindata-dev:
	$(BINDATA) -debug -prefix templates -o templates/bindata.go -pkg templates templates/...
	$(BINDATA) -debug -prefix public -o public/bindata.go -pkg public public/...

format:
	$(GC) fmt ./...

test:
	$(GC) test ./...
	$(GC) vet ./...

clean:
	$(RM) $(TARGET)

