-- Moved from the platform's history, where these tables were created by
-- 00003_business_apps.sql (products, warehouses, stock_levels, stock_movements)
-- and 00004_sessions_and_business_apps.sql (billing_invoices).
--
-- Neither file could be moved as it stood, which is the whole reason commerce
-- was the last split rather than the first. 00003 creates `contacts` in the
-- same file — a table the platform keeps — and 00004 creates `sessions` and
-- `oauth2_clients` beside the invoices. An applied migration cannot be
-- rewritten, so the tables are re-declared here instead, and both originals
-- stay where they are in the platform's history.
--
-- Three deliberate differences from the originals:
--
--   * The `INSERT INTO apps` rows are gone. They registered these apps under
--     `io.example.*` ids that nothing has used since; the row every deployment
--     actually has was written by the catalogue seed from the manifests. Copying
--     the insert would have added dead app rows to a live database.
--
--   * Every index is `IF NOT EXISTS`. The originals were not — `CREATE INDEX
--     idx_products_tenant` with no guard — which is harmless when a migration
--     runs once on an empty database and fatal here, where this history runs
--     against databases that already carry the platform's copy of these tables.
--
--   * `contacts` is not here. Commerce reads no contact rows: billing keeps a
--     `contact_name` string rather than a foreign key, which is what made this
--     split possible at all.

-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    sku VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, sku)
);
CREATE INDEX IF NOT EXISTS idx_products_tenant ON products(tenant_id);

CREATE TABLE IF NOT EXISTS warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant ON warehouses(tenant_id);

CREATE TABLE IF NOT EXISTS stock_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity NUMERIC(15, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, warehouse_id, product_id)
);
CREATE INDEX IF NOT EXISTS idx_stock_levels_tenant ON stock_levels(tenant_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity_change NUMERIC(15, 4) NOT NULL,
    reference VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stock_movements_tenant ON stock_movements(tenant_id);

CREATE TABLE IF NOT EXISTS billing_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    invoice_number VARCHAR(64) NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    vat_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    ebarimt_status VARCHAR(32) NOT NULL DEFAULT 'CREATED',
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, invoice_number)
);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_tenant ON billing_invoices(tenant_id);

-- +goose Down
DROP TABLE IF EXISTS billing_invoices;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS stock_levels;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS products;
