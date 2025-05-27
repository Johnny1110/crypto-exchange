delete from orders where TRUE;
delete from trades where TRUE;
update balances set available = 0, locked = 0 where TRUE;
update balances set available = 10000000, locked = 0 where user_id = '8c26c994-af9e-4ef2-8f09-2bf48b1a1b83' and asset = 'USDT';
update balances set available = 1000, locked = 0 where user_id = '8c26c994-af9e-4ef2-8f09-2bf48b1a1b83' and asset = 'ETH';
