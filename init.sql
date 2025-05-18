-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUM Types
CREATE TYPE all_order_status AS ENUM ('PENDING', 'COMPLETED', 'CANCELLED');
CREATE TYPE all_order_payment_method AS ENUM ('CASH', 'CARD');
CREATE TYPE all_inventory_transaction_action AS ENUM ('ADD', 'REMOVE', 'ADJUST');
CREATE TYPE all_unit AS ENUM ('KG', 'G', 'L', );

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
    notes TEXT NOT NULL DEFAULT '',
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
    total_price DECIMAL(10,2) GENERATED ALWAYS AS (quantity * unit_price) STORED
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

-- Mock data for Restaurant Management System

-- Add mock data for customers
INSERT INTO customers (customer_id, full_name, phone_number, email, preferences)
VALUES
  ('81a130d2-502f-4cf1-a376-63edeb000e9f', 'John Smith', '555-123-4567', 'john.smith@email.com', '{"allergies": ["nuts", "shellfish"], "favorite_cuisine": "Italian"}'::JSONB),
  ('94729f09-0caa-4182-9f7b-6894a91b5310', 'Jane Doe', '555-234-5678', 'jane.doe@email.com', '{"allergies": ["dairy"], "dietary": "vegetarian"}'::JSONB),
  ('b6bf5612-12fb-4e77-aaa6-34d8cf4a38d3', 'Robert Johnson', '555-345-6789', 'robert.johnson@email.com', '{"favorite_dish": "Steak", "preferred_seating": "window"}'::JSONB),
  ('c2f5478d-7c71-4a78-9ebd-b73438513578', 'Emily Williams', '555-456-7890', 'emily.williams@email.com', '{"allergies": ["gluten"], "dietary": "gluten-free"}'::JSONB),
  ('f7b33c69-a46c-4922-9cd5-2902696ef54e', 'Michael Brown', '555-567-8901', 'michael.brown@email.com', '{"preferred_seating": "outdoor", "favorite_cuisine": "Mexican"}'::JSONB),
  ('ee7bd2d1-61e3-4b95-a4e2-1bc6f842523b', 'Sarah Davis', '555-678-9012', 'sarah.davis@email.com', '{"favorite_dish": "Salmon", "dietary": "pescatarian"}'::JSONB),
  ('e4cdfd94-651a-42d9-9e9b-de277fda91c5', 'David Miller', '555-789-0123', 'david.miller@email.com', '{"allergies": ["eggs"], "preferred_payment": "card"}'::JSONB),
  ('24c7ddd2-7168-46a6-a32a-27cf9be9f23a', 'Jessica Wilson', '555-890-1234', 'jessica.wilson@email.com', '{"favorite_cuisine": "Japanese", "preferred_seating": "booth"}'::JSONB),
  ('3c1e3cad-e5a8-45b0-9d77-463a2c79db7a', 'Christopher Moore', '555-901-2345', 'christopher.moore@email.com', '{"allergies": ["soy"], "favorite_dish": "Burger"}'::JSONB),
  ('5d9a3a9a-7b8d-4a26-9c87-4d94e5ab15b0', 'Amanda Taylor', '555-012-3456', 'amanda.taylor@email.com', '{"dietary": "vegan", "preferred_payment": "cash"}'::JSONB);

