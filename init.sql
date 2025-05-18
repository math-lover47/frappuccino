-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUM Types
CREATE TYPE all_order_status AS ENUM ('PENDING', 'COMPLETED', 'CANCELLED');
CREATE TYPE all_order_payment_method AS ENUM ('CASH', 'CARD');
CREATE TYPE all_inventory_transaction_action AS ENUM ('ADD', 'REMOVE', 'ADJUST');

-- Tables
CREATE TABLE customers (
    customer_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    email VARCHAR(255) NOT NULL,
    preferences JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE menu_items (
    menu_item_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_name VARCHAR(255) NOT NULL DEFAULT '',
    item_description TEXT NOT NULL DEFAULT '',
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    categories TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE inventory (
    ingredient_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ingredient_name VARCHAR(255) NOT NULL UNIQUE,
    unit VARCHAR(15) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity >= 0),
    reorder_level DECIMAL(10,2) NOT NULL CHECK (reorder_level >= 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()  
);

CREATE TABLE orders (
    order_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(customer_id) ON DELETE RESTRICT,
    special_instructions JSONB NOT NULL DEFAULT '{}'::JSONB,
    total_price DECIMAL(10,2) NOT NULL CHECK (total_price >= 0),
    order_status all_order_status NOT NULL DEFAULT 'PENDING',
    order_payment_method all_order_payment_method NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE order_status_history (
    order_status_history_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(order_id) ON DELETE CASCADE,
    order_status all_order_status NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    order_item_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id) ON DELETE RESTRICT,
    customizations JSONB NOT NULL DEFAULT '{}'::JSONB,
    item_name VARCHAR(255) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity >= 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
);

CREATE TABLE price_history (
    price_history_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE menu_item_ingredients (
    menu_item_ingredients_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES inventory(ingredient_id) ON DELETE RESTRICT,
    ingredient_name VARCHAR(255) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity > 0),
    UNIQUE(menu_item_id, ingredient_id)
);

CREATE TABLE inventory_transactions (
    inventory_transactions_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ingredient_id UUID REFERENCES inventory(ingredient_id) ON DELETE RESTRICT NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    inventory_transaction_action all_inventory_transaction_action NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);    

-- Indexes for order_items table
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_menu_item_id ON order_items(menu_item_id);

-- Indexes for menu_items table
CREATE INDEX idx_menu_items_categories ON menu_items USING GIN(categories);
CREATE INDEX idx_menu_items_price ON menu_items(price);

-- Indexes for price_history table
CREATE INDEX idx_price_history_menu_item_id ON price_history(menu_item_id);
CREATE INDEX idx_price_history_updated_at ON price_history(updated_at);

-- Indexes for inventory table
CREATE INDEX idx_inventory_ingredient_name ON inventory(ingredient_name);
CREATE INDEX idx_inventory_reorder_level ON inventory(reorder_level);

-- Indexes for menu_item_ingredients table
CREATE INDEX idx_menu_item_ingredients_menu_item_id ON menu_item_ingredients(menu_item_id);
CREATE INDEX idx_menu_item_ingredients_ingredient_id ON menu_item_ingredients(ingredient_id);

-- Indexes for customer table
CREATE INDEX idx_customers_full_name ON customers(full_name);
CREATE INDEX idx_customers_email ON customers(email);

-- Indexes for inventory_transactions table
CREATE INDEX idx_inventory_transactions_ingredient_id ON inventory_transactions(ingredient_id);
CREATE INDEX idx_inventory_transactions_created_at ON inventory_transactions(created_at);
CREATE INDEX idx_inventory_transactions_action ON inventory_transactions(inventory_transaction_action);

-- Indexes for orders table
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_order_status ON orders(order_status);
CREATE INDEX idx_orders_payment_method ON orders(order_payment_method); 

-- Indexes for order_status_history table
CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_order_status ON order_status_history(order_status);
CREATE INDEX idx_order_status_history_updated_at ON order_status_history(updated_at);

CREATE INDEX idx_price_history_effective_dates ON price_history (effective_from, effective_to);

-- Trigger for updated_at on tables
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_customers_timestamp
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_menu_items_timestamp
    BEFORE UPDATE ON menu_items
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_inventory_timestamp
    BEFORE UPDATE ON inventory
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_orders_timestamp
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- Automatically log status changes to order_status_history
CREATE OR REPLACE FUNCTION log_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.order_status <> OLD.order_status THEN
        INSERT INTO order_status_history (order_id, order_status, notes)
        VALUES (NEW.order_id, NEW.order_status, 'Status changed from ' || OLD.order_status || ' to ' || NEW.order_status);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_order_status_change
AFTER UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION log_order_status_change();

-- Function to update inventory when order is completed
CREATE OR REPLACE FUNCTION update_inventory_on_order_complete()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.order_status = 'COMPLETE' AND OLD.order_status <> 'COMPLETE' THEN
        -- Reduce inventory for each ingredient used in this order
        INSERT INTO inventory_transactions (ingredient_id, quantity, inventory_transaction_action, reference_id, notes)
        SELECT mii.ingredient_id, -1 * (mii.quantity * oi.quantity), 'REMOVE', NEW.order_id, 'Automatic deduction for order ' || NEW.order_id
        FROM order_items oi
        JOIN menu_item_ingredients mii ON oi.menu_item_id = mii.menu_item_id
        WHERE oi.order_id = NEW.order_id;
        
        -- Update inventory quantities
        UPDATE inventory i
        SET quantity = i.quantity - subquery.total_quantity
        FROM (
            SELECT mii.ingredient_id, SUM(mii.quantity * oi.quantity) as total_quantity
            FROM order_items oi
            JOIN menu_item_ingredients mii ON oi.menu_item_id = mii.menu_item_id
            WHERE oi.order_id = NEW.order_id
            GROUP BY mii.ingredient_id
        ) as subquery
        WHERE i.ingredient_id = subquery.ingredient_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_inventory_on_order_complete
AFTER UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION update_inventory_on_order_complete();

-- Function to track price changes in price_history (consolidated version)
CREATE OR REPLACE FUNCTION track_menu_item_price_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Insert initial price history record for new menu items
        INSERT INTO price_history (menu_item_id, price, effective_from)
        VALUES (NEW.menu_item_id, NEW.price, now());
    ELSIF TG_OP = 'UPDATE' AND NEW.price <> OLD.price THEN
        -- Close previous price period
        UPDATE price_history
        SET effective_to = now()
        WHERE menu_item_id = NEW.menu_item_id 
          AND effective_to IS NULL;
        
        -- Insert new price with NULL effective_to
        INSERT INTO price_history (menu_item_id, price, effective_from)
        VALUES (NEW.menu_item_id, NEW.price, now());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for price history tracking
CREATE TRIGGER trigger_track_menu_item_price_change
AFTER UPDATE ON menu_items
FOR EACH ROW EXECUTE FUNCTION track_menu_item_price_change();

CREATE TRIGGER trigger_create_initial_price_history
AFTER INSERT ON menu_items
FOR EACH ROW EXECUTE FUNCTION track_menu_item_price_change();

-- Function to check inventory levels and alert on low stock
CREATE OR REPLACE FUNCTION check_inventory_levels()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.quantity <= NEW.reorder_level AND OLD.quantity > OLD.reorder_level THEN
        -- In a real system, this would send alerts
        RAISE NOTICE 'Inventory for % is low (% %)', NEW.ingredient_name, NEW.quantity, NEW.unit;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_inventory_levels
AFTER UPDATE ON inventory
FOR EACH ROW EXECUTE FUNCTION check_inventory_levels();

-- Function to update order total price
CREATE OR REPLACE FUNCTION update_order_total_price()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE orders
    SET total_price = (
        SELECT COALESCE(SUM(total_price), 0)
        FROM order_items
        WHERE order_id = 
            CASE 
              WHEN TG_OP = 'DELETE' THEN OLD.order_id
              ELSE NEW.order_id
            END
    )
    WHERE order_id = 
        CASE 
          WHEN TG_OP = 'DELETE' THEN OLD.order_id
          ELSE NEW.order_id
        END;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_order_total_price
AFTER INSERT OR UPDATE OR DELETE ON order_items
FOR EACH ROW EXECUTE FUNCTION update_order_total_price();

