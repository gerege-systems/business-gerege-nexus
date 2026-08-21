-- Харилцагчийн хүснэгт эцэст нь commerce-тэй хамт ирлээ.
--
-- 00001-ийн тайлбар "contacts is not here. Commerce reads no contact rows:
-- billing keeps a contact_name string rather than a foreign key" гэж бичсэн.
-- Тэр нь billing-ийн тухайд үнэн бөгөөд бүхэл commerce-ийн тухайд биш:
-- modules/contacts/contacts.go нь энэ хүснэгтээс уншиж, бичиж, шинэчилдэг.
-- Хүснэгтийг платформ үүсгэсээр байсан тул зөрүү мэдэгдээгүй.
--
-- Одоо мэдэгдэнэ. open-gerege-nexus-ийн 00075 нь явсан аппуудын 28
-- хүснэгтийг унагаасан, тэдгээрийн дунд contacts бий: цөмд түүнийг уншдаг
-- production код байгаагүй, дэлгэцүүд нь байхгүй endpoint рүү ханддаг байсан,
-- өөрөөр хэлбэл тэр нь энэ модулийн хүснэгт байсан бөгөөд зөвхөн платформын
-- түүхэнд амьдарч байв.
--
-- Тодорхойлолт нь платформын 00003_business_apps.sql-аас үг үсгээрээ, 00001-ийн
-- бусад хүснэгттэй ижил дүрмээр: CREATE ба индекс хоёулаа IF NOT EXISTS, учир
-- нь энэ түүх нь платформын хуулбарыг аль хэдийн үүрсэн өгөгдлийн сан дээр ч
-- ажиллана.

-- +goose Up
CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL DEFAULT '',
    phone VARCHAR(64) NOT NULL DEFAULT '',
    company VARCHAR(255) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contacts_tenant ON contacts(tenant_id);

-- Тенант тусгаарлалт. Платформын 00037-ийн хэлбэр: уншихдаа session-ий
-- зөвшөөрөгдсөн байгууллагууд, бичихдээ зөвхөн идэвхтэй нэг нь.
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON contacts;
CREATE POLICY tenant_isolation ON contacts TO gerege_nexus_app
    USING (tenant_id IS NULL OR tenant_id = ANY (COALESCE(
        NULLIF(current_setting('app.allowed_tenants', true), '')::uuid[],
        ARRAY[NULLIF(current_setting('app.current_tenant', true), '')::uuid])))
    WITH CHECK (tenant_id = NULLIF(current_setting('app.current_tenant', true), '')::uuid);

-- +goose Down
DROP TABLE IF EXISTS contacts CASCADE;
