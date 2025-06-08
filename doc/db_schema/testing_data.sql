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
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('UID25060650F50001', 'shiqi', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 1, 0.001, 0.002);
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('UID25060650QA0001', 'kai_btc', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 1, 0.001, 0.002);
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('UID25060650QA0002', 'kai_eth', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 1, 0.001, 0.002);
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('UID25060650QA0003', 'kai_dot', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 1, 0.001, 0.002);

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

-- Create Balances for johnny Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('UID25060650F57788', 'USDT', 500000, 0),
       ('UID25060650F57788', 'BTC', 10, 0),
       ('UID25060650F57788', 'ETH', 10, 0),
       ('UID25060650F57788', 'DOT', 100, 0);

-- Create Balances for shiqi Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('UID25060650F50001', 'USDT', 500000, 0),
       ('UID25060650F50001', 'BTC', 10, 0),
       ('UID25060650F50001', 'ETH', 10, 0),
       ('UID25060650F50001', 'DOT', 100, 0);

-- Create Balances for kai Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('UID25060650QA0001', 'USDT', 1000000, 0),
       ('UID25060650QA0001', 'BTC', 10, 0),
       ('UID25060650QA0001', 'ETH', 1000, 0),
       ('UID25060650QA0001', 'DOT', 10000, 0);

INSERT INTO balances(user_id,asset,available,locked)
VALUES ('UID25060650QA0002', 'USDT', 1000000, 0),
       ('UID25060650QA0002', 'BTC', 10, 0),
       ('UID25060650QA0002', 'ETH', 1000, 0),
       ('UID25060650QA0002', 'DOT', 10000, 0);

INSERT INTO balances(user_id,asset,available,locked)
VALUES ('UID25060650QA0003', 'USDT', 1000000, 0),
       ('UID25060650QA0003', 'BTC', 10, 0),
       ('UID25060650QA0003', 'ETH', 1000, 0),
       ('UID25060650QA0003', 'DOT', 10000, 0);

