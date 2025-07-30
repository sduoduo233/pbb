package controllers

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	//go:embed NotoSansSC.ttf
	NOTOSANSSC []byte

	BLACK   = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	RED     = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	WHITE   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	NO_DATA = hexColor("#fffee8ff")
	SMOKE   = hexColor("#dddddd")

	FONT16 font.Face
	FONT20 font.Face
	FONT12 font.Face

	drawTextLock = sync.Mutex{}
)

func drawFilledRect(img draw.Image, x, y, w, h int, c color.RGBA) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), image.NewUniform(c), image.Point{X: 0, Y: 0}, draw.Src)
}

func drawFilledRectMask(img draw.Image, x, y, w, h int, c color.RGBA, mask image.Rectangle) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h).Intersect(mask), image.NewUniform(c), image.Point{X: 0, Y: 0}, draw.Src)
}

func drawText(img draw.Image, str string, face font.Face, x, y int, c color.RGBA) {
	drawTextLock.Lock()
	defer drawTextLock.Unlock()

	a := face.Metrics().Ascent.Ceil()

	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, y+a),
	}
	d.DrawString(str)
}

func measureWidth(face font.Face, str string) int {
	drawTextLock.Lock()
	defer drawTextLock.Unlock()

	w := 0

	for _, r := range str {
		a, ok := face.GlyphAdvance(r)
		if !ok {
			continue
		}
		w += a.Ceil()
	}

	return w
}

func hexColor(str string) color.RGBA {
	str = strings.TrimPrefix(str, "#")
	b, err := hex.DecodeString(str)
	if err != nil || len(b) < 3 {
		panic(fmt.Errorf("bad hex color: %s", str))
	}
	return color.RGBA{A: 255, R: b[0], G: b[1], B: b[2]}
}