-- Add mock data for menu_items
INSERT INTO menu_items (menu_item_id, item_name, item_description, price, categories)
VALUES
  ('c0a80121-0001-4000-8000-000000000001', 'Margherita Pizza', 'Classic pizza with tomato sauce, mozzarella, and fresh basil', 14.99, ARRAY['Italian', 'Pizza', 'Vegetarian']),
  ('c0a80121-0001-4000-8000-000000000002', 'Beef Burger', 'Juicy beef patty with lettuce, tomato, cheese, and special sauce', 12.99, ARRAY['American', 'Burgers', 'Meat']),
  ('c0a80121-0001-4000-8000-000000000003', 'Caesar Salad', 'Fresh romaine lettuce with Caesar dressing, croutons, and parmesan', 9.99, ARRAY['Salads', 'Starters']),
  ('c0a80121-0001-4000-8000-000000000004', 'Grilled Salmon', 'Fresh salmon fillet grilled with herbs and served with seasonal vegetables', 18.99, ARRAY['Seafood', 'Main Course']),
  ('c0a80121-0001-4000-8000-000000000005', 'Spaghetti Carbonara', 'Classic Italian pasta with egg, cheese, pancetta, and black pepper', 13.99, ARRAY['Italian', 'Pasta']),
  ('c0a80121-0001-4000-8000-000000000006', 'Mushroom Risotto', 'Creamy arborio rice with wild mushrooms and parmesan', 14.99, ARRAY['Italian', 'Vegetarian', 'Rice']),
  ('c0a80121-0001-4000-8000-000000000007', 'Tiramisu', 'Italian dessert with coffee-soaked ladyfingers and mascarpone cream', 7.99, ARRAY['Dessert', 'Italian']),
  ('c0a80121-0001-4000-8000-000000000008', 'French Fries', 'Crispy fried potatoes served with ketchup', 4.99, ARRAY['Sides', 'Vegetarian']),
  ('c0a80121-0001-4000-8000-000000000009', 'Chocolate Cake', 'Rich chocolate cake with ganache frosting', 6.99, ARRAY['Dessert', 'Sweet']),
  ('c0a80121-0001-4000-8000-000000000010', 'Iced Tea', 'Fresh brewed tea served with ice and lemon', 2.99, ARRAY['Beverages', 'Non-alcoholic']);

