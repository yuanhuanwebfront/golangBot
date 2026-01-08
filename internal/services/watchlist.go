package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/luckfunc/golangBot/internal/models"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

const watchlistFileName = "watchlist.json"
const dailyPushTime = "15:05"

var watchlistMu sync.Mutex
var lastPushDateMu sync.Mutex
var lastPushDate = make(map[string]string)
var intervalPushMu sync.Mutex
var lastIntervalPush = make(map[string]time.Time)

type indexSnapshot struct {
	Name  string
	Stock *models.StockData
}

// HandleStockCommand handles stock-related commands.
func HandleStockCommand(msg *openwechat.Message) {
	content := strings.TrimSpace(msg.Content)
	switch {
	case strings.HasPrefix(content, "股票添加"):
		handleWatchlistAdd(msg, strings.TrimSpace(strings.TrimPrefix(content, "股票添加")))
	case strings.HasPrefix(content, "股票删除"):
		handleWatchlistRemove(msg, strings.TrimSpace(strings.TrimPrefix(content, "股票删除")))
	case strings.HasPrefix(content, "股票移除"):
		handleWatchlistRemove(msg, strings.TrimSpace(strings.TrimPrefix(content, "股票移除")))
	case strings.HasPrefix(content, "股票列表"):
		handleWatchlistList(msg)
	case strings.HasPrefix(content, "股票波动"):
		handleWatchlistOverview(msg)
	case strings.HasPrefix(content, "股票订阅"):
		handleWatchlistSubscribe(msg, true)
	case strings.HasPrefix(content, "股票取消订阅"):
		handleWatchlistSubscribe(msg, false)
	case strings.HasPrefix(content, "股票退订"):
		handleWatchlistSubscribe(msg, false)
	case strings.HasPrefix(content, "股票关闭"):
		handleWatchlistEnabled(msg, false)
	case strings.HasPrefix(content, "股票开启"):
		handleWatchlistEnabled(msg, true)
	case strings.HasPrefix(content, "股票定时列表"):
		handleWatchlistIntervalList(msg)
	case strings.HasPrefix(content, "股票定时"):
		handleWatchlistIntervalSet(msg, strings.TrimSpace(strings.TrimPrefix(content, "股票定时")))
	case strings.HasPrefix(content, "股票帮助"):
		replyStockHelp(msg)
	default:
		HandleStockQuery(msg)
	}
}

// StartDailyWatchlistPush sends daily watchlist overview to subscribed groups.
func StartDailyWatchlistPush(bot *openwechat.Bot) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			if now.Format("15:04") != dailyPushTime {
				continue
			}
			store, err := loadWatchlistStore()
			if err != nil || len(store.Groups) == 0 {
				continue
			}
			self, err := bot.GetCurrentUser()
			if err != nil {
				continue
			}
			groups, err := self.Groups()
			if err != nil {
				continue
			}
			for groupID, group := range store.Groups {
				if !group.Enabled {
					continue
				}
				if !group.Subscribed || len(group.Stocks) == 0 {
					continue
				}
				if pushedToday(groupID, now) {
					continue
				}
				target := groups.SearchByUserName(1, groupID)
				if target.Count() == 0 {
					continue
				}
				image, err := buildWatchlistOverviewImage(group.Stocks, group.GroupName, "每日收盘")
				if err == nil {
					_, _ = target.First().SendImage(bytes.NewReader(image))
				} else {
					message := buildWatchlistOverview(group.Stocks, group.GroupName, "每日收盘")
					_, _ = target.First().SendText(message)
				}
				markPushed(groupID, now)
			}
		}
	}()
}

