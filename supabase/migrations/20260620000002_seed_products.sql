-- Seed: 20 products from Roastery Bakery Kitchen
-- Run this after 20260620000001_init_schema.sql
-- Safe to re-run: uses INSERT ... ON CONFLICT DO NOTHING

insert into goods (name, category, price, volume, quantity_in_stock, image_link) values
('Эспрессо',           'Кофе',     100,  50,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/28a5684a-84ee-44fc-a07b-d226d95da72b.png'),
('Американо',          'Кофе',     100, 200,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/c0b63299-1dec-479e-b0a4-6c839b21cc34.png'),
('Капучино',           'Кофе',     120, 200,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/fcfbd9b2-0e98-4d09-9305-ed6d2ca7d42c.png'),
('Латте',              'Кофе',     140, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/2f0cb865-e0cb-44bc-892c-1ef5bd2baf82.png'),
('Флэт уайт',          'Кофе',     140, 200,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/daa7f3e4-bb08-4c50-a9bd-17d6c5c5a898.png'),
('Маккиато',           'Кофе',     160, 200,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/7f6db896-8bc2-49c2-aab2-a1283e91f081.png'),
('Раф',                'Кофе',     160, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/063954b7-710c-4eb8-bc77-f7c3a8818dc5.png'),
('Айс Латте',          'Кофе',     190, 400,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/a5ea317c-f82d-48fc-8e72-f632f23ff3aa.png'),
('Эспрессо Тоник',     'Кофе',     190, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/0b6a2417-11bf-42f3-b8ad-54726fdf473e.png'),
('Матча Латте',        'Матча',    180, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/f29d345e-9437-4649-a2b6-8d5326d2c01e.png'),
('Айс Матча Латте',    'Матча',    190, 400,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/eb0fa8ca-0d98-47bb-8a0b-67bcdd0b92c2.png'),
('Матча Тоник',        'Матча',    190, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/769f9112-2d1a-424c-b860-4a10ff6ba611.png'),
('Матча Бамбл',        'Матча',    190, 400,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/d7f473b2-90b9-4adb-a843-4cadf4179062.png'),
('Какао',              'Напитки',  120, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/de7d6868-bef8-45a5-8616-44d7130ccf70.png'),
('Ораним Бамбл',       'Напитки',  190, 400,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/8a575aa1-0eb8-40ed-b2b2-16b3d9a89372.png'),
('Лимонад',            'Напитки',  190, 400,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/12f62e01-d668-42a0-9e08-76ccec998ee7.png'),
('Чай',                'Чай',       60, 300,  100, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/b088dd7a-e2aa-4615-ac14-124e908bd0d2.png'),
('Растительное молоко','Другое',    60, 200,  200, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/e00facf8-4fa1-44fd-98bb-14f3049f3124.png'),
('Сэндвич с курицей',  'Сэндвичи', 160,   1,   50, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/af452896-8982-41c1-8dfb-9eca07e19ebe.png'),
('Сэндвич с лососем',  'Сэндвичи', 160,   1,   50, 'https://bmsjqsbildaenhkhtphi.storage.supabase.co/storage/v1/object/public/order_microservices/product-images/3e1ecaab-f56b-4143-8d3d-d8e932da8a14.png')
on conflict do nothing;
