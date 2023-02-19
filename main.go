package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/nfnt/resize"
	"golang.design/x/clipboard"
)

type oldImage struct {
	oldBased string
}

func main() {
	err := clipboard.Init()
	if err != nil {
		panic(err.Error())
	}

	//Para comparação de base64 para não entrar no loop
	old := &oldImage{}

	//Monitora o evento
	ch := clipboard.Watch(context.TODO(), clipboard.FmtImage)
	// data -> evento da imagem
	for data := range ch {

		//Capturando evento de print
		based := base64.StdEncoding.EncodeToString(data)

		// Comparação feita para não entrar em loop
		// Na linha 92 é salva nova image no clipboard
		if !compareBase64(based, old.oldBased) {

			unbased, err := base64.StdEncoding.DecodeString(based)
			if err != nil {
				panic(err)
			}
			r := bytes.NewReader(unbased)
			img, e := png.Decode(r)
			if e != nil {
				panic(e)
			}

			//Recupera a marca d'água
			watermark, err := os.Open("watermark.png")
			if err != nil {
				panic(err)
			}

			watermarkDecode, err := png.Decode(watermark)
			if err != nil {
				panic(err)
			}

			//Salva em disco a imagem
			filename := fmt.Sprintf("screenshot-%d.png", time.Now().Unix())
			newFile, err := os.Create(filename)
			if err != nil {
				panic(err)
			}

			// Fazendo resize da marca d'água
			// Tem que ser feito pois prints menores com uma tamanho fixo pode ficar cortada
			width := img.Bounds().Size().X / 2
			heigth := img.Bounds().Size().Y / 2
			resizeWaterMark := resize.Resize(uint(width), uint(heigth), watermarkDecode, resize.Lanczos3)

			bounds := img.Bounds()
			screenshotWithWaterMark := image.NewRGBA(bounds)

			//adicionando uma posição da marca d'água
			offset := image.Point{
				width,
				heigth,
			}

			// fazendo merge das images
			draw.Draw(screenshotWithWaterMark, bounds, img, image.ZP, draw.Src)
			draw.Draw(screenshotWithWaterMark, resizeWaterMark.Bounds().Add(offset), resizeWaterMark, image.ZP, draw.Over)
			png.Encode(newFile, screenshotWithWaterMark)

			// fazendo copy pastego
			buf := new(bytes.Buffer)
			png.Encode(buf, screenshotWithWaterMark)

			//Salve em memoria a image e colocada no clipboard
			old.oldBased = base64.StdEncoding.EncodeToString(buf.Bytes())
			clipboard.Write(clipboard.FmtImage, buf.Bytes())
		}

	}

}

// Compara base64
func compareBase64(a string, b string) bool {
	return a == b
}