// StartIntervalWatchlistPush sends interval-based updates for selected stocks.
func StartIntervalWatchlistPush(bot *openwechat.Bot) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			store, err := loadWatchlistStore()
			if err != nil || len(store.Groups) == 0 {
				continue
			}
			self, err := bot.GetCurrentUser()
			if err != nil {
				continue
			}
			groups, err := self.Groups()
			if err != nil {
				continue
			}
			now := time.Now()
			for groupID, group := range store.Groups {
				if !group.Enabled {
					continue
				}
				if len(group.StockIntervals) == 0 {
					continue
				}
				target := groups.SearchByUserName(1, groupID)
				if target.Count() == 0 {
					continue
				}
				for code, minutes := range group.StockIntervals {
					if minutes <= 0 {
						continue
					}
					if !shouldIntervalPush(groupID, code, minutes, now) {
						continue
					}
					stock, err := getStockData(code)
					if err != nil {
						continue
					}
					title := fmt.Sprintf("股票定时提醒（%d分钟）", minutes)
					indices := fetchMarketIndexSnapshots()
					image, err := renderWatchlistHTMLImage(title, indices, []*models.StockData{stock}, now.Format("15:04:05"))
					if err == nil {
						_, _ = target.First().SendImage(bytes.NewReader(image))
					} else {
						message := fmt.Sprintf("%s\n%s\n更新时间：%s",
							title,
							formatWatchlistTable([]*models.StockData{stock}),
							now.Format("15:04:05"))
						_, _ = target.First().SendText(message)
					}
					markIntervalPushed(groupID, code, now)
				}
			}
		}
	}()
}

func handleWatchlistAdd(msg *openwechat.Message, args string) {
	codes := parseStockCodes(args)
	if len(codes) == 0 {
		msg.ReplyText("用法：股票添加 600519 / 股票添加 sh600519 sz000001")
		return
	}
	groupID, groupName := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中添加关注股票")
		return
	}
	resolved := resolveCodes(codes)
	if len(resolved) == 0 {
		msg.ReplyText("没有识别到有效的股票代码")
		return
	}
	added, existed, err := addStocksToWatchlist(groupID, groupName, resolved)
	if err != nil {
		msg.ReplyText(fmt.Sprintf("添加失败：%v", err))
		return
	}
	var parts []string
	if len(added) > 0 {
		parts = append(parts, fmt.Sprintf("已添加：%s", strings.Join(added, ", ")))
	}
	if len(existed) > 0 {
		parts = append(parts, fmt.Sprintf("已存在：%s", strings.Join(existed, ", ")))
	}
	msg.ReplyText(strings.Join(parts, "\n"))
}

func handleWatchlistRemove(msg *openwechat.Message, args string) {
	codes := parseStockCodes(args)
	if len(codes) == 0 {
		msg.ReplyText("用法：股票删除 600519 / 股票删除 sh600519 sz000001")
		return
	}
	groupID, groupName := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中删除关注股票")
		return
	}
	resolved := resolveCodes(codes)
	if len(resolved) == 0 {
		msg.ReplyText("没有识别到有效的股票代码")
		return
	}
	removed, missed, err := removeStocksFromWatchlist(groupID, groupName, resolved)
	if err != nil {
		msg.ReplyText(fmt.Sprintf("删除失败：%v", err))
		return
	}
	var parts []string
	if len(removed) > 0 {
		parts = append(parts, fmt.Sprintf("已删除：%s", strings.Join(removed, ", ")))
	}
	if len(missed) > 0 {
		parts = append(parts, fmt.Sprintf("未关注：%s", strings.Join(missed, ", ")))
	}
	msg.ReplyText(strings.Join(parts, "\n"))
}

func handleWatchlistList(msg *openwechat.Message) {
	groupID, _ := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中查看列表")
		return
	}
	store, err := loadWatchlistStore()
	if err != nil {
		msg.ReplyText(fmt.Sprintf("读取列表失败：%v", err))
		return
	}
	group := store.Groups[groupID]
	if group == nil || len(group.Stocks) == 0 {
		msg.ReplyText("当前没有关注股票，可用：股票添加 600519")
		return
	}
	msg.ReplyText(fmt.Sprintf("关注列表（%d）：%s", len(group.Stocks), strings.Join(group.Stocks, ", ")))
}

