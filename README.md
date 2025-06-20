# lzss

[![GoDoc](https://godoc.org/github.com/blacktop/lzss?status.svg)](https://pkg.go.dev/github.com/blacktop/lzss) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> LZSS compression package for Go.

---

## Install

### As a Go package

```bash
go get github.com/blacktop/lzss
```

### As a CLI tool

```bash
go install github.com/blacktop/lzss/cmd@latest
```

## CLI Usage

```bash
# Compress a file
lzss -c input.txt                    # creates input.txt.lzss

# Decompress a file  
lzss -d input.txt.lzss               # creates input.txt

# Custom output file
lzss -c -o compressed.lzss input.txt
lzss -d -o output.txt input.lzss

# Show help
lzss -h
```

## Examples

### Compression

```golang
import (
    "io/ioutil"
    "log"

    "github.com/blacktop/lzss"
    "github.com/pkg/errors"
)

func main() {
    // Compress data
    data, err := ioutil.ReadFile("input.txt")
    if err != nil {
        log.Fatal(errors.Wrap(err, "failed to read input file"))
    }

    compressed := lzss.Compress(data)
    err = ioutil.WriteFile("compressed.bin", compressed, 0644)
    if err != nil {
        log.Fatal(errors.Wrap(err, "failed to write compressed file"))
    }
}
```

### Decompression

```golang
import (
    "io/ioutil"
    "log"

    "github.com/blacktop/lzss"
    "github.com/pkg/errors"
)

func main() {
    dat, err := ioutil.ReadFile("compressed.bin")
    if err != nil {
        log.Fatal(errors.Wrap(err, "failed to read compressed file"))
    }

    decompressed := lzss.Decompress(dat)
    err = ioutil.WriteFile("compressed.bin.decompressed", decompressed, 0644)
    if err != nil {
        log.Fatal(errors.Wrap(err, "failed to decompress file"))
    }
}
```

> **NOTE:** I believe lzss expects the data to be word aligned.

## Credit

Converted to Golang from `https://github.com/opensource-apple/kext_tools/blob/master/compression.c`

## License

MIT Copyright (c) 2018-2025 blacktop
