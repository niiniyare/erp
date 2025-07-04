-- Down migration for INVENTORY MODULE

-- Drop inventory_movements table
DROP TABLE IF EXISTS inventory_movements;

-- Drop inventory_balances table
DROP TABLE IF EXISTS inventory_balances;

-- Drop items table
DROP TABLE IF EXISTS items;

-- Drop item_categories table
DROP TABLE IF EXISTS item_categories;

-- Drop warehouses table
DROP TABLE IF EXISTS warehouses;