func handleWatchlistOverview(msg *openwechat.Message) {
	groupID, _ := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中查看波动")
		return
	}
	store, err := loadWatchlistStore()
	if err != nil {
		msg.ReplyText(fmt.Sprintf("读取列表失败：%v", err))
		return
	}
	group := store.Groups[groupID]
	if group == nil || len(group.Stocks) == 0 {
		msg.ReplyText("当前没有关注股票，可用：股票添加 600519")
		return
	}
	image, err := buildWatchlistOverviewImage(group.Stocks, group.GroupName, "当前行情")
	if err == nil {
		_, _ = msg.ReplyImage(bytes.NewReader(image))
		return
	}
	message := buildWatchlistOverview(group.Stocks, group.GroupName, "当前行情")
	msg.ReplyText(fmt.Sprintf("生成图片失败：%v\n%s", err, message))
}

func handleWatchlistSubscribe(msg *openwechat.Message, subscribe bool) {
	groupID, groupName := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中订阅")
		return
	}
	if err := setWatchlistSubscription(groupID, groupName, subscribe); err != nil {
		msg.ReplyText(fmt.Sprintf("订阅设置失败：%v", err))
		return
	}
	if subscribe {
		msg.ReplyText(fmt.Sprintf("已开启每日推送（%s）", dailyPushTime))
		return
	}
	msg.ReplyText("已关闭每日推送")
}

func handleWatchlistEnabled(msg *openwechat.Message, enabled bool) {
	groupID, groupName := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中设置")
		return
	}
	if err := setWatchlistEnabled(groupID, groupName, enabled); err != nil {
		msg.ReplyText(fmt.Sprintf("设置失败：%v", err))
		return
	}
	if enabled {
		msg.ReplyText("已开启股票消息推送")
		return
	}
	msg.ReplyText("已关闭股票消息推送")
}

func handleWatchlistIntervalSet(msg *openwechat.Message, args string) {
	fields := strings.Fields(args)
	if len(fields) < 2 {
		msg.ReplyText("用法：股票定时 600519 30（单位分钟，0 为关闭）")
		return
	}
	code := strings.TrimSpace(fields[0])
	minutesText := strings.TrimSpace(fields[1])
	if minutesText == "关闭" {
		minutesText = "0"
	}
	minutes, err := strconv.Atoi(minutesText)
	if err != nil || minutes < 0 {
		msg.ReplyText("时间格式不正确，请输入分钟数，例如：股票定时 600519 30")
		return
	}
	groupID, groupName := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中设置定时")
		return
	}
	resolved := resolveCodes([]string{code})
	if len(resolved) == 0 {
		msg.ReplyText("没有识别到有效的股票代码")
		return
	}
	if err := setWatchlistInterval(groupID, groupName, resolved[0], minutes); err != nil {
		msg.ReplyText(fmt.Sprintf("设置失败：%v", err))
		return
	}
	if minutes == 0 {
		msg.ReplyText(fmt.Sprintf("已关闭 %s 定时提醒", resolved[0]))
		return
	}
	msg.ReplyText(fmt.Sprintf("已设置 %s 每 %d 分钟提醒", resolved[0], minutes))
}

func handleWatchlistIntervalList(msg *openwechat.Message) {
	groupID, _ := resolveGroupInfo(msg)
	if groupID == "" {
		msg.ReplyText("只支持在群聊中查看定时列表")
		return
	}
	store, err := loadWatchlistStore()
	if err != nil {
		msg.ReplyText(fmt.Sprintf("读取列表失败：%v", err))
		return
	}
	group := store.Groups[groupID]
	if group == nil || len(group.StockIntervals) == 0 {
		msg.ReplyText("当前没有定时提醒，可用：股票定时 600519 30")
		return
	}
	var lines []string
	for code, minutes := range group.StockIntervals {
		if minutes <= 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s 每%d分钟", code, minutes))
	}
	if len(lines) == 0 {
		msg.ReplyText("当前没有定时提醒，可用：股票定时 600519 30")
		return
	}
	msg.ReplyText("定时列表：\n" + strings.Join(lines, "\n"))
}

