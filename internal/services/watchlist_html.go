package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/luckfunc/golangBot/internal/models"
	"html/template"
	"strings"
	"time"
)

const watchlistImageWidth = 1280

type watchlistIndexView struct {
	Name  string
	Price string
	Pct   string
	Class string
}

type watchlistRowView struct {
	Code  string
	Name  string
	Price string
	Pct   string
	Chg   string
	Class string
}

type watchlistView struct {
	Title     string
	Timestamp string
	Indices   []watchlistIndexView
	Rows      []watchlistRowView
}

func renderWatchlistHTMLImage(title string, indices []indexSnapshot, stocks []*models.StockData, timestamp string) ([]byte, error) {
	view := watchlistView{
		Title:     title,
		Timestamp: timestamp,
		Indices:   buildIndexViews(indices),
		Rows:      buildRowViews(stocks),
	}
	html, err := renderWatchlistHTML(view)
	if err != nil {
		return nil, err
	}
	height := estimateWatchlistHeight(len(view.Rows), len(view.Indices))
	return renderHTMLToPNG(html, watchlistImageWidth, height)
}

func renderWatchlistHTML(view watchlistView) (string, error) {
	tpl, err := template.New("watchlist").Parse(watchlistHTMLTemplate)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tpl.Execute(&builder, view); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func buildIndexViews(indices []indexSnapshot) []watchlistIndexView {
	out := make([]watchlistIndexView, 0, len(indices))
	for _, idx := range indices {
		out = append(out, watchlistIndexView{
			Name:  idx.Name,
			Price: fmt.Sprintf("%.2f", idx.Stock.Price),
			Pct:   fmt.Sprintf("%+.2f%%", idx.Stock.ChangePct),
			Class: trendClass(idx.Stock.Change),
		})
	}
	return out
}

func buildRowViews(stocks []*models.StockData) []watchlistRowView {
	out := make([]watchlistRowView, 0, len(stocks))
	for _, stock := range stocks {
		out = append(out, watchlistRowView{
			Code:  stock.Code,
			Name:  stock.Name,
			Price: fmt.Sprintf("%.2f", stock.Price),
			Pct:   fmt.Sprintf("%+.2f%%", stock.ChangePct),
			Chg:   fmt.Sprintf("%+.2f", stock.Change),
			Class: trendClass(stock.Change),
		})
	}
	return out
}

func trendClass(change float64) string {
	if change > 0 {
		return "up"
	}
	if change < 0 {
		return "down"
	}
	return "flat"
}

func estimateWatchlistHeight(rows int, indices int) int64 {
	const (
		basePadding    = 80
		titleHeight    = 42
		indexHeight    = 34
		headerHeight   = 44
		rowHeight      = 48
		footerHeight   = 28
		sectionSpacing = 18
	)
	height := basePadding + titleHeight + headerHeight + footerHeight + sectionSpacing*2
	if indices > 0 {
		height += indexHeight + sectionSpacing
	}
	if rows < 1 {
		rows = 1
	}
	height += rows * rowHeight
	return int64(height)
}

func renderHTMLToPNG(html string, width int, height int64) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(html))
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(width), height),
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.FullScreenshot(&buf, 100),
	)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

const watchlistHTMLTemplate = `<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="UTF-8" />
  <style>
    :root {
      --bg: #ffffff;
      --text: #1f1f1f;
      --muted: #6f6f6f;
      --line: #f0f0f0;
      --header: #f7f7f7;
      --up: #d83a3a;
      --down: #1ca05c;
      --flat: #8f8f8f;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      background: var(--bg);
      font-family: "Maple Mono NF CN", "PingFang SC", "PingFang TC", "Microsoft Yahei", sans-serif;
      color: var(--text);
    }
    .container {
      width: 1200px;
      padding: 32px 40px 36px 40px;
    }
    .title {
      font-size: 30px;
      font-weight: 600;
      margin-bottom: 14px;
    }
    .indices {
      display: flex;
      align-items: center;
      gap: 16px;
      font-size: 18px;
      color: var(--muted);
      margin-bottom: 18px;
      flex-wrap: wrap;
    }
    .indices span {
      margin-left: 8px;
    }
    .table {
      width: 100%;
      border-collapse: collapse;
      font-size: 18px;
    }
    .table thead th {
      background: var(--header);
      color: var(--muted);
      font-weight: 500;
      padding: 12px 12px;
      text-align: left;
      border-bottom: 1px solid var(--line);
    }
    .table tbody td {
      padding: 14px 12px;
      border-bottom: 1px solid var(--line);
    }
    .table tbody tr:nth-child(even) td {
      background: #fbfbfb;
    }
    .num { text-align: left; font-variant-numeric: tabular-nums; }
    .up { color: var(--up); }
    .down { color: var(--down); }
    .flat { color: var(--flat); }
    .footer {
      margin-top: 12px;
      font-size: 14px;
      color: var(--muted);
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="title">{{.Title}}</div>
    {{if .Indices}}
    <div class="indices">
      <span>大盘：</span>
      {{range .Indices}}
        <span>{{.Name}} {{.Price}} <span class="{{.Class}}">{{.Pct}}</span></span>
      {{end}}
    </div>
    {{end}}
    <table class="table">
      <thead>
        <tr>
          <th style="width: 140px;">代码</th>
          <th>名称</th>
          <th class="num" style="width: 160px;">现价</th>
          <th class="num" style="width: 160px;">涨幅</th>
          <th class="num" style="width: 160px;">涨跌</th>
        </tr>
      </thead>
      <tbody>
        {{if .Rows}}
          {{range .Rows}}
            <tr>
              <td>{{.Code}}</td>
              <td>{{.Name}}</td>
              <td class="num">{{.Price}}</td>
              <td class="num {{.Class}}">{{.Pct}}</td>
              <td class="num {{.Class}}">{{.Chg}}</td>
            </tr>
          {{end}}
        {{else}}
          <tr>
            <td colspan="5" style="color: var(--muted);">暂无可展示的股票数据</td>
          </tr>
        {{end}}
      </tbody>
    </table>
    <div class="footer">更新时间：{{.Timestamp}}</div>
  </div>
</body>
</html>`