-- Add mock data for inventory
INSERT INTO inventory (ingredient_id, ingredient_name, unit, quantity, reorder_level)
VALUES
  ('d0a80121-0001-4000-8000-000000000001', 'Flour', 'kg', 50.00, 10.00),
  ('d0a80121-0001-4000-8000-000000000002', 'Mozzarella Cheese', 'kg', 15.00, 5.00),
  ('d0a80121-0001-4000-8000-000000000003', 'Tomatoes', 'kg', 25.00, 8.00),
  ('d0a80121-0001-4000-8000-000000000004', 'Beef', 'kg', 20.00, 5.00),
  ('d0a80121-0001-4000-8000-000000000005', 'Lettuce', 'kg', 10.00, 3.00),
  ('d0a80121-0001-4000-8000-000000000006', 'Salmon', 'kg', 12.00, 4.00),
  ('d0a80121-0001-4000-8000-000000000007', 'Pasta', 'kg', 30.00, 8.00),
  ('d0a80121-0001-4000-8000-000000000008', 'Rice', 'kg', 25.00, 7.00),
  ('d0a80121-0001-4000-8000-000000000009', 'Mushrooms', 'kg', 8.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000010', 'Eggs', 'unit', 120.00, 24.00),
  ('d0a80121-0001-4000-8000-000000000011', 'Potatoes', 'kg', 40.00, 10.00),
  ('d0a80121-0001-4000-8000-000000000012', 'Sugar', 'kg', 15.00, 5.00),
  ('d0a80121-0001-4000-8000-000000000013', 'Chocolate', 'kg', 7.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000014', 'Tea Leaves', 'kg', 5.00, 1.00),
  ('d0a80121-0001-4000-8000-000000000015', 'Parmesan Cheese', 'kg', 6.00, 2.00);

-- Add mock data for orders
INSERT INTO orders (order_id, customer_id, special_instructions, total_price, order_status, order_payment_method)
VALUES
  ('e0a80121-0001-4000-8000-000000000001', '81a130d2-502f-4cf1-a376-63edeb000e9f', '{"notes": "Extra napkins please"}'::JSONB, 27.98, 'COMPLETE', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000002', '94729f09-0caa-4182-9f7b-6894a91b5310', '{"notes": "No onions on burger"}'::JSONB, 17.98, 'COMPLETE', 'CASH'),
  ('e0a80121-0001-4000-8000-000000000003', 'b6bf5612-12fb-4e77-aaa6-34d8cf4a38d3', '{"notes": "Medium rare steak"}'::JSONB, 18.99, 'COMPLETE', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000004', 'c2f5478d-7c71-4a78-9ebd-b73438513578', '{"notes": "Gluten free option"}'::JSONB, 14.99, 'PENDING', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000005', 'f7b33c69-a46c-4922-9cd5-2902696ef54e', '{"notes": "Extra spicy"}'::JSONB, 22.98, 'PENDING', 'CASH'),
  ('e0a80121-0001-4000-8000-000000000006', 'ee7bd2d1-61e3-4b95-a4e2-1bc6f842523b', '{}'::JSONB, 18.99, 'CANCELLED', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000007', 'e4cdfd94-651a-42d9-9e9b-de277fda91c5', '{"notes": "Birthday celebration"}'::JSONB, 35.96, 'COMPLETE', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000008', '24c7ddd2-7168-46a6-a32a-27cf9be9f23a', '{"notes": "Allergy to nuts"}'::JSONB, 16.98, 'PENDING', 'CASH'),
  ('e0a80121-0001-4000-8000-000000000009', '3c1e3cad-e5a8-45b0-9d77-463a2c79db7a', '{"notes": "To go please"}'::JSONB, 12.99, 'COMPLETE', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000010', '5d9a3a9a-7b8d-4a26-9c87-4d94e5ab15b0', '{"notes": "Vegan options only"}'::JSONB, 14.99, 'PENDING', 'CASH');

-- Add mock data for order_items
INSERT INTO order_items (order_item_id, order_id, menu_item_id, customizations, item_name, quantity, unit_price)
VALUES
  ('f0a80121-0001-4000-8000-000000000001', 'e0a80121-0001-4000-8000-000000000001', 'c0a80121-0001-4000-8000-000000000001', '{"extra_cheese": true}'::JSONB, 'Margherita Pizza', 1.00, 14.99),
  ('f0a80121-0001-4000-8000-000000000002', 'e0a80121-0001-4000-8000-000000000001', 'c0a80121-0001-4000-8000-000000000008', '{}'::JSONB, 'French Fries', 1.00, 4.99),
  ('f0a80121-0001-4000-8000-000000000003', 'e0a80121-0001-4000-8000-000000000001', 'c0a80121-0001-4000-8000-000000000010', '{}'::JSONB, 'Iced Tea', 1.00, 2.99),
  ('f0a80121-0001-4000-8000-000000000004', 'e0a80121-0001-4000-8000-000000000002', 'c0a80121-0001-4000-8000-000000000002', '{"no_onions": true}'::JSONB, 'Beef Burger', 1.00, 12.99),
  ('f0a80121-0001-4000-8000-000000000005', 'e0a80121-0001-4000-8000-000000000002', 'c0a80121-0001-4000-8000-000000000010', '{}'::JSONB, 'Iced Tea', 1.00, 2.99),
  ('f0a80121-0001-4000-8000-000000000006', 'e0a80121-0001-4000-8000-000000000003', 'c0a80121-0001-4000-8000-000000000004', '{"medium_rare": true}'::JSONB, 'Grilled Salmon', 1.00, 18.99),
  ('f0a80121-0001-4000-8000-000000000007', 'e0a80121-0001-4000-8000-000000000004', 'c0a80121-0001-4000-8000-000000000006', '{"gluten_free": true}'::JSONB, 'Mushroom Risotto', 1.00, 14.99),
  ('f0a80121-0001-4000-8000-000000000008', 'e0a80121-0001-4000-8000-000000000005', 'c0a80121-0001-4000-8000-000000000002', '{"extra_spicy": true}'::JSONB, 'Beef Burger', 1.00, 12.99),
  ('f0a80121-0001-4000-8000-000000000009', 'e0a80121-0001-4000-8000-000000000005', 'c0a80121-0001-4000-8000-000000000008', '{}'::JSONB, 'French Fries', 2.00, 4.99),
  ('f0a80121-0001-4000-8000-000000000010', 'e0a80121-0001-4000-8000-000000000006', 'c0a80121-0001-4000-8000-000000000004', '{}'::JSONB, 'Grilled Salmon', 1.00, 18.99),
  ('f0a80121-0001-4000-8000-000000000011', 'e0a80121-0001-4000-8000-000000000007', 'c0a80121-0001-4000-8000-000000000005', '{}'::JSONB, 'Spaghetti Carbonara', 1.00, 13.99),
  ('f0a80121-0001-4000-8000-000000000012', 'e0a80121-0001-4000-8000-000000000007', 'c0a80121-0001-4000-8000-000000000007', '{"birthday_message": true}'::JSONB, 'Tiramisu', 2.00, 7.99),
  ('f0a80121-0001-4000-8000-000000000013', 'e0a80121-0001-4000-8000-000000000007', 'c0a80121-0001-4000-8000-000000000010', '{}'::JSONB, 'Iced Tea', 2.00, 2.99),
  ('f0a80121-0001-4000-8000-000000000014', 'e0a80121-0001-4000-8000-000000000008', 'c0a80121-0001-4000-8000-000000000003', '{"no_nuts": true}'::JSONB, 'Caesar Salad', 1.00, 9.99),
  ('f0a80121-0001-4000-8000-000000000015', 'e0a80121-0001-4000-8000-000000000008', 'c0a80121-0001-4000-8000-000000000010', '{}'::JSONB, 'Iced Tea', 1.00, 2.99),
  ('f0a80121-0001-4000-8000-000000000016', 'e0a80121-0001-4000-8000-000000000009', 'c0a80121-0001-4000-8000-000000000002', '{"to_go": true}'::JSONB, 'Beef Burger', 1.00, 12.99),
  ('f0a80121-0001-4000-8000-000000000017', 'e0a80121-0001-4000-8000-000000000010', 'c0a80121-0001-4000-8000-000000000006', '{"vegan_option": true}'::JSONB, 'Mushroom Risotto', 1.00, 14.99);

-- Add mock data for order_status_history
INSERT INTO order_status_history (order_status_history_id, order_id, notes, order_status)
VALUES
  ('g0a80121-0001-4000-8000-000000000001', 'e0a80121-0001-4000-8000-000000000001', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000002', 'e0a80121-0001-4000-8000-000000000001', 'Order completed and delivered', 'COMPLETE'),
  ('g0a80121-0001-4000-8000-000000000003', 'e0a80121-0001-4000-8000-000000000002', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000004', 'e0a80121-0001-4000-8000-000000000002', 'Order completed and delivered', 'COMPLETE'),
  ('g0a80121-0001-4000-8000-000000000005', 'e0a80121-0001-4000-8000-000000000003', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000006', 'e0a80121-0001-4000-8000-000000000003', 'Order completed and delivered', 'COMPLETE'),
  ('g0a80121-0001-4000-8000-000000000007', 'e0a80121-0001-4000-8000-000000000004', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000008', 'e0a80121-0001-4000-8000-000000000005', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000009', 'e0a80121-0001-4000-8000-000000000006', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000010', 'e0a80121-0001-4000-8000-000000000006', 'Customer requested cancellation', 'CANCELLED'),
  ('g0a80121-0001-4000-8000-000000000011', 'e0a80121-0001-4000-8000-000000000007', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000012', 'e0a80121-0001-4000-8000-000000000007', 'Order completed and delivered', 'COMPLETE'),
  ('g0a80121-0001-4000-8000-000000000013', 'e0a80121-0001-4000-8000-000000000008', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000014', 'e0a80121-0001-4000-8000-000000000009', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000015', 'e0a80121-0001-4000-8000-000000000009', 'Order completed and delivered', 'COMPLETE'),
  ('g0a80121-0001-4000-8000-000000000016', 'e0a80121-0001-4000-8000-000000000010', 'Order received', 'PENDING');

-- Add mock data for price_history
INSERT INTO price_history (price_history_id, menu_item_id, price)
VALUES
  ('h0a80121-0001-4000-8000-000000000001', 'c0a80121-0001-4000-8000-000000000001', 12.99)
  ('h0a80121-0001-4000-8000-000000000002', 'c0a80121-0001-4000-8000-000000000001', 13.99)
  ('h0a80121-0001-4000-8000-000000000003', 'c0a80121-0001-4000-8000-000000000001', 14.99)
  ('h0a80121-0001-4000-8000-000000000004', 'c0a80121-0001-4000-8000-000000000002', 10.99)
  ('h0a80121-0001-4000-8000-000000000005', 'c0a80121-0001-4000-8000-000000000002', 12.99)
  ('h0a80121-0001-4000-8000-000000000006', 'c0a80121-0001-4000-8000-000000000003', 8.99)
  ('h0a80121-0001-4000-8000-000000000007', 'c0a80121-0001-4000-8000-000000000003', 9.99)
  ('h0a80121-0001-4000-8000-000000000008', 'c0a80121-0001-4000-8000-000000000004', 16.99)
  ('h0a80121-0001-4000-8000-000000000009', 'c0a80121-0001-4000-8000-000000000004', 18.99)
  ('h0a80121-0001-4000-8000-000000000010', 'c0a80121-0001-4000-8000-000000000005', 12.99)
  ('h0a80121-0001-4000-8000-000000000011', 'c0a80121-0001-4000-8000-000000000005', 13.99)

-- Add mock data for menu_item_ingredients
INSERT INTO menu_item_ingredients (menu_item_ingredients_id, menu_item_id, ingredient_id, ingredient_name, quantity)
VALUES
  ('i0a80121-0001-4000-8000-000000000001', 'c0a80121-0001-4000-8000-000000000001', 'd0a80121-0001-4000-8000-000000000001', 'Flour', 0.25),
  ('i0a80121-0001-4000-8000-000000000002', 'c0a80121-0001-4000-8000-000000000001', 'd0a80121-0001-4000-8000-000000000002', 'Mozzarella Cheese', 0.20),
  ('i0a80121-0001-4000-8000-000000000003', 'c0a80121-0001-4000-8000-000000000001', 'd0a80121-0001-4000-8000-000000000003', 'Tomatoes', 0.15),
  ('i0a80121-0001-4000-8000-000000000004', 'c0a80121-0001-4000-8000-000000000002', 'd0a80121-0001-4000-8000-000000000004', 'Beef', 0.20),
  ('i0a80121-0001-4000-8000-000000000005', 'c0a80121-0001-4000-8000-000000000002', 'd0a80121-0001-4000-8000-000000000005', 'Lettuce', 0.05),
  ('i0a80121-0001-4000-8000-000000000006', 'c0a80121-0001-4000-8000-000000000003', 'd0a80121-0001-4000-8000-000000000005', 'Lettuce', 0.15),
  ('i0a80121-0001-4000-8000-000000000007', 'c0a80121-0001-4000-8000-000000000003', 'd0a80121-0001-4000-8000-000000000015', 'Parmesan Cheese', 0.05),
  ('i0a80121-0001-4000-8000-000000000008', 'c0a80121-0001-4000-8000-000000000004', 'd0a80121-0001-4000-8000-000000000006', 'Salmon', 0.25),
  ('i0a80121-0001-4000-8000-000000000009', 'c0a80121-0001-4000-8000-000000000005', 'd0a80121-0001-4000-8000-000000000007', 'Pasta', 0.20),
  ('i0a80121-0001-4000-8000-000000000010', 'c0a80121-0001-4000-8000-000000000005', 'd0a80121-0001-4000-8000-000000000010', 'Eggs', 2.00),
  ('i0a80121-0001-4000-8000-000000000011', 'c0a80121-0001-4000-8000-000000000005', 'd0a80121-0001-4000-8000-000000000015', 'Parmesan Cheese', 0.10),
  ('i0a80121-0001-4000-8000-000000000012', 'c0a80121-0001-4000-8000-000000000006', 'd0a80121-0001-4000-8000-000000000008', 'Rice', 0.20),
  ('i0a80121-0001-4000-8000-000000000013', 'c0a80121-0001-4000-8000-000000000006', 'd0a80121-0001-4000-8000-000000000009', 'Mushrooms', 0.15),
  ('i0a80121-0001-4000-8000-000000000014', 'c0a80121-0001-4000-8000-000000000006', 'd0a80121-0001-4000-8000-000000000015', 'Parmesan Cheese', 0.08),
  ('i0a80121-0001-4000-8000-000000000015', 'c0a80121-0001-4000-8000-000000000007', 'd0a80121-0001-4000-8000-000000000010', 'Eggs', 2.00),
  ('i0a80121-0001-4000-8000-000000000016', 'c0a80121-0001-4000-8000-000000000008', 'd0a80121-0001-4000-8000-000000000011', 'Potatoes', 0.30),
  ('i0a80121-0001-4000-8000-000000000017', 'c0a80121-0001-4000-8000-000000000009', 'd0a80121-0001-4000-8000-000000000001', 'Flour', 0.15),
  ('i0a80121-0001-4000-8000-000000000018', 'c0a80121-0001-4000-8000-000000000009', 'd0a80121-0001-4000-8000-000000000012', 'Sugar', 0.10),
  ('i0a80121-0001-4000-8000-000000000019', 'c0a80121-0001-4000-8000-000000000009', 'd0a80121-0001-4000-8000-000000000013', 'Chocolate', 0.10),
  ('i0a80121-0001-4000-8000-000000000020', 'c0a80121-0001-4000-8000-000000000010', 'd0a80121-0001-4000-8000-0000000000114   ', 'Chocolate', 0.10),
  ('i0a80121-0001-4000-8000-000000000020', 'c0a80121-0001-4000-8000-000000000010', 'd0a80121-0001-4000-8000-000000000014', 'Tea Leaves', 0.05);

-- Add 2 new customers
INSERT INTO customers (customer_id, full_name, phone_number, email, preferences)
VALUES
  ('a1b2c3d4-502f-4cf1-a376-63edeb000e9f', 'Michael Johnson', '555-112-2233', 'michael.johnson@email.com', '{"allergies": ["peanuts"], "favorite_cuisine": "Mexican"}'::JSONB),
  ('e5f6g7h8-0caa-4182-9f7b-6894a91b5310', 'Sarah Wilson', '555-334-4556', 'sarah.wilson@email.com', '{"dietary": "vegetarian", "preferred_seating": "outdoor"}'::JSONB);

-- Add 4 new menu items
INSERT INTO menu_items (menu_item_id, item_name, item_description, price, categories)
VALUES
  ('c0a80121-0001-4000-8000-000000000011', 'Chicken Alfredo', 'Creamy Alfredo sauce with grilled chicken and fettuccine', 16.99, ARRAY['Italian', 'Pasta', 'Meat']),
  ('c0a80121-0001-4000-8000-000000000012', 'Veggie Burger', 'Plant-based burger with avocado and vegan mayo', 11.99, ARRAY['American', 'Burgers', 'Vegetarian', 'Vegan']),
  ('c0a80121-0001-4000-8000-000000000013', 'Greek Salad', 'Crispy vegetables with feta cheese and olives', 10.99, ARRAY['Salads', 'Vegetarian']),
  ('c0a80121-0001-4000-8000-000000000014', 'Cappuccino', 'Espresso with steamed milk and foam', 4.99, ARRAY['Beverages', 'Coffee']);

-- Add 8 new inventory ingredients
INSERT INTO inventory (ingredient_id, ingredient_name, unit, quantity, reorder_level)
VALUES
  ('d0a80121-0001-4000-8000-000000000016', 'Chicken Breast', 'kg', 15.00, 5.00),
  ('d0a80121-0001-4000-8000-000000000017', 'Heavy Cream', 'L', 10.00, 3.00),
  ('d0a80121-0001-4000-8000-000000000018', 'Chickpeas', 'kg', 8.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000019', 'Avocado', 'kg', 5.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000020', 'Feta Cheese', 'kg', 7.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000021', 'Olives', 'kg', 6.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000022', 'Coffee Beans', 'kg', 5.00, 2.00),
  ('d0a80121-0001-4000-8000-000000000023', 'Milk', 'L', 20.00, 5.00);

-- Add 2 new orders
INSERT INTO orders (order_id, customer_id, special_instructions, total_price, order_status, order_payment_method)
VALUES
  ('e0a80121-0001-4000-8000-000000000011', 'a1b2c3d4-502f-4cf1-a376-63edeb000e9f', '{"notes": "Extra sauce on the side"}'::JSONB, 0, 'PENDING', 'CARD'),
  ('e0a80121-0001-4000-8000-000000000012', 'e5f6g7h8-0caa-4182-9f7b-6894a91b5310', '{"notes": "No onions in salad"}'::JSONB, 0, 'COMPLETE', 'CASH');

-- Add order items for new orders
INSERT INTO order_items (order_item_id, order_id, menu_item_id, customizations, item_name, quantity, unit_price)
VALUES
  ('f0a80121-0001-4000-8000-000000000018', 'e0a80121-0001-4000-8000-000000000011', 'c0a80121-0001-4000-8000-000000000011', '{"extra_sauce": true}'::JSONB, 'Chicken Alfredo', 1, 16.99),
  ('f0a80121-0001-4000-8000-000000000019', 'e0a80121-0001-4000-8000-000000000011', 'c0a80121-0001-4000-8000-000000000014', '{"milk_type": "almond"}'::JSONB, 'Cappuccino', 2, 4.99),
  ('f0a80121-0001-4000-8000-000000000020', 'e0a80121-0001-4000-8000-000000000012', 'c0a80121-0001-4000-8000-000000000012', '{"no_onions": true}'::JSONB, 'Veggie Burger', 1, 11.99),
  ('f0a80121-0001-4000-8000-000000000021', 'e0a80121-0001-4000-8000-000000000012', 'c0a80121-0001-4000-8000-000000000013', '{}'::JSONB, 'Greek Salad', 1, 10.99);

-- Add order status history for new orders
INSERT INTO order_status_history (order_status_history_id, order_id, notes, order_status)
VALUES
  ('g0a80121-0001-4000-8000-000000000017', 'e0a80121-0001-4000-8000-000000000011', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000018', 'e0a80121-0001-4000-8000-000000000012', 'Order received', 'PENDING'),
  ('g0a80121-0001-4000-8000-000000000019', 'e0a80121-0001-4000-8000-000000000012', 'Order completed', 'COMPLETE');

-- Add price history for new menu items
INSERT INTO price_history (price_history_id, menu_item_id, price)
VALUES
  ('h0a80121-0001-4000-8000-000000000012', 'c0a80121-0001-4000-8000-000000000011', 15.99)
  ('h0a80121-0001-4000-8000-000000000013', 'c0a80121-0001-4000-8000-000000000011', 16.99)
  ('h0a80121-0001-4000-8000-000000000014', 'c0a80121-0001-4000-8000-000000000012', 11.99)
  ('h0a80121-0001-4000-8000-000000000015', 'c0a80121-0001-4000-8000-000000000013', 10.99)
  ('h0a80121-0001-4000-8000-000000000016', 'c0a80121-0001-4000-8000-000000000014', 4.99)

-- Add menu item ingredients for new dishes
INSERT INTO menu_item_ingredients (menu_item_ingredients_id, menu_item_id, ingredient_id, ingredient_name, quantity)
VALUES
  ('i0a80121-0001-4000-8000-000000000021', 'c0a80121-0001-4000-8000-000000000011', 'd0a80121-0001-4000-8000-000000000016', 'Chicken Breast', 0.2),
  ('i0a80121-0001-4000-8000-000000000022', 'c0a80121-0001-4000-8000-000000000011', 'd0a80121-0001-4000-8000-000000000017', 'Heavy Cream', 0.1),
  ('i0a80121-0001-4000-8000-000000000023', 'c0a80121-0001-4000-8000-000000000011', 'd0a80121-0001-4000-8000-000000000007', 'Pasta', 0.15),
  ('i0a80121-0001-4000-8000-000000000024', 'c0a80121-0001-4000-8000-000000000011', 'd0a80121-0001-4000-8000-000000000015', 'Parmesan Cheese', 0.05),
  ('i0a80121-0001-4000-8000-000000000025', 'c0a80121-0001-4000-8000-000000000012', 'd0a80121-0001-4000-8000-000000000018', 'Chickpeas', 0.15),
  ('i0a80121-0001-4000-8000-000000000026', 'c0a80121-0001-4000-8000-000000000012', 'd0a80121-0001-4000-8000-000000000019', 'Avocado', 0.1),
  ('i0a80121-0001-4000-8000-000000000027', 'c0a80121-0001-4000-8000-000000000012', 'd0a80121-0001-4000-8000-000000000005', 'Lettuce', 0.05),
  ('i0a80121-0001-4000-8000-000000000028', 'c0a80121-0001-4000-8000-000000000012', 'd0a80121-0001-4000-8000-000000000003', 'Tomatoes', 0.05),
  ('i0a80121-0001-4000-8000-000000000029', 'c0a80121-0001-4000-8000-000000000013', 'd0a80121-0001-4000-8000-000000000005', 'Lettuce', 0.1),
  ('i0a80121-0001-4000-8000-000000000030', 'c0a80121-0001-4000-8000-000000000013', 'd0a80121-0001-4000-8000-000000000020', 'Feta Cheese', 0.08),
  ('i0a80121-0001-4000-8000-000000000031', 'c0a80121-0001-4000-8000-000000000013', 'd0a80121-0001-4000-8000-000000000021', 'Olives', 0.06),
  ('i0a80121-0001-4000-8000-000000000032', 'c0a80121-0001-4000-8000-000000000013', 'd0a80121-0001-4000-8000-000000000003', 'Tomatoes', 0.07),
  ('i0a80121-0001-4000-8000-000000000033', 'c0a80121-0001-4000-8000-000000000014', 'd0a80121-0001-4000-8000-000000000022', 'Coffee Beans', 0.02),
  ('i0a80121-0001-4000-8000-000000000034', 'c0a80121-0001-4000-8000-000000000014', 'd0a80121-0001-4000-8000-000000000023', 'Milk', 0.2);

-- Add inventory transactions
INSERT INTO inventory_transactions (inventory_transactions_id, ingredient_id, quantity, inventory_transaction_action)
VALUES
  ('j0a80121-0001-4000-8000-000000000001', 'd0a80121-0001-4000-8000-000000000022', 5.00, 'ADD'),
  ('j0a80121-0001-4000-8000-000000000002', 'd0a80121-0001-4000-8000-000000000016', 10.00, 'ADD'),
  ('j0a80121-0001-4000-8000-000000000003', 'd0a80121-0001-4000-8000-000000000018', -0.15, 'REMOVE')
  ('j0a80121-0001-4000-8000-000000000004', 'd0a80121-0001-4000-8000-000000000019', -0.10, 'REMOVE')
  ('j0a80121-0001-4000-8000-000000000005', 'd0a80121-0001-4000-8000-000000000005', -0.15, 'REMOVE')
  ('j0a80121-0001-4000-8000-000000000006', 'd0a80121-0001-4000-8000-000000000003', -0.12, 'REMOVE')
  ('j0a80121-0001-4000-8000-000000000007', 'd0a80121-0001-4000-8000-000000000020', -0.08, 'REMOVE')
  ('j0a80121-0001-4000-8000-000000000008', 'd0a80121-0001-4000-8000-000000000021', -0.06, 'REMOVE')