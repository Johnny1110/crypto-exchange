-- =============================================
-- K線資料庫SQLite版本完整表結構設計
-- 適用於加密貨幣現貨交易所（輕量化版本）
-- =============================================

-- 啟用外鍵約束
PRAGMA foreign_keys = ON;

-- 1. 交易對基礎資訊表
CREATE TABLE trading_pairs (
                               id INTEGER PRIMARY KEY AUTOINCREMENT,
                               symbol TEXT NOT NULL UNIQUE, -- 交易對符號，如BTCUSDT
                               base_asset TEXT NOT NULL, -- 基礎資產，如BTC
                               quote_asset TEXT NOT NULL, -- 計價資產，如USDT
                               status TEXT DEFAULT 'TRADING' CHECK (status IN ('TRADING', 'HALT', 'BREAK')), -- 交易狀態
                               price_precision INTEGER DEFAULT 8, -- 價格精度
                               quantity_precision INTEGER DEFAULT 8, -- 數量精度
                               min_notional REAL DEFAULT 0, -- 最小名義價值
                               is_active INTEGER DEFAULT 1 CHECK (is_active IN (0,1)), -- 是否啟用（SQLite用0/1代替布林值）
                               created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                               updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 為trading_pairs建立索引
CREATE INDEX idx_trading_pairs_symbol ON trading_pairs(symbol);
CREATE INDEX idx_trading_pairs_assets ON trading_pairs(base_asset, quote_asset);
CREATE INDEX idx_trading_pairs_status ON trading_pairs(status, is_active);

-- 2. K線主資料表（分時間間隔存儲）

-- 1分鐘K線表
CREATE TABLE klines_1m (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL, -- 交易對符號
                           open_time INTEGER NOT NULL, -- 開盤時間（毫秒時間戳）
                           close_time INTEGER NOT NULL, -- 收盤時間（毫秒時間戳）
                           open_price REAL NOT NULL, -- 開盤價
                           high_price REAL NOT NULL, -- 最高價
                           low_price REAL NOT NULL, -- 最低價
                           close_price REAL NOT NULL, -- 收盤價
                           volume REAL NOT NULL DEFAULT 0, -- 成交量（基礎資產）
                           quote_volume REAL NOT NULL DEFAULT 0, -- 成交額（計價資產）
                           trade_count INTEGER NOT NULL DEFAULT 0, -- 成交筆數
                           taker_buy_volume REAL NOT NULL DEFAULT 0, -- 主動買入成交量
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0, -- 主動買入成交額
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)), -- 該K線是否已完結
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

-- 為1分鐘K線表建立索引
CREATE INDEX idx_klines_1m_symbol_time ON klines_1m(symbol, open_time);
CREATE INDEX idx_klines_1m_symbol_time_range ON klines_1m(symbol, open_time, close_time);
CREATE INDEX idx_klines_1m_close_time ON klines_1m(close_time);
CREATE INDEX idx_klines_1m_created_at ON klines_1m(created_at);

