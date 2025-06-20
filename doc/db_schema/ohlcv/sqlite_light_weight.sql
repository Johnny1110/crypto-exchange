-- =============================================
-- Crypto ohlcv database SQLite version
-- =============================================

-- trading_pairs basic info
DROP TABLE IF EXISTS trading_pairs;
CREATE TABLE trading_pairs (
                               id INTEGER PRIMARY KEY AUTOINCREMENT,
                               symbol TEXT NOT NULL UNIQUE, -- trading-pair(symbol) ex: BTC-USDT
                               base_asset TEXT NOT NULL, -- ex: BTC
                               quote_asset TEXT NOT NULL, -- ex: USDT
                               status TEXT DEFAULT 'TRADING' CHECK (status IN ('TRADING', 'HALT', 'BREAK')),
                               price_precision INTEGER DEFAULT 7, -- price precision
                               size_precision INTEGER DEFAULT 7, -- qty precision
                               min_notional REAL DEFAULT 0,
                               is_active INTEGER DEFAULT 1 CHECK (is_active IN (0,1)),
                               created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                               updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- trading_pairs index
CREATE INDEX idx_trading_pairs_symbol ON trading_pairs(symbol);
CREATE INDEX idx_trading_pairs_assets ON trading_pairs(base_asset, quote_asset);
CREATE INDEX idx_trading_pairs_status ON trading_pairs(status, is_active);


-- =============================================
-- ohlcv main tables (Splitted by time range)
-- =============================================
-- 15 min ohlcv
DROP TABLE IF EXISTS ohlcv_15min;
CREATE TABLE ohlcv_15min (
                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                          symbol TEXT NOT NULL,
                          open_price REAL NOT NULL, -- o
                          high_price REAL NOT NULL, -- h
                          low_price REAL NOT NULL, -- l
                          close_price REAL NOT NULL, -- c
                          volume REAL NOT NULL DEFAULT 0, -- v
                          quote_volume REAL NOT NULL DEFAULT 0,
                          open_time INTEGER NOT NULL,
                          close_time INTEGER NOT NULL,
                          trade_count INTEGER NOT NULL DEFAULT 0,
                          is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          UNIQUE(symbol, open_time)
);

CREATE INDEX idx_ohlcv_15min_symbol_time ON ohlcv_15min(symbol, open_time);
CREATE INDEX idx_ohlcv_15min_symbol_time_range ON ohlcv_15min(symbol, open_time, close_time);
CREATE INDEX idx_ohlcv_15min_close_time ON ohlcv_15min(close_time);

-- 1 hr ohlcv
DROP TABLE IF EXISTS ohlcv_1h;
CREATE TABLE ohlcv_1h (
                           id INTEGER PRIMARY KEY AUTOINCREMENT,
                           symbol TEXT NOT NULL,
                           open_price REAL NOT NULL, -- o
                           high_price REAL NOT NULL, -- h
                           low_price REAL NOT NULL, -- l
                           close_price REAL NOT NULL, -- c
                           volume REAL NOT NULL DEFAULT 0, -- v
                           quote_volume REAL NOT NULL DEFAULT 0,
                           open_time INTEGER NOT NULL,
                           close_time INTEGER NOT NULL,
                           trade_count INTEGER NOT NULL DEFAULT 0,
                           is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                           created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                           UNIQUE(symbol, open_time)
);

CREATE INDEX idx_ohlcv_1h_symbol_time ON ohlcv_1h(symbol, open_time);
CREATE INDEX idx_ohlcv_1h_symbol_time_range ON ohlcv_1h(symbol, open_time, close_time);
CREATE INDEX idx_ohlcv_1h_close_time ON ohlcv_1h(close_time);

-- 1 day ohlcv
DROP TABLE IF EXISTS ohlcv_1d;
CREATE TABLE ohlcv_1d (
                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                          symbol TEXT NOT NULL,
                          open_price REAL NOT NULL, -- o
                          high_price REAL NOT NULL, -- h
                          low_price REAL NOT NULL, -- l
                          close_price REAL NOT NULL, -- c
                          volume REAL NOT NULL DEFAULT 0, -- v
                          quote_volume REAL NOT NULL DEFAULT 0,
                          open_time INTEGER NOT NULL,
                          close_time INTEGER NOT NULL,
                          trade_count INTEGER NOT NULL DEFAULT 0,
                          is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          UNIQUE(symbol, open_time)
);

CREATE INDEX idx_ohlcv_1d_symbol_time ON ohlcv_1d(symbol, open_time);
CREATE INDEX idx_ohlcv_1d_symbol_time_range ON ohlcv_1d(symbol, open_time, close_time);
CREATE INDEX idx_ohlcv_1dclose_time ON ohlcv_1d(close_time);

-- 1 week ohlcv
DROP TABLE IF EXISTS ohlcv_1w;
CREATE TABLE ohlcv_1w (
                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                          symbol TEXT NOT NULL,
                          open_price REAL NOT NULL, -- o
                          high_price REAL NOT NULL, -- h
                          low_price REAL NOT NULL, -- l
                          close_price REAL NOT NULL, -- c
                          volume REAL NOT NULL DEFAULT 0, -- v
                          quote_volume REAL NOT NULL DEFAULT 0,
                          open_time INTEGER NOT NULL,
                          close_time INTEGER NOT NULL,
                          trade_count INTEGER NOT NULL DEFAULT 0,
                          is_closed INTEGER DEFAULT 0 CHECK (is_closed IN (0,1)),
                          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          UNIQUE(symbol, open_time)
);

CREATE INDEX idx_ohlcv_1w_symbol_time ON ohlcv_1w(symbol, open_time);
CREATE INDEX idx_ohlcv_1w_symbol_time_range ON ohlcv_1w(symbol, open_time, close_time);
CREATE INDEX idx_ohlcv_1w_close_time ON ohlcv_1w(close_time);

-- 3. ohlcv real time cache（unclosed ohlcv candle）
DROP TABLE IF EXISTS ohlcv_realtime;
CREATE TABLE ohlcv_realtime (
                                 id INTEGER PRIMARY KEY AUTOINCREMENT,
                                 symbol TEXT NOT NULL,
                                 interval_type TEXT NOT NULL, --：1h,1d,1w ...
                                 open_price REAL NOT NULL, -- o
                                 high_price REAL NOT NULL, -- h
                                 low_price REAL NOT NULL, -- l
                                 close_price REAL NOT NULL, -- c
                                 volume REAL NOT NULL DEFAULT 0, -- v
                                 quote_volume REAL NOT NULL DEFAULT 0,
                                 open_time INTEGER NOT NULL,
                                 close_time INTEGER NOT NULL,
                                 trade_count INTEGER NOT NULL DEFAULT 0,
                                 last_trade_id INTEGER, -- last trade ID
                                 updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                 UNIQUE(symbol, interval_type)
);

CREATE INDEX idx_ohlcv_realtime_symbol_interval ON ohlcv_realtime(symbol, interval_type);
CREATE INDEX idx_ohlcv_realtime_updated_at ON ohlcv_realtime(updated_at);

-- 4. optimize query
DROP TABLE IF EXISTS ohlcv_statistics;
CREATE TABLE ohlcv_statistics (
                                   id INTEGER PRIMARY KEY AUTOINCREMENT,
                                   symbol TEXT NOT NULL,
                                   interval_type TEXT NOT NULL, --：1h,1d,1w ...
                                   date_key DATE NOT NULL,
                                   record_count INTEGER NOT NULL DEFAULT 0,
                                   min_open_time INTEGER,
                                   max_close_time INTEGER,
                                   avg_volume REAL, -- avg trade volume
                                   total_volume REAL, -- total trade volume
                                   created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                   updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                   UNIQUE(symbol, interval_type, date_key)
);

CREATE INDEX idx_ohlcv_statistics_date_key ON ohlcv_statistics(date_key);
CREATE INDEX idx_ohlcv_statistics_symbol ON ohlcv_statistics(symbol);

-- 5. create view for query all time interval_type ohlcv.
DROP VIEW IF EXISTS v_all_ohlcv;
CREATE VIEW v_all_ohlcv AS
SELECT symbol, '15min' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM ohlcv_15min WHERE is_closed = 1
UNION ALL
SELECT symbol, '1h' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM ohlcv_1h WHERE is_closed = 1
UNION ALL
SELECT symbol, '1d' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM ohlcv_1d WHERE is_closed = 1
UNION ALL
SELECT symbol, '1w' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM ohlcv_1w WHERE is_closed = 1;
