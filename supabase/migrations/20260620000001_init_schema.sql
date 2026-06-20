-- Enable UUID extension
create extension if not exists "uuid-ossp";

-- Users
create table if not exists users (
    id               uuid primary key default uuid_generate_v4(),
    email            text unique not null,
    pass_hash        bytea not null,
    is_admin         boolean not null default false
);

-- Goods (inventory)
create table if not exists goods (
    id                 uuid primary key default uuid_generate_v4(),
    name               text not null,
    category           text not null,
    image_link         text,
    description        text,
    price              integer not null,
    volume             integer,
    quantity_in_stock  integer
);

-- Orders
create table if not exists orders (
    id          uuid primary key default uuid_generate_v4(),
    user_id     uuid not null,
    total       decimal(10,2) not null,
    status      varchar(20) not null default 'CREATED',
    created_at  timestamptz default now(),
    updated_at  timestamptz default now()
);
create index if not exists idx_orders_user_id on orders(user_id);

-- Order items
create table if not exists order_items (
    id          uuid primary key default uuid_generate_v4(),
    order_id    uuid not null references orders(id) on delete cascade,
    product_id  uuid not null,
    quantity    integer not null
);
create index if not exists idx_order_items_order_id on order_items(order_id);

-- Sagas
create table if not exists sagas (
    id            uuid primary key default uuid_generate_v4(),
    current_step  text not null,
    user_id       uuid not null,
    created_at    timestamptz default now(),
    updated_at    timestamptz default now(),
    error_reason  text
);
