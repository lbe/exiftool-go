// Command regenjson reads exiftool -json stdout from stdin and writes normalized JSON to a file.
// Used by scripts/exiftoolgen json-samsung.
package main

import (
	"io"
	"log"
	"os"

	"github.com/lbe/exiftool-go/internal/golden"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: regenjson <output-path>")
	}
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	out, err := golden.NormalizeExiftoolJSONGolden(in)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(os.Args[1], out, 0o644); err != nil {
		log.Fatal(err)
	}
}