func replyStockHelp(msg *openwechat.Message) {
	msg.ReplyText("股票功能：\n" +
		"1) 查询：股票600519 / 股票 sh600519\n" +
		"2) 添加：股票添加 600519\n" +
		"3) 删除：股票删除 600519\n" +
		"4) 列表：股票列表\n" +
		"5) 波动：股票波动\n" +
		"6) 订阅：股票订阅 / 股票取消订阅\n" +
		"7) 定时：股票定时 600519 30\n" +
		"8) 定时列表：股票定时列表\n" +
		"9) 推送开关：股票开启 / 股票关闭")
}

// HandleStockHelp replies stock help content.
func HandleStockHelp(msg *openwechat.Message) {
	replyStockHelp(msg)
}

func resolveGroupInfo(msg *openwechat.Message) (string, string) {
	if !msg.IsSendByGroup() {
		return "", ""
	}
	groupID := msg.FromUserName
	groupName := ""
	group, err := msg.Receiver()
	if err == nil && group != nil {
		groupName = group.NickName
	}
	return groupID, groupName
}

func parseStockCodes(args string) []string {
	if args == "" {
		return nil
	}
	args = strings.ReplaceAll(args, "，", " ")
	args = strings.ReplaceAll(args, ",", " ")
	fields := strings.Fields(args)
	var codes []string
	for _, field := range fields {
		if field != "" {
			codes = append(codes, strings.TrimSpace(field))
		}
	}
	return codes
}

func resolveCodes(codes []string) []string {
	var resolved []string
	for _, code := range codes {
		if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") {
			resolved = append(resolved, code)
			continue
		}
		if len(code) == 6 && isNumeric(code) {
			guess := resolveStockCode(code)
			if guess != "" {
				resolved = append(resolved, guess)
			}
		}
	}
	return uniqStrings(resolved)
}

func resolveStockCode(code string) string {
	url := fmt.Sprintf("http://hq.sinajs.cn/list=sh%s,sz%s", code, code)
	resp, err := httpGet(url)
	if err != nil {
		return ""
	}
	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\"")
		if len(parts) < 2 {
			continue
		}
		if strings.Contains(line, "sh"+code) && len(parts[1]) > 0 {
			return "sh" + code
		}
		if strings.Contains(line, "sz"+code) && len(parts[1]) > 0 {
			return "sz" + code
		}
	}
	return ""
}

func uniqStrings(values []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, val := range values {
		if seen[val] {
			continue
		}
		seen[val] = true
		out = append(out, val)
	}
	return out
}

func addStocksToWatchlist(groupID, groupName string, codes []string) ([]string, []string, error) {
	watchlistMu.Lock()
	defer watchlistMu.Unlock()
	store, err := loadWatchlistStore()
	if err != nil {
		return nil, nil, err
	}
	group := ensureGroupWatchlist(store, groupID, groupName)
	existing := make(map[string]bool)
	for _, code := range group.Stocks {
		existing[code] = true
	}
	var added []string
	var existed []string
	for _, code := range codes {
		if existing[code] {
			existed = append(existed, code)
			continue
		}
		group.Stocks = append(group.Stocks, code)
		existing[code] = true
		added = append(added, code)
	}
	group.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := saveWatchlistStore(store); err != nil {
		return nil, nil, err
	}
	return added, existed, nil
}

