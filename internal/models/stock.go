package models

// StockData 股票数据结构
type StockData struct {
	Name      string  // 股票名称
	Code      string  // 股票代码
	Price     float64 // 当前价格
	Change    float64 // 涨跌额
	ChangePct float64 // 涨跌幅
	High      float64 // 最高价
	Low       float64 // 最低价
}
