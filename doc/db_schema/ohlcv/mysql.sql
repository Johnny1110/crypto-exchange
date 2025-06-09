-- =============================================
-- K線資料庫完整表結構設計
-- 適用於加密貨幣現貨交易所
-- =============================================

-- 1. 交易對基礎資訊表
CREATE TABLE trading_pairs (
                               id BIGINT PRIMARY KEY AUTO_INCREMENT,
                               symbol VARCHAR(20) NOT NULL UNIQUE COMMENT '交易對符號，如BTCUSDT',
                               base_asset VARCHAR(10) NOT NULL COMMENT '基礎資產，如BTC',
                               quote_asset VARCHAR(10) NOT NULL COMMENT '計價資產，如USDT',
                               status ENUM('TRADING', 'HALT', 'BREAK') DEFAULT 'TRADING' COMMENT '交易狀態',
    price_precision INT DEFAULT 8 COMMENT '價格精度',
    quantity_precision INT DEFAULT 8 COMMENT '數量精度',
    min_notional DECIMAL(20,8) DEFAULT 0 COMMENT '最小名義價值',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否啟用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_symbol (symbol),
    INDEX idx_assets (base_asset, quote_asset),
    INDEX idx_status (status, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易對基礎資訊表';

-- 2. K線主資料表（分時間間隔存儲）
-- 建議為每個時間間隔創建獨立表，提升查詢性能

-- 1分鐘K線表
CREATE TABLE klines_1m (
                           id BIGINT PRIMARY KEY AUTO_INCREMENT,
                           symbol VARCHAR(20) NOT NULL COMMENT '交易對符號',
                           open_time BIGINT NOT NULL COMMENT '開盤時間（毫秒時間戳）',
                           close_time BIGINT NOT NULL COMMENT '收盤時間（毫秒時間戳）',
                           open_price DECIMAL(20,8) NOT NULL COMMENT '開盤價',
                           high_price DECIMAL(20,8) NOT NULL COMMENT '最高價',
                           low_price DECIMAL(20,8) NOT NULL COMMENT '最低價',
                           close_price DECIMAL(20,8) NOT NULL COMMENT '收盤價',
                           volume DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '成交量（基礎資產）',
                           quote_volume DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '成交額（計價資產）',
                           trade_count INT NOT NULL DEFAULT 0 COMMENT '成交筆數',
                           taker_buy_volume DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '主動買入成交量',
                           taker_buy_quote_volume DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '主動買入成交額',
                           is_closed BOOLEAN DEFAULT FALSE COMMENT '該K線是否已完結',
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                           UNIQUE KEY uk_symbol_time (symbol, open_time),
                           INDEX idx_symbol_time_range (symbol, open_time, close_time),
    INDEX idx_close_time (close_time),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
COMMENT='1分鐘K線資料表'
PARTITION BY RANGE (open_time) (
    PARTITION p202401 VALUES LESS THAN (1706745600000), -- 2024-02-01
    PARTITION p202402 VALUES LESS THAN (1709251200000), -- 2024-03-01
    PARTITION p202403 VALUES LESS THAN (1711929600000), -- 2024-04-01
    PARTITION p202404 VALUES LESS THAN (1714521600000), -- 2024-05-01
    PARTITION p202405 VALUES LESS THAN (1717200000000), -- 2024-06-01
    PARTITION p202406 VALUES LESS THAN (1719792000000), -- 2024-07-01
    -- 可根據需要繼續添加分區
    PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 5分鐘K線表
CREATE TABLE klines_5m (
                           id BIGINT PRIMARY KEY AUTO_INCREMENT,
                           symbol VARCHAR(20) NOT NULL,
                           open_time BIGINT NOT NULL,
                           close_time BIGINT NOT NULL,
                           open_price DECIMAL(20,8) NOT NULL,
                           high_price DECIMAL(20,8) NOT NULL,
                           low_price DECIMAL(20,8) NOT NULL,
                           close_price DECIMAL(20,8) NOT NULL,
                           volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                           quote_volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                           trade_count INT NOT NULL DEFAULT 0,
                           taker_buy_volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                           taker_buy_quote_volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                           is_closed BOOLEAN DEFAULT FALSE,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                           UNIQUE KEY uk_symbol_time (symbol, open_time),
                           INDEX idx_symbol_time_range (symbol, open_time, close_time),
    INDEX idx_close_time (close_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='5分鐘K線資料表';

-- 15分鐘K線表
CREATE TABLE klines_15m LIKE klines_5m;
ALTER TABLE klines_15m COMMENT='15分鐘K線資料表';

-- 30分鐘K線表
CREATE TABLE klines_30m LIKE klines_5m;
ALTER TABLE klines_30m COMMENT='30分鐘K線資料表';

-- 1小時K線表
CREATE TABLE klines_1h LIKE klines_5m;
ALTER TABLE klines_1h COMMENT='1小時K線資料表';

-- 4小時K線表
CREATE TABLE klines_4h LIKE klines_5m;
ALTER TABLE klines_4h COMMENT='4小時K線資料表';

-- 1天K線表
CREATE TABLE klines_1d LIKE klines_5m;
ALTER TABLE klines_1d COMMENT='1天K線資料表';

-- 1週K線表
CREATE TABLE klines_1w LIKE klines_5m;
ALTER TABLE klines_1w COMMENT='1週K線資料表';

-- 1月K線表
CREATE TABLE klines_1M LIKE klines_5m;
ALTER TABLE klines_1M COMMENT='1月K線資料表';

-- 3. 實時K線緩存表（用於當前未完結的K線）
CREATE TABLE klines_realtime (
                                 id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                 symbol VARCHAR(20) NOT NULL,
                                 interval_type VARCHAR(10) NOT NULL COMMENT '時間間隔類型：1m,5m,15m,30m,1h,4h,1d,1w,1M',
                                 open_time BIGINT NOT NULL,
                                 close_time BIGINT NOT NULL,
                                 open_price DECIMAL(20,8) NOT NULL,
                                 high_price DECIMAL(20,8) NOT NULL,
                                 low_price DECIMAL(20,8) NOT NULL,
                                 close_price DECIMAL(20,8) NOT NULL,
                                 volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                                 quote_volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                                 trade_count INT NOT NULL DEFAULT 0,
                                 taker_buy_volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                                 taker_buy_quote_volume DECIMAL(20,8) NOT NULL DEFAULT 0,
                                 last_trade_id BIGINT COMMENT '最後一筆交易ID',
                                 updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                                 UNIQUE KEY uk_symbol_interval_time (symbol, interval_type, open_time),
                                 INDEX idx_symbol_interval (symbol, interval_type),
    INDEX idx_updated_at (updated_at)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='實時K線緩存表';

-- 4. K線資料統計表（用於優化查詢）
CREATE TABLE klines_statistics (
                                   id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                   symbol VARCHAR(20) NOT NULL,
                                   interval_type VARCHAR(10) NOT NULL,
                                   date_key DATE NOT NULL COMMENT '統計日期',
                                   record_count INT NOT NULL DEFAULT 0 COMMENT '當日K線記錄數',
                                   min_open_time BIGINT COMMENT '最小開盤時間',
                                   max_close_time BIGINT COMMENT '最大收盤時間',
                                   avg_volume DECIMAL(20,8) COMMENT '平均成交量',
                                   total_volume DECIMAL(20,8) COMMENT '總成交量',
                                   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                                   UNIQUE KEY uk_symbol_interval_date (symbol, interval_type, date_key),
                                   INDEX idx_date_key (date_key),
    INDEX idx_symbol (symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K線資料統計表';

-- 5. 建立視圖便於查詢所有時間間隔的K線
CREATE VIEW v_all_klines AS
SELECT symbol, '1m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1m WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '5m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_5m WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '15m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_15m WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '30m' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_30m WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '1h' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1h WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '4h' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_4h WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '1d' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1d WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '1w' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1w WHERE is_closed = TRUE
UNION ALL
SELECT symbol, '1M' as interval_type, open_time, close_time, open_price, high_price, low_price, close_price, volume, quote_volume, trade_count FROM klines_1M WHERE is_closed = TRUE;

-- =============================================
-- 常用查詢範例
-- =============================================

-- 查詢指定交易對最近的K線資料
-- SELECT * FROM klines_1m
-- WHERE symbol = 'BTCUSDT' AND open_time >= 1640995200000
-- ORDER BY open_time DESC LIMIT 100;

-- 查詢指定時間範圍的K線資料
-- SELECT * FROM klines_1h
-- WHERE symbol = 'BTCUSDT'
-- AND open_time >= 1640995200000
-- AND close_time <= 1641081600000
-- ORDER BY open_time ASC;

-- 透過視圖查詢多時間週期資料
-- SELECT * FROM v_all_klines
-- WHERE symbol = 'BTCUSDT' AND interval_type IN ('1m', '5m', '1h')
-- ORDER BY interval_type, open_time DESC;

-- =============================================
-- 索引優化建議
-- =============================================

-- 如果查詢頻繁，可考慮添加覆蓋索引
-- ALTER TABLE klines_1m ADD INDEX idx_symbol_time_ohlcv (symbol, open_time, open_price, high_price, low_price, close_price, volume);

-- 對於歷史資料查詢，可考慮創建分區索引
-- ALTER TABLE klines_1m ADD INDEX idx_partition_symbol_time (open_time, symbol) USING BTREE;