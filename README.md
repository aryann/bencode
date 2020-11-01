# bencode

Package `bencode` implements encoding and decoding of the
[Bencode](https://en.wikipedia.org/wiki/Bencode) serialization format. Bencode
is used by the peer-to-peer file sharing system BitTorrent.

[![Build Status](https://travis-ci.org/aryann/bencode.svg?branch=main)](https://travis-ci.org/aryann/bencode)

## Installation

To install this package, run:

```shell
$ go get github.com/aryann/bencode
```

## Usage

The following is an example that shows how to encode and decode values using
this library. This example is also available at
[the Go Playground](https://play.golang.org/p/QwEanuzOqYt).

```Go
package main

import (
	"fmt"
	"log"

	"github.com/aryann/bencode"
)

func main() {
	type MyData struct {
		MyString   string `bencode:"my-string"`
		MyIntegers []int64  `bencode:"my-integers"`
	}

	myData := MyData{
		MyString:   "Hello, world!",
		MyIntegers: []int64{1, 22, 333},
	}
	encoded, err := bencode.Marshal(myData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Encoded data: %s\n", string(encoded))

	var myDecodedData MyData
	if err := bencode.Unmarshal(encoded, &myDecodedData); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded data: %+v\n", myDecodedData)
}
```

Running this example produces the following output:

```
Encoded data: d11:my-integersli1ei22ei333ee9:my-string13:Hello, world!e
Decoded data: {MyString:Hello, world! MyIntegers:[1 22 333]}
```