func graph(c echo.Context) error {
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	reqStart := time.Now()
	reqTimeout := func() {
		if time.Since(reqStart) > time.Second*5 {
			panic("graph timeout")
		}
	}

	timestampEnd, err1 := strconv.Atoi(c.QueryParam("end"))
	timestampStart, err2 := strconv.Atoi(c.QueryParam("start"))
	if err1 != nil || err2 != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	timestampStart = timestampStart / 300 * 300
	timestampEnd = timestampEnd / 300 * 300

	user := GetUser(c)
	showHidden := user != nil

	timeZoneOffset, err3 := strconv.Atoi(c.QueryParam("tz"))
	if err3 != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	timeZoneOffset *= 60 // convert minutes to seconds

	// fetch data

	var server db.Server
	err := db.DB.Get(&server, "SELECT * FROM servers WHERE id = ?", from)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "404 Not Found")
		}
		return fmt.Errorf("db: %w", err)
	}

	if !showHidden && server.Hidden {
		return c.Render(http.StatusNotFound, "error", D{"error": "404 Not Found"})
	}

	var service db.Service
	err = db.DB.Get(&service, "SELECT * FROM services WHERE id = ?", to)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "404 Not Found")
		}
		return fmt.Errorf("db: %w", err)
	}

	var metrics []db.ServiceMetric
	err = db.DB.Select(&metrics, "SELECT * FROM service_metrics WHERE `from` = ? AND `to` = ? AND timestamp <= ? AND timestamp >= ? ORDER BY timestamp", from, to, timestampEnd, timestampStart)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	// render

	img := image.NewRGBA(image.Rect(0, 0, 600, 400))

	// background
	drawFilledRect(img, 0, 0, 600, 400, WHITE)

	// title
	pingType := "ICMP Ping"
	if service.Type == "tcp" {
		pingType = "TCP Ping"
	}
	drawText(img, pingType+" from "+server.Label+" to "+service.Label, FONT20, 20, 0, BLACK)

	// graph outline
	drawFilledRect(img, 60, 40, 2, 400-40-40, BLACK)
	drawFilledRect(img, 60, 400-40, 600-60-20, 2, BLACK)

	// y axis
	var maxY, minY int64
	if len(metrics) > 0 {
		maxY = slices.MaxFunc(metrics, func(a, b db.ServiceMetric) int {
			return int(a.Max.Int64) - int(b.Max.Int64)
		}).Max.Int64
		minY = max(slices.MinFunc(metrics, func(a, b db.ServiceMetric) int {
			return int(a.Min.Int64) - int(b.Min.Int64)
		}).Min.Int64, 0)

		medianSum := 0
		for _, m := range metrics {
			medianSum += int(m.Median.Int64)
		}
		avgMedian := medianSum / len(metrics)

		if maxY > int64(avgMedian)*5 {
			maxY = int64(avgMedian) * 5
		}
		if minY < int64(avgMedian)/5 {
			minY = int64(avgMedian) / 5
		}
	}

	ms := true
	var niceMinY, niceMaxY int64

	if maxY-minY < 1000 {
		ms = false
		niceMinY = (minY/100 - 1) * 100
		niceMaxY = (maxY/100 + 1) * 100
	} else if maxY-minY < 30000 {
		niceMinY = (minY/2000 - 1) * 2000
		niceMaxY = (maxY/2000 + 1) * 2000
	} else if maxY-minY < 50000 {
		niceMinY = (minY/5000 - 1) * 5000
		niceMaxY = (maxY/5000 + 1) * 5000
	} else if maxY-minY < 100000 {
		niceMinY = (minY/10000 - 1) * 10000
		niceMaxY = (maxY/10000 + 1) * 10000
	} else if maxY-minY < 300000 {
		niceMinY = (minY/20000 - 1) * 20000
		niceMaxY = (maxY/20000 + 1) * 20000
	} else {
		niceMinY = (minY/50000 - 1) * 50000
		niceMaxY = (maxY/50000 + 1) * 50000
	}

	valuePerStep := int64(0)
	if niceMaxY-niceMinY <= 500 {
		valuePerStep = 20
	} else if niceMaxY-niceMinY <= 1000 {
		valuePerStep = 100
	} else if niceMaxY-niceMinY <= 10000 {
		valuePerStep = 1000
	} else if niceMaxY-niceMinY <= 50000 {
		valuePerStep = 5000
	} else if niceMaxY-niceMinY <= 100000 {
		valuePerStep = 20000
	} else if niceMaxY-niceMinY <= 300000 {
		valuePerStep = 50000
	} else if niceMaxY-niceMinY <= 600000 {
		valuePerStep = 100000
	} else {
		valuePerStep = 200000
	}

	steps := (niceMaxY-niceMinY)/int64(valuePerStep) + 1
	heightPerStep := (400 - 40 - 40) / steps

	for i := range steps + 1 {
		reqTimeout()

		y := int(400 - 40 - int64(i)*heightPerStep)
		drawFilledRect(img, 60-5, y, 5, 2, BLACK)
		str := strconv.Itoa(int(valuePerStep*i+niceMinY)) + "μs"
		if ms {
			str = strconv.Itoa(int(valuePerStep*i+niceMinY)/1000) + "ms"
		}
		drawText(img, str, FONT12, 60-5-5-measureWidth(FONT12, str), y-10, BLACK)
	}

	minValue := niceMinY
	maxValue := valuePerStep*(steps) + niceMinY
	minValueY := int64(400 - 40)
	maxValueY := 400 - 40 - int64(steps)*heightPerStep

	// main content

	linear := func(value int64, x0 int64, x1 int64, y0 int64, y1 int64) int {
		// linear interpolate 'value' and round to the nearest int
		return int(float64(value-x0)*float64(y1-y0)/float64(x1-x0) + float64(y0) + 0.5)
	}

	linearY := func(i int64) int {
		// map latency to y-coordinate
		return linear(i, minValue, maxValue, minValueY, maxValueY)
	}

	linearT := func(x int) int {
		// map x-coordinate to timestamp
		return linear(int64(x), 60+2, 600-20, int64(timestampStart), int64(timestampEnd))
	}

	lossToColor := func(loss float32) color.RGBA {
		k := (1 - loss) * (1 - loss)
		return color.RGBA{A: 255, R: uint8(200 * (1 - k)), G: uint8(255 * k), B: 0}
	}

	indexMetrics := 0
	for x := 60 + 2; x < 600-20; x += 1 {
		timestamp := linearT(x) / 300 * 300

		var theMetrics db.ServiceMetric
		found := false
		for ; indexMetrics < len(metrics); indexMetrics += 1 {
			theMetrics = metrics[indexMetrics]
			if int(theMetrics.Timestamp) == timestamp {
				found = true
				break
			}

			if int(theMetrics.Timestamp) > timestamp {
				found = false
				break
			}
		}

		mask := image.Rect(60+2, 40, 600-20, 400-40)

		if found {
			drawFilledRectMask(img, x, linearY(theMetrics.Max.Int64)+1, 1, linearY(theMetrics.Min.Int64)-linearY(theMetrics.Max.Int64), SMOKE, mask)
			drawFilledRectMask(img, x, linearY(theMetrics.Median.Int64), 1, 2, lossToColor(theMetrics.Loss), mask)
		} else {
			drawFilledRectMask(img, x, 40, 1, 400-40-40, NO_DATA, mask)
		}
	}

	// draw x axis ticks

	tickTable := []int{300, 600, 1800, 3600, 3600 * 2, 3600 * 4, 3600 * 8}

	tickInterval := 3600 * 24

	for _, v := range tickTable {
		if float64(600-60-20)/float64(timestampEnd-timestampStart)*float64(v) > 50 {
			tickInterval = v
			break
		}
	}

	linearX := func(t int) int {
		// map timestamp x-coordinate
		return linear(int64(t), int64(timestampStart), int64(timestampEnd), 60+2, 600-20)
	}

	t := (timestampStart + tickInterval - 1) / tickInterval * tickInterval

	for ; ; t += tickInterval {
		x := linearX(t)

		if x >= 600-20 {
			break
		}

		drawFilledRect(img, int(x), 400-40, 2, 5, BLACK)

		tLocal := time.Unix(int64(t), 0).In(time.FixedZone("", timeZoneOffset))

		var s string
		if tickInterval < 3600*24 {
			s = tLocal.Format("15:04")
		} else {
			s = tLocal.Format("01-02")
		}
		drawText(img, s, FONT12, x-measureWidth(FONT12, s)/2, 400-30, BLACK)
	}

	// encode to png

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	return c.Stream(http.StatusOK, "image/png", &buf)
}

func init() {
	f, err := opentype.Parse(NOTOSANSSC)
	if err != nil {
		panic(err)
	}

	FONT16, err = opentype.NewFace(f, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		panic(err)
	}

	FONT12, err = opentype.NewFace(f, &opentype.FaceOptions{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		panic(err)
	}

	FONT20, err = opentype.NewFace(f, &opentype.FaceOptions{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		panic(err)
	}

}
