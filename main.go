package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	e := echo.New()

	e.GET("/ticket/:ticketID", serveTicketImage)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func serveTicketImage(c echo.Context) error {
	ticketID := c.Param("ticketID")

	baseFile, err := ioutil.ReadFile("ticket_template.jpg")
	if err != nil {
		return err
	}

	baseJpg, err := jpeg.Decode(bytes.NewReader(baseFile))
	if err != nil {
		return err
	}

	qrcodeBytes, err := qrcode.Encode(ticketID, qrcode.Medium, 300)
	if err != nil {
		return err
	}

	qrcodeImg, _, err := image.Decode(bytes.NewReader(qrcodeBytes))
	if err != nil {
		return err
	}

	bounds := baseJpg.Bounds()
	overlay := image.NewRGBA(bounds)

	draw.Draw(overlay, bounds, baseJpg, image.ZP, draw.Src)
	draw.Draw(overlay, bounds, qrcodeImg, image.Pt(-1550, -200), draw.Over)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, overlay, nil)
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+ticketID+".jpg")
	c.Response().Header().Set("Content-Type", "image/jpeg")

	return c.Blob(http.StatusOK, "image/jpeg", buf.Bytes())
}
