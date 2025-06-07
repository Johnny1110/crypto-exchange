delete from orders where TRUE;
delete from trades where TRUE;
update balances set available = 0, locked = 0 where TRUE;

-- Create Margin Account
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('0', 'margin_account', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 0, 0, 0);

-- Create Testing Maker Account
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('MID250606CXAZ1199', 'market_maker', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 7, 0.0001, 0.002);

-- Create Testing User Account
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('UID25060650F57788', 'johnny', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 1, 0.001, 0.002);

-- Create Balances for Margin Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('0', 'USDT', 0, 0),
       ('0', 'BTC', 0, 0),
       ('0', 'ETH', 0, 0),
       ('0', 'DOT', 0, 0);

-- Create Balances for Maker Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('MID250606CXAZ1199', 'USDT', 75000000, 0),
       ('MID250606CXAZ1199', 'BTC', 500, 0), -- BTC: 50000000 USDT
       ('MID250606CXAZ1199', 'ETH', 8000, 0), -- ETH: 20000000 USDT
       ('MID250606CXAZ1199', 'DOT', 500000, 0); -- DOT: 2000000 USDT

-- Create Balances for Maker Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('UID25060650F57788', 'USDT', 300000, 0),
       ('UID25060650F57788', 'BTC', 0, 0),
       ('UID25060650F57788', 'ETH', 10, 0),
       ('UID25060650F57788', 'DOT', 10, 0);

