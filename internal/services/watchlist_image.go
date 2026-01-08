package services

import (
	"bytes"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/luckfunc/golangBot/internal/models"
	"image/color"
	"image/png"
	"strings"
)

const watchlistFontPath = "/Users/xdd/Library/Fonts/MapleMono-NF-CN-Regular.ttf"

type indexSnapshot struct {
	Name  string
	Stock *models.StockData
}

func renderWatchlistImage(title string, indices []indexSnapshot, stocks []*models.StockData, timestamp string) ([]byte, error) {
	const (
		width      = 1280
		padding    = 40
		titleSize  = 32
		subSize    = 22
		headerSize = 19
		rowSize    = 19
		footerSize = 16
		headerH    = 52
		rowH       = 46
		sectionGap = 22
		rowGap     = 10
		footerGap  = 12
		minRows    = 1
	)

	rowCount := len(stocks)
	if rowCount < minRows {
		rowCount = minRows
	}

	height := padding + titleSize + sectionGap
	if len(indices) > 0 {
		height += subSize + sectionGap
	}
	height += headerH + rowGap + rowCount*rowH + footerGap + footerSize + padding

	dc := gg.NewContext(width, height)
	dc.SetColor(color.White)
	dc.Clear()

	if err := dc.LoadFontFace(watchlistFontPath, titleSize); err != nil {
		return nil, err
	}
	dc.SetColor(color.Black)
	titleHeight := dc.FontHeight()
	y := float64(padding) + titleHeight
	dc.DrawStringAnchored(title, float64(padding), y, 0, 1)

	y += sectionGap
	if len(indices) > 0 {
		if err := dc.LoadFontFace(watchlistFontPath, subSize); err != nil {
			return nil, err
		}
		subFontHeight := dc.FontHeight()
		subBaseline := y + sectionGap + subFontHeight
		x := float64(padding)
		dc.SetColor(color.RGBA{R: 90, G: 90, B: 90, A: 255})
		dc.DrawStringAnchored("大盘：", x, subBaseline, 0, 1)
		labelWidth, _ := dc.MeasureString("大盘：")
		x += labelWidth + 6
		for i, idx := range indices {
			text := fmt.Sprintf("%s %.2f (%+.2f%%)", idx.Name, idx.Stock.Price, idx.Stock.ChangePct)
			dc.SetColor(trendColor(idx.Stock.Change))
			dc.DrawStringAnchored(text, x, subBaseline, 0, 1)
			textWidth, _ := dc.MeasureString(text)
			x += textWidth + 12
			if i == len(indices)-1 {
				break
			}
		}
		y = subBaseline + sectionGap
	}

	headerTop := y
	dc.SetColor(color.RGBA{R: 245, G: 245, B: 245, A: 255})
	dc.DrawRectangle(float64(padding), headerTop, float64(width-padding*2), headerH)
	dc.Fill()

	if err := dc.LoadFontFace(watchlistFontPath, headerSize); err != nil {
		return nil, err
	}
	headerBaseline := headerTop + (headerH+dc.FontHeight())/2
	dc.SetColor(color.RGBA{R: 100, G: 100, B: 100, A: 255})
	columnX := watchlistColumnPositions(width, padding)
	dc.DrawStringAnchored("代码", columnX[0], headerBaseline, 0, 1)
	dc.DrawStringAnchored("名称", columnX[1], headerBaseline, 0, 1)
	dc.DrawStringAnchored("现价", columnX[2], headerBaseline, 0, 1)
	dc.DrawStringAnchored("涨幅", columnX[3], headerBaseline, 0, 1)
	dc.DrawStringAnchored("涨跌", columnX[4], headerBaseline, 0, 1)

	y = headerTop + headerH + rowGap
	if err := dc.LoadFontFace(watchlistFontPath, rowSize); err != nil {
		return nil, err
	}
	rowFontHeight := dc.FontHeight()
	rowBaselineOffset := (float64(rowH)-rowFontHeight)/2 + rowFontHeight

	if len(stocks) == 0 {
		dc.SetColor(color.RGBA{R: 120, G: 120, B: 120, A: 255})
		dc.DrawStringAnchored("暂无可展示的股票数据", float64(padding), y+rowBaselineOffset, 0, 1)
		y += rowH
	} else {
		for i, stock := range stocks {
			rowTop := y + float64(i*rowH)
			if i%2 == 1 {
				dc.SetColor(color.RGBA{R: 250, G: 250, B: 250, A: 255})
				dc.DrawRectangle(float64(padding), rowTop, float64(width-padding*2), rowH)
				dc.Fill()
			}
			rowY := rowTop + rowBaselineOffset
			dc.SetColor(color.Black)
			dc.DrawStringAnchored(stock.Code, columnX[0], rowY, 0, 1)
			dc.DrawStringAnchored(trimName(stock.Name, 10), columnX[1], rowY, 0, 1)
			dc.DrawStringAnchored(fmt.Sprintf("%.2f", stock.Price), columnX[2], rowY, 0, 1)
			dc.SetColor(trendColor(stock.Change))
			dc.DrawStringAnchored(fmt.Sprintf("%+.2f%%", stock.ChangePct), columnX[3], rowY, 0, 1)
			dc.DrawStringAnchored(fmt.Sprintf("%+.2f", stock.Change), columnX[4], rowY, 0, 1)
		}
		y += float64(len(stocks) * rowH)
	}

	y += footerGap
	if err := dc.LoadFontFace(watchlistFontPath, footerSize); err != nil {
		return nil, err
	}
	dc.SetColor(color.RGBA{R: 130, G: 130, B: 130, A: 255})
	dc.DrawStringAnchored("更新时间："+timestamp, float64(padding), y, 0, 1)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func trendColor(change float64) color.Color {
	if change > 0 {
		return color.RGBA{R: 220, G: 68, B: 68, A: 255}
	}
	if change < 0 {
		return color.RGBA{R: 28, G: 160, B: 92, A: 255}
	}
	return color.RGBA{R: 120, G: 120, B: 120, A: 255}
}

func watchlistColumnPositions(width, padding int) []float64 {
	base := float64(padding)
	return []float64{
		base,
		base + 190,
		base + 620,
		base + 860,
		base + 1080,
	}
}

func trimName(name string, max int) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) <= max {
		return name
	}
	return strings.TrimSpace(string(runes[:max])) + "..."
}
