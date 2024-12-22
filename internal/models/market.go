package models

// MarketOverview 大盘概览
type MarketOverview struct {
    ShangHai  *StockData // 上证指数
    ShenZhen  *StockData // 深证成指
    ChiNext   *StockData // 创业板指
} 