-- 5分鐘K線表
CREATE TABLE klines_5m (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           open_price REAL NOT NULL,
                           high_price REAL NOT NULL,
                           low_price REAL NOT NULL,
                           close_price REAL NOT NULL,
                           volume REAL NOT NULL DEFAULT 0,
                           quote_volume REAL NOT NULL DEFAULT 0,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           taker_buy_volume REAL NOT NULL DEFAULT 0,
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_5m_symbol_time ON klines_5m(symbol, open_time);
CREATE INDEX idx_klines_5m_symbol_time_range ON klines_5m(symbol, open_time, close_time);
CREATE INDEX idx_klines_5m_close_time ON klines_5m(close_time);

-- 15分鐘K線表
CREATE TABLE klines_15m (
                            id INTEGER PRIMARY KEY AUTOINCREMENT,
                            symbol TEXT NOT NULL,
                            open_time INTEGER NOT NULL,
                            close_time INTEGER NOT NULL,
                            open_price REAL NOT NULL,
                            high_price REAL NOT NULL,
                            low_price REAL NOT NULL,
                            close_price REAL NOT NULL,
                            volume REAL NOT NULL DEFAULT 0,
                            quote_volume REAL NOT NULL DEFAULT 0,
                            trade_count INTEGER NOT NULL DEFAULT 0,
                            taker_buy_volume REAL NOT NULL DEFAULT 0,
                            taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                            is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                            UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_15m_symbol_time ON klines_15m(symbol, open_time);
CREATE INDEX idx_klines_15m_symbol_time_range ON klines_15m(symbol, open_time, close_time);
CREATE INDEX idx_klines_15m_close_time ON klines_15m(close_time);

-- 30分鐘K線表
CREATE TABLE klines_30m (
                            id INTEGER PRIMARY KEY AUTOINCREMENT,
                            symbol TEXT NOT NULL,
                            open_time INTEGER NOT NULL,
                            close_time INTEGER NOT NULL,
                            open_price REAL NOT NULL,
                            high_price REAL NOT NULL,
                            low_price REAL NOT NULL,
                            close_price REAL NOT NULL,
                            volume REAL NOT NULL DEFAULT 0,
                            quote_volume REAL NOT NULL DEFAULT 0,
                            trade_count INTEGER NOT NULL DEFAULT 0,
                            taker_buy_volume REAL NOT NULL DEFAULT 0,
                            taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                            is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                            UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_30m_symbol_time ON klines_30m(symbol, open_time);
CREATE INDEX idx_klines_30m_symbol_time_range ON klines_30m(symbol, open_time, close_time);
CREATE INDEX idx_klines_30m_close_time ON klines_30m(close_time);

-- 1小時K線表
CREATE TABLE klines_1h (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           open_price REAL NOT NULL,
                           high_price REAL NOT NULL,
                           low_price REAL NOT NULL,
                           close_price REAL NOT NULL,
                           volume REAL NOT NULL DEFAULT 0,
                           quote_volume REAL NOT NULL DEFAULT 0,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           taker_buy_volume REAL NOT NULL DEFAULT 0,
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_1h_symbol_time ON klines_1h(symbol, open_time);
CREATE INDEX idx_klines_1h_symbol_time_range ON klines_1h(symbol, open_time, close_time);
CREATE INDEX idx_klines_1h_close_time ON klines_1h(close_time);

-- 4小時K線表
CREATE TABLE klines_4h (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           open_price REAL NOT NULL,
                           high_price REAL NOT NULL,
                           low_price REAL NOT NULL,
                           close_price REAL NOT NULL,
                           volume REAL NOT NULL DEFAULT 0,
                           quote_volume REAL NOT NULL DEFAULT 0,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           taker_buy_volume REAL NOT NULL DEFAULT 0,
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_4h_symbol_time ON klines_4h(symbol, open_time);
CREATE INDEX idx_klines_4h_symbol_time_range ON klines_4h(symbol, open_time, close_time);
CREATE INDEX idx_klines_4h_close_time ON klines_4h(close_time);

-- 1天K線表
CREATE TABLE klines_1d (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           open_price REAL NOT NULL,
                           high_price REAL NOT NULL,
                           low_price REAL NOT NULL,
                           close_price REAL NOT NULL,
                           volume REAL NOT NULL DEFAULT 0,
                           quote_volume REAL NOT NULL DEFAULT 0,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           taker_buy_volume REAL NOT NULL DEFAULT 0,
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_1d_symbol_time ON klines_1d(symbol, open_time);
CREATE INDEX idx_klines_1d_symbol_time_range ON klines_1d(symbol, open_time, close_time);
CREATE INDEX idx_klines_1d_close_time ON klines_1d(close_time);

-- 1週K線表
CREATE TABLE klines_1w (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           open_price REAL NOT NULL,
                           high_price REAL NOT NULL,
                           low_price REAL NOT NULL,
                           close_price REAL NOT NULL,
                           volume REAL NOT NULL DEFAULT 0,
                           quote_volume REAL NOT NULL DEFAULT 0,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           taker_buy_volume REAL NOT NULL DEFAULT 0,
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_1w_symbol_time ON klines_1w(symbol, open_time);
CREATE INDEX idx_klines_1w_symbol_time_range ON klines_1w(symbol, open_time, close_time);
CREATE INDEX idx_klines_1w_close_time ON klines_1w(close_time);

-- 1月K線表
CREATE TABLE klines_1M (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           open_price REAL NOT NULL,
                           high_price REAL NOT NULL,
                           low_price REAL NOT NULL,
                           close_price REAL NOT NULL,
                           volume REAL NOT NULL DEFAULT 0,
                           quote_volume REAL NOT NULL DEFAULT 0,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           taker_buy_volume REAL NOT NULL DEFAULT 0,
                           taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_klines_1M_symbol_time ON klines_1M(symbol, open_time);
CREATE INDEX idx_klines_1M_symbol_time_range ON klines_1M(symbol, open_time, close_time);
CREATE INDEX idx_klines_1M_close_time ON klines_1M(close_time);

-- 3. 實時K線緩存表（用於當前未完結的K線）
CREATE TABLE klines_realtime (
                                 id INTEGER PRIMARY KEY AUTOINCREMENT,
                                 symbol TEXT NOT NULL,
                                 interval_type TEXT NOT NULL, -- 時間間隔類型：1m,5m,15m,30m,1h,4h,1d,1w,1M
                                 open_time INTEGER NOT NULL,
                                 close_time INTEGER NOT NULL,
                                 open_price REAL NOT NULL,
                                 high_price REAL NOT NULL,
                                 low_price REAL NOT NULL,
                                 close_price REAL NOT NULL,
                                 volume REAL NOT NULL DEFAULT 0,
                                 quote_volume REAL NOT NULL DEFAULT 0,
                                 trade_count INTEGER NOT NULL DEFAULT 0,
                                 taker_buy_volume REAL NOT NULL DEFAULT 0,
                                 taker_buy_quote_volume REAL NOT NULL DEFAULT 0,
                                 last_trade_id INTEGER, -- 最後一筆交易ID
                                 updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                 UNIQUE(symbol, interval_type, open_time)
);

CREATE INDEX idx_klines_realtime_symbol_interval ON klines_realtime(symbol, interval_type);
CREATE INDEX idx_klines_realtime_updated_at ON klines_realtime(updated_at);

-- 4. K線資料統計表（用於優化查詢）
CREATE TABLE klines_statistics (
                                   id INTEGER PRIMARY KEY AUTOINCREMENT,
                                   symbol TEXT NOT NULL,
                                   interval_type TEXT NOT NULL,
                                   date_key DATE NOT NULL, -- 統計日期
                                   record_count INTEGER NOT NULL DEFAULT 0, -- 當日K線記錄數
                                   min_open_time INTEGER, -- 最小開盤時間
                                   max_close_time INTEGER, -- 最大收盤時間
                                   avg_volume REAL, -- 平均成交量
                                   total_volume REAL, -- 總成交量
                                   created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                   updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                   UNIQUE(symbol, interval_type, date_key)
);

CREATE INDEX idx_klines_statistics_date_key ON klines_statistics(date_key);
CREATE INDEX idx_klines_statistics_symbol ON klines_statistics(symbol);

-- 5. 建立視圖便於查詢所有時間間隔的K線
CREATE VIEW v_all_klines AS
SELECT symbol, '1m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1m WHERE is_closed = 1
UNION ALL
SELECT symbol, '5m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_5m WHERE is_closed = 1
UNION ALL
SELECT symbol, '15m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_15m WHERE is_closed = 1
UNION ALL
SELECT symbol, '30m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_30m WHERE is_closed = 1
UNION ALL
SELECT symbol, '1h' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1h WHERE is_closed = 1
UNION ALL
SELECT symbol, '4h' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_4h WHERE is_closed = 1
UNION ALL
SELECT symbol, '1d' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1d WHERE is_closed = 1
UNION ALL
SELECT symbol, '1w' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1w WHERE is_closed = 1
UNION ALL
SELECT symbol, '1M' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1M WHERE is_closed = 1;

-- 6. 觸發器（自動更新updated_at欄位，SQLite版本）
-- 為trading_pairs表建立觸發器
CREATE TRIGGER trigger_trading_pairs_updated_at
    AFTER UPDATE ON trading_pairs
    FOR EACH ROW
BEGIN
    UPDATE trading_pairs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 為klines_1m表建立觸發器
CREATE TRIGGER trigger_klines_1m_updated_at
    AFTER UPDATE ON klines_1m
    FOR EACH ROW
BEGIN
    UPDATE klines_1m SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 為klines_statistics表建立觸發器
CREATE TRIGGER trigger_klines_statistics_updated_at
    AFTER UPDATE ON klines_statistics
    FOR EACH ROW
BEGIN
    UPDATE klines_statistics SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 7. 示例插入資料
-- 插入交易對範例
INSERT INTO trading_pairs (symbol, base_asset, quote_asset, status, price_precision, quantity_precision, min_notional, is_active)
VALUES
    ('BTCUSDT', 'BTC', 'USDT', 'TRADING', 2, 6, 10.0, 1),
    ('ETHUSDT', 'ETH', 'USDT', 'TRADING', 2, 5, 10.0, 1),
    ('ADAUSDT', 'ADA', 'USDT', 'TRADING', 4, 1, 10.0, 1);

-- 插入1分鐘K線範例資料
INSERT INTO klines_1m (symbol, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count, is_closed)
VALUES
    ('BTCUSDT', 1640995200000, 1640995259999, 47000.50, 47100.00, 46950.00, 47050.25, 1.25, 58812.81, 45, 1),
    ('ETHUSDT', 1640995200000, 1640995259999, 3800.25, 3820.50, 3795.00, 3815.75, 15.50, 59044.12, 28, 1);

-- 8. 實用查詢範例
-- 查詢特定交易對的最新K線資料
-- SELECT * FROM klines_1m WHERE symbol = 'BTCUSDT' ORDER BY open_time DESC LIMIT 100;

-- 查詢某個時間範圍內的K線資料
-- SELECT * FROM klines_1h WHERE symbol = 'BTCUSDT' AND open_time >= 1640995200000 AND close_time <= 1641081600000;

-- 計算移動平均線（例如：20期簡單移動平均）
-- SELECT symbol, open_time, close_price,
--        AVG(close_price) OVER (PARTITION BY symbol ORDER BY open_time ROWS 19 PRECEDING) as sma_20
-- FROM klines_1h WHERE symbol = 'BTCUSDT' ORDER BY open_time DESC LIMIT 100;