package services

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/luckfunc/golangBot/internal/models"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HandleStockQuery 处理股票查询
func HandleStockQuery(msg *openwechat.Message) {
	// 去掉"股票"二字，保留后面的代码部分
	content := strings.TrimPrefix(msg.Content, "股票")
	content = strings.TrimSpace(content) // 去掉可能的空格

	// 从消息中提取股票代码
	code := extractStockCode(content)
	if code == "" {
		msg.ReplyText("请输入正确的股票代码，例如：\n" +
			"1. 直接输入代码：股票600519 或 股票000001")
		return
	}

	// 获取股票数据
	stock, err := getStockData(code)
	if err != nil {
		msg.ReplyText(fmt.Sprintf("获取股票数据失败: %v", err))
		return
	}

	// 构造回复消息
	reply := formatStockMessage(stock)
	msg.ReplyText(reply)
}

// 从新浪财经API获取股票数据
func getStockData(code string) (*models.StockData, error) {
	url := fmt.Sprintf("http://hq.sinajs.cn/list=%s", code)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头，模拟浏览器访问
	req.Header.Set("Referer", "https://finance.sina.com.cn")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 将 GBK 编码转换为 UTF-8
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Body, err := decoder.Bytes(body)
	if err != nil {
		return nil, err
	}

	return parseStockData(string(utf8Body), code)
}

// 解析股票数据
func parseStockData(data string, code string) (*models.StockData, error) {
	parts := strings.Split(data, "\"")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid stock data")
	}

	values := strings.Split(parts[1], ",")
	if len(values) < 32 {
		return nil, fmt.Errorf("insufficient stock data")
	}

	// 清理股票名称中的 XD 标记
	stockName := strings.ReplaceAll(values[0], "XD", "")
	stockName = strings.TrimSpace(stockName)

	// 解析价格数据
	currentPrice, _ := strconv.ParseFloat(values[3], 64)
	yesterdayClose, _ := strconv.ParseFloat(values[2], 64)
	high, _ := strconv.ParseFloat(values[4], 64)
	low, _ := strconv.ParseFloat(values[5], 64)

	// 计算涨跌
	change := currentPrice - yesterdayClose
	changePct := change / yesterdayClose * 100

	return &models.StockData{
		Name:      stockName, // 使用清理后的名称
		Code:      code,
		Price:     currentPrice,
		Change:    change,
		ChangePct: changePct,
		High:      high,
		Low:       low,
	}, nil
}

// 格式化股票消息
func formatStockMessage(stock *models.StockData) string {
	// 根据涨跌选择不同的emoji
	var trend string
	if stock.Change > 0 {
		trend = "📈"
	} else if stock.Change < 0 {
		trend = "📉"
	} else {
		trend = "➖"
	}

	return fmt.Sprintf("%s %s (%s)\n"+
		"当前价：%.2f\n"+
		"涨跌额：%.2f\n"+
		"涨跌幅：%.2f%%\n"+
		"最高价：%.2f\n"+
		"最低价：%.2f\n"+
		"更新时间：%s",
		trend, stock.Name, stock.Code,
		stock.Price,
		stock.Change,
		stock.ChangePct,
		stock.High,
		stock.Low,
		time.Now().Format("15:04:05"))
}

// 从消息中提取股票代码
func extractStockCode(content string) string {
	parts := strings.Fields(content)
	for _, part := range parts {
		// 如果直接输入 sh/sz 开头的代码
		if strings.HasPrefix(part, "sh") || strings.HasPrefix(part, "sz") {
			return part
		}

		// 如果是6位数字，先查询股票信息
		if len(part) == 6 && isNumeric(part) {
			// 同时查询沪深两市的股票信息
			url := fmt.Sprintf("http://hq.sinajs.cn/list=sh%s,sz%s", part, part)
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				continue
			}

			req.Header.Set("Referer", "https://finance.sina.com.cn")
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			// 将 GBK 编码转换为 UTF-8
			decoder := simplifiedchinese.GBK.NewDecoder()
			utf8Body, err := decoder.Bytes(body)
			if err != nil {
				continue
			}

			// 解析返回数据，确定是沪市还是深市
			lines := strings.Split(string(utf8Body), "\n")
			for _, line := range lines {
				if strings.Contains(line, "sh"+part) && len(strings.Split(line, "\"")[1]) > 0 {
					return "sh" + part
				}
				if strings.Contains(line, "sz"+part) && len(strings.Split(line, "\"")[1]) > 0 {
					return "sz" + part
				}
			}
		}
	}
	return ""
}

// 判断字符串是否为数字
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// HandleMarketOverview 处理大盘行情查询
func HandleMarketOverview(msg *openwechat.Message) {
	// 获取三大指数数据
	sh, err := getStockData("sh000001") // 上证指数
	if err != nil {
		msg.ReplyText("获取上证指数失败")
		return
	}

	sz, err := getStockData("sz399001") // 深证成指
	if err != nil {
		msg.ReplyText("获取深证成指失败")
		return
	}

	cyb, err := getStockData("sz399006") // 创业板指
	if err != nil {
		msg.ReplyText("获取创业板指失败")
		return
	}

	// 构造回复消息
	reply := formatMarketOverview(sh, sz, cyb)
	msg.ReplyText(reply)
}

// formatMarketOverview 格式化大盘概览消息
func formatMarketOverview(sh, sz, cyb *models.StockData) string {
	// 获取整体趋势图标
	var overallTrend string
	if sh.Change > 0 && sz.Change > 0 && cyb.Change > 0 {
		overallTrend = "🔥 大盘全线上涨"
	} else if sh.Change < 0 && sz.Change < 0 && cyb.Change < 0 {
		overallTrend = "💧 大盘全线下跌"
	} else {
		overallTrend = "📊 大盘涨跌互现"
	}

	return fmt.Sprintf("%s\n\n"+
		"上证指数：%.2f (%+.2f%%)\n"+
		"深证成指：%.2f (%+.2f%%)\n"+
		"创业板指：%.2f (%+.2f%%)\n\n"+
		"更新时间：%s",
		overallTrend,
		sh.Price, sh.ChangePct,
		sz.Price, sz.ChangePct,
		cyb.Price, cyb.ChangePct,
		time.Now().Format("15:04:05"))
}
