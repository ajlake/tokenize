package main

import (
	"fmt"
	"github.com/gammazero/deque"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func readImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch {
	case strings.HasSuffix(path, ".png"):
		return png.Decode(f)
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return jpeg.Decode(f)
	}
	return nil, fmt.Errorf("unsupported image type: %s", path)
}

func writeImage(sourcePath, variant string, img image.Image) error {
	dirname := filepath.Dir(sourcePath)
	basename := filepath.Base(sourcePath)
	lastIndex := strings.LastIndex(basename, ".")
	root := basename
	if lastIndex != -1 {
		root = basename[0:lastIndex]
	}
	destPath := filepath.Join(dirname, fmt.Sprintf("%s_%s.png", root, variant))

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func hasTransparency(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a < 0xffff
}

func areaOf(inputImage image.Image, isContour func(color.Color) bool, startPoints ...image.Point) []image.Point {
	var results []image.Point

	maxX := inputImage.Bounds().Dx()
	maxY := inputImage.Bounds().Dy()
	seen := make([]bool, maxX*maxY)

	queue := deque.New()

	for _, startPoint := range startPoints {
		queue.PushBack(startPoint)
		seen[startPoint.X+startPoint.Y*inputImage.Bounds().Dy()] = true
	}

	for queue.Len() > 0 {
		curr := queue.PopFront().(image.Point)

		currColor := inputImage.At(curr.X, curr.Y)
		if !isContour(currColor) {
			continue
		}
		results = append(results, curr)

		above := curr.Add(image.Pt(0, -1))
		aboveIdx := above.X + above.Y*inputImage.Bounds().Dy()
		if above.X >= 0 && above.X < maxX && above.Y >= 0 && above.Y < maxY && !seen[aboveIdx] {
			queue.PushBack(above)
			seen[aboveIdx] = true
		}

		below := curr.Add(image.Pt(0, 1))
		belowIdx := below.X + below.Y*inputImage.Bounds().Dy()
		if below.X >= 0 && below.X < maxX && below.Y >= 0 && below.Y < maxY && !seen[belowIdx] {
			queue.PushBack(below)
			seen[belowIdx] = true
		}

		left := curr.Add(image.Pt(-1, 0))
		leftIdx := left.X + left.Y*inputImage.Bounds().Dy()
		if left.X >= 0 && left.X < maxX && left.Y >= 0 && left.Y < maxY && !seen[leftIdx] {
			queue.PushBack(left)
			seen[leftIdx] = true
		}

		right := curr.Add(image.Pt(1, 0))
		rightIdx := right.X + right.Y*inputImage.Bounds().Dy()
		if right.X >= 0 && right.X < maxX && right.Y >= 0 && right.Y < maxY && !seen[rightIdx] {
			queue.PushBack(right)
			seen[rightIdx] = true
		}
	}

	return results
}

func makeIcons(borders map[string]image.Image, inputImagePaths []string) error {
	borderAreas := make(map[string][]image.Point)
	for name, border := range borders {
		borderAreas[name] = areaOf(
			border,
			hasTransparency,
			image.Pt(border.Bounds().Max.X/2, border.Bounds().Max.Y/2),
		)
	}

	for _, inputImagePath := range inputImagePaths {
		inputImage, err := readImage(inputImagePath)
		if err != nil {
			return err
		}

		for name, border := range borders {
			borderBounds := border.Bounds()
			scaledImage := image.NewRGBA(image.Rect(0, 0, borderBounds.Dx(), borderBounds.Dy()))
			draw.CatmullRom.Scale(scaledImage, scaledImage.Bounds(), inputImage, inputImage.Bounds(), draw.Over, nil)

			newImage := image.NewRGBA(image.Rect(0, 0, borderBounds.Dx(), borderBounds.Dy()))
			for _, pt := range borderAreas[name] {
				newImage.SetRGBA(pt.X, pt.Y, scaledImage.RGBAAt(pt.X, pt.Y))
			}
			draw.Draw(newImage, newImage.Bounds(), border, border.Bounds().Min, draw.Over)

			if err := writeImage(inputImagePath, name, newImage); err != nil {
				return err
			}
		}

	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "Must specify input image files to tokenize.")
		os.Exit(1)
	}

	borders := readIconBorders()
	if err := makeIcons(borders, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
