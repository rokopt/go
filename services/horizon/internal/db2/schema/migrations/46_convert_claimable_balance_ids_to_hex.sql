-- +migrate Up

UPDATE claimable_balances SET id = encode(decode(id, 'base64'), 'hex');
UPDATE history_claimable_balances SET claimable_balance_id = encode(decode(claimable_balance_id, 'base64'), 'hex');

-- +migrate Down

UPDATE claimable_balances SET id = encode(decode(id, 'hex'), 'base64');
UPDATE history_claimable_balances SET claimable_balance_id = encode(decode(claimable_balance_id, 'hex'), 'base64');