func removeStocksFromWatchlist(groupID, groupName string, codes []string) ([]string, []string, error) {
	watchlistMu.Lock()
	defer watchlistMu.Unlock()
	store, err := loadWatchlistStore()
	if err != nil {
		return nil, nil, err
	}
	group := ensureGroupWatchlist(store, groupID, groupName)
	if len(group.Stocks) == 0 {
		return nil, codes, nil
	}
	toRemove := make(map[string]bool)
	for _, code := range codes {
		toRemove[code] = true
	}
	var kept []string
	var removed []string
	for _, code := range group.Stocks {
		if toRemove[code] {
			removed = append(removed, code)
		} else {
			kept = append(kept, code)
		}
	}
	var missed []string
	for _, code := range codes {
		found := false
		for _, rem := range removed {
			if rem == code {
				found = true
				break
			}
		}
		if !found {
			missed = append(missed, code)
		}
	}
	group.Stocks = kept
	group.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := saveWatchlistStore(store); err != nil {
		return nil, nil, err
	}
	return removed, missed, nil
}

func setWatchlistSubscription(groupID, groupName string, subscribe bool) error {
	watchlistMu.Lock()
	defer watchlistMu.Unlock()
	store, err := loadWatchlistStore()
	if err != nil {
		return err
	}
	group := ensureGroupWatchlist(store, groupID, groupName)
	group.Subscribed = subscribe
	group.UpdatedAt = time.Now().Format(time.RFC3339)
	return saveWatchlistStore(store)
}

func ensureGroupWatchlist(store *models.WatchlistStore, groupID, groupName string) *models.GroupWatchlist {
	if store.Groups == nil {
		store.Groups = make(map[string]*models.GroupWatchlist)
	}
	group, ok := store.Groups[groupID]
	if !ok {
		group = &models.GroupWatchlist{
			GroupID:        groupID,
			GroupName:      groupName,
			Stocks:         []string{},
			StockIntervals: make(map[string]int),
			Enabled:        true,
		}
		store.Groups[groupID] = group
	}
	if group.StockIntervals == nil {
		group.StockIntervals = make(map[string]int)
	}
	if groupName != "" && group.GroupName != groupName {
		group.GroupName = groupName
	}
	return group
}

func loadWatchlistStore() (*models.WatchlistStore, error) {
	path := watchlistFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.WatchlistStore{
				Version: 2,
				Groups:  make(map[string]*models.GroupWatchlist),
			}, nil
		}
		return nil, err
	}
	var store models.WatchlistStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Groups == nil {
		store.Groups = make(map[string]*models.GroupWatchlist)
	}
	if store.Version == 0 {
		store.Version = 1
	}
	if store.Version < 2 {
		for _, group := range store.Groups {
			group.Enabled = true
		}
		store.Version = 2
	}
	return &store, nil
}

