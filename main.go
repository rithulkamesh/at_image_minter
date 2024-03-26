package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func generateImage(ticketID string, vip bool) (*bytes.Buffer, error) {
	baseFile, err := ioutil.ReadFile("ticket_template.jpg")
	if err != nil {
		return nil, err
	}

	baseJpg, err := jpeg.Decode(bytes.NewReader(baseFile))
	if err != nil {
		return nil, err
	}

	qrcodeBytes, err := qrcode.Encode("https://at-ticket-scan.vercel.app/api/qr/"+ticketID, qrcode.Medium, 300)
	if err != nil {
		return nil, err
	}

	qrcodeImg, _, err := image.Decode(bytes.NewReader(qrcodeBytes))
	if err != nil {
		return nil, err
	}

	bounds := baseJpg.Bounds()
	overlay := image.NewRGBA(bounds)
	draw.Draw(overlay, bounds, baseJpg, image.ZP, draw.Src)

	draw.Draw(overlay, bounds, qrcodeImg, image.Pt(-1550, -200), draw.Over)

	face := basicfont.Face7x13 // Use a larger font face
	face.Width = 16
	drawer := &font.Drawer{
		Dst:  overlay,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
		Dot:  fixed.Point26_6{X: 1623 * 64, Y: 70 * 64}, // Adjust the coordinates as needed
	}

	if vip {
		drawer.DrawString(ticketID + " (VIP)")

	} else {
		drawer.DrawString(ticketID + " (GEN)")

	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, overlay, nil)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func main() {
	e := echo.New()

	e.GET("/ticket/vip/:ticketID", serveTicketImageVIP)
	e.GET("/ticket/gen/:ticketID", serveTicketImage)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func serveTicketImageVIP(c echo.Context) error {
	ticketID := c.Param("ticketID")

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+ticketID+".jpg")
	c.Response().Header().Set("Content-Type", "image/jpeg")

	buf, err := generateImage(ticketID, true)
	if err != nil {
		return err
	}

	return c.Blob(http.StatusOK, "image/jpeg", buf.Bytes())
}

func serveTicketImage(c echo.Context) error {
	ticketID := c.Param("ticketID")

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+ticketID+".jpg")
	c.Response().Header().Set("Content-Type", "image/jpeg")

	buf, err := generateImage(ticketID, false)
	if err != nil {
		return err
	}

	return c.Blob(http.StatusOK, "image/jpeg", buf.Bytes())
}
