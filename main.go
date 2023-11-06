package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/ubavic/bas-celik/document"
	"github.com/ubavic/bas-celik/gui"
)

//go:embed assets/free-sans-regular.ttf
var font embed.FS

//go:embed assets/rfzo.png
var rfzoLogo embed.FS

func main() {
	flag.Parse()

	err := document.SetData(font, rfzoLogo)
	if err != nil {
		fmt.Println("Error establishing context:", err)
		return
	}
	gui.StartGui()
}