func saveWatchlistStore(store *models.WatchlistStore) error {
	path := watchlistFilePath()
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func setWatchlistInterval(groupID, groupName, code string, minutes int) error {
	watchlistMu.Lock()
	defer watchlistMu.Unlock()
	store, err := loadWatchlistStore()
	if err != nil {
		return err
	}
	group := ensureGroupWatchlist(store, groupID, groupName)
	if minutes == 0 {
		delete(group.StockIntervals, code)
	} else {
		group.StockIntervals[code] = minutes
	}
	group.UpdatedAt = time.Now().Format(time.RFC3339)
	return saveWatchlistStore(store)
}

func setWatchlistEnabled(groupID, groupName string, enabled bool) error {
	watchlistMu.Lock()
	defer watchlistMu.Unlock()
	store, err := loadWatchlistStore()
	if err != nil {
		return err
	}
	group := ensureGroupWatchlist(store, groupID, groupName)
	group.Enabled = enabled
	group.UpdatedAt = time.Now().Format(time.RFC3339)
	return saveWatchlistStore(store)
}

func watchlistFilePath() string {
	return filepath.Join(".", watchlistFileName)
}

func buildWatchlistOverview(codes []string, groupName, title string) string {
	var stocks []*models.StockData
	for _, code := range codes {
		stock, err := getStockData(code)
		if err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}
	head := "股票波动"
	if groupName != "" {
		head = fmt.Sprintf("%s - %s", head, groupName)
	}
	return fmt.Sprintf("%s（%s）\n%s\n%s\n更新时间：%s",
		head,
		title,
		formatMarketIndexSummary(),
		formatWatchlistTable(stocks),
		time.Now().Format("15:04:05"))
}

func buildWatchlistOverviewImage(codes []string, groupName, title string) ([]byte, error) {
	stocks := fetchStocksByCodes(codes)
	indices := fetchMarketIndexSnapshots()
	head := "自选行情"
	fullTitle := fmt.Sprintf("%s（%s）", head, title)
	return renderWatchlistHTMLImage(fullTitle, indices, stocks, time.Now().Format("15:04:05"))
}

func fetchStocksByCodes(codes []string) []*models.StockData {
	var stocks []*models.StockData
	for _, code := range codes {
		stock, err := getStockData(code)
		if err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}
	return stocks
}

func fetchMarketIndexSnapshots() []indexSnapshot {
	snapshots := make([]indexSnapshot, 0, 3)
	if sh, err := getStockData("sh000001"); err == nil {
		snapshots = append(snapshots, indexSnapshot{Name: "上证", Stock: sh})
	}
	if sz, err := getStockData("sz399001"); err == nil {
		snapshots = append(snapshots, indexSnapshot{Name: "深证", Stock: sz})
	}
	if cyb, err := getStockData("sz399006"); err == nil {
		snapshots = append(snapshots, indexSnapshot{Name: "创业板", Stock: cyb})
	}
	return snapshots
}

func formatWatchlistTable(stocks []*models.StockData) string {
	if len(stocks) == 0 {
		return "暂无可展示的股票数据"
	}
	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintln(writer, "代码\t名称\t现价\t涨幅\t涨跌")
	for _, stock := range stocks {
		fmt.Fprintf(writer, "%s\t%s\t%.2f\t%+.2f%%\t%+.2f\n",
			stock.Code,
			stock.Name,
			stock.Price,
			stock.ChangePct,
			stock.Change)
	}
	_ = writer.Flush()
	return strings.TrimRight(buf.String(), "\n")
}

func formatMarketIndexSummary() string {
	sh, err := getStockData("sh000001")
	if err != nil {
		return "大盘指数：获取失败"
	}
	sz, err := getStockData("sz399001")
	if err != nil {
		return "大盘指数：获取失败"
	}
	cyb, err := getStockData("sz399006")
	if err != nil {
		return "大盘指数：获取失败"
	}
	return fmt.Sprintf("大盘指数：上证 %.2f(%+.2f%%)  深证 %.2f(%+.2f%%)  创业板 %.2f(%+.2f%%)",
		sh.Price, sh.ChangePct,
		sz.Price, sz.ChangePct,
		cyb.Price, cyb.ChangePct)
}

func pushedToday(groupID string, now time.Time) bool {
	lastPushDateMu.Lock()
	defer lastPushDateMu.Unlock()
	date := now.Format("2006-01-02")
	if lastPushDate[groupID] == date {
		return true
	}
	return false
}

func markPushed(groupID string, now time.Time) {
	lastPushDateMu.Lock()
	defer lastPushDateMu.Unlock()
	lastPushDate[groupID] = now.Format("2006-01-02")
}

func shouldIntervalPush(groupID, code string, minutes int, now time.Time) bool {
	intervalPushMu.Lock()
	defer intervalPushMu.Unlock()
	key := groupID + "|" + code
	last, ok := lastIntervalPush[key]
	if !ok {
		return true
	}
	return now.Sub(last) >= time.Duration(minutes)*time.Minute
}

func markIntervalPushed(groupID, code string, now time.Time) {
	intervalPushMu.Lock()
	defer intervalPushMu.Unlock()
	key := groupID + "|" + code
	lastIntervalPush[key] = now
}

func httpGet(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Body, err := decoder.Bytes(body)
	if err != nil {
		return "", err
	}
	return string(utf8Body), nil
}
