package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const varTemplate = `// Generated Code: DO NOT EDIT

package main

import (
	"bytes"
	"encoding/hex"
	"image"
	"image/png"
)

var iconMap = map[string]string{
{{- range $icon := . }}
    "{{ $icon.Name }}": "{{ $icon.Hex }}",
{{- end }}
}

func readIconBorders() map[string]image.Image{
	result := make(map[string]image.Image)
	for name, icon := range iconMap {
		decodedBytes, err := hex.DecodeString(icon)
		if err != nil {
			panic(err)
		}
		
		img, err := png.Decode(bytes.NewReader(decodedBytes))
		if err != nil {
			panic(err)
		}
		result[name] = img
	}
	return result
}
`

type Icon struct {
	Name string
	Hex  string
}

func embedImages(dir string) error {
	var tmplInput []Icon
	tmpl, err := template.New("icons").Parse(varTemplate)
	if err != nil {
		return err
	}

	icons, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, icon := range icons {
		path := filepath.Join(dir, icon.Name())
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		tmplInput = append(tmplInput, Icon{
			Name: strings.TrimSuffix(icon.Name(), ".png"),
			Hex:  hex.EncodeToString(bytes),
		})
	}

	output, err := os.Create("icons_generated.go")
	if err != nil {
		return err
	}
	defer output.Close()

	return tmpl.Execute(output, tmplInput)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "Must provide input directory containing borders.")
		os.Exit(1)
	}

	if err := embedImages(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
