package models

// WatchlistStore stores per-group stock watchlists.
type WatchlistStore struct {
	Version int                        `json:"version"`
	Groups  map[string]*GroupWatchlist `json:"groups"`
}

// GroupWatchlist represents a group's watchlist and subscription settings.
type GroupWatchlist struct {
	GroupID        string         `json:"group_id"`
	GroupName      string         `json:"group_name"`
	Stocks         []string       `json:"stocks"`
	Subscribed     bool           `json:"subscribed"`
	StockIntervals map[string]int `json:"stock_intervals"`
	Enabled        bool           `json:"enabled"`
	DefaultLimit   int            `json:"default_limit"`
	WindowMinutes  int            `json:"window_minutes"`
	UserLimits     map[string]int `json:"user_limits"`
	UpdatedAt      string         `json:"updated_at"`
}
