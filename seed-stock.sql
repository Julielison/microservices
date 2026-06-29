-- Execute após o docker-compose subir:
-- docker exec -i <mysql-container> mysql -u root -ps3cr3t order < seed-stock.sql

USE `order`;

INSERT IGNORE INTO stock_items (created_at, updated_at, deleted_at, product_code, description, unit_price)
VALUES
  (NOW(), NOW(), NULL, 'ABC123',  'Produto A',  10.50),
  (NOW(), NOW(), NULL, 'XYZ789',  'Produto B',  20.00),
  (NOW(), NOW(), NULL, 'PROD001', 'Produto C',   5.00),
  (NOW(), NOW(), NULL, 'PROD002', 'Produto D',  15.75),
  (NOW(), NOW(), NULL, 'PROD003', 'Produto E',  99.99);
