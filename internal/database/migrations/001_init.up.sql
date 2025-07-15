CREATE SCHEMA IF NOT EXISTS atamlink;

-- UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUMs
CREATE TYPE business_type AS ENUM (
    'retail',
    'service',
    'manufacturing',
    'technology',
    'hospitality',
    'healthcare',
    'education',
    'other'
);

CREATE TYPE business_role AS ENUM (
    'owner',
    'admin',
    'editor',
    'viewer'
);

CREATE TYPE subscription_status AS ENUM (
    'active',
    'expired',
    'cancelled',
    'pending',
    'suspended'
);

CREATE TYPE section_type AS ENUM (
    'hero',
    'cards',
    'carousel',
    'faqs',
    'links',
    'socials',
    'testimonials',
    'cta',
    'text',
    'video'
);

CREATE TYPE card_type AS ENUM (
    'product',
    'service',
    'portfolio',
    'article',
    'event',
    'offer'
);

CREATE TYPE link_type AS ENUM (
    'whatsapp',
    'shopee',
    'tokopedia',
    'website',
    'tiktokshop',
    'facebook',
    'instagram',
    'telegram',
    'email',
    'phone',
    'custom'
);

CREATE TYPE media_type AS ENUM (
    'thumbnail',
    'cover',
    'gallery',
    'video',
    'document'
);

CREATE TYPE social_platform AS ENUM (
    'facebook',
    'instagram',
    'twitter',
    'linkedin',
    'youtube',
    'tiktok',
    'whatsapp',
    'telegram',
    'pinterest',
    'github'
);

CREATE TYPE theme_type AS ENUM (
    'minimal',
    'modern',
    'classic',
    'bold',
    'elegant',
    'playful',
    'professional',
    'creative'
);

CREATE TYPE currency_type AS ENUM (
    'IDR'
);

-- Tables

CREATE TABLE atamlink.users (
    "u_id" uuid DEFAULT uuid_generate_v4() NOT NULL,
    "u_email" character varying(255) NOT NULL,
    "u_username" character varying(100) NOT NULL,
    "u_password_hash" character varying(255) NOT NULL,
    "u_is_active" boolean NOT NULL,
    "u_is_verified" boolean NOT NULL,
    "u_is_locked" boolean NOT NULL,
    "u_email_verified_at" timestamptz,
    "u_last_login_at" timestamptz,
    "u_failed_login_attempts" integer NOT NULL,
    "u_locked_until" timestamptz,
    "u_metadata" json,
    "u_ip_address" character varying(45),
    "u_user_agent" character varying,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz,
    CONSTRAINT "users_pkey" PRIMARY KEY ("u_id"),
    CONSTRAINT "ck_users_email_length" CHECK ((length((u_email)::text) >= 3)),
    CONSTRAINT "ck_users_username_length" CHECK ((length((u_username)::text) >= 3))
) WITH (oids = false);

CREATE TABLE atamlink.master_plans (
    mp_id BIGSERIAL PRIMARY KEY,
    mp_name VARCHAR(100) NOT NULL UNIQUE,
    mp_price INTEGER NOT NULL DEFAULT 0,
    mp_duration INTERVAL NOT NULL,
    mp_features JSONB NOT NULL DEFAULT '{}',
    mp_is_active BOOLEAN NOT NULL DEFAULT true,
    mp_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE atamlink.master_themes (
    mt_id BIGSERIAL PRIMARY KEY,
    mt_name VARCHAR(100) NOT NULL UNIQUE,
    mt_description TEXT,
    mt_type theme_type NOT NULL,
    mt_default_settings JSONB NOT NULL DEFAULT '{}',
    mt_is_premium BOOLEAN NOT NULL DEFAULT false,
    mt_is_active BOOLEAN NOT NULL DEFAULT true,
    mt_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE atamlink.user_profiles (
    up_id BIGSERIAL PRIMARY KEY,
    up_u_id UUID NOT NULL UNIQUE REFERENCES atamlink.users(u_id) ON DELETE CASCADE,
    up_phone VARCHAR(20) UNIQUE,
    up_display_name VARCHAR(200),
    up_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    up_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.businesses (
    b_id BIGSERIAL PRIMARY KEY,
    b_slug VARCHAR(100) NOT NULL UNIQUE,
    b_name VARCHAR(200) NOT NULL,
    b_logo_url VARCHAR,
    b_type business_type NOT NULL,
    b_is_active BOOLEAN NOT NULL DEFAULT true,
    b_is_suspended BOOLEAN NOT NULL DEFAULT false,
    b_suspension_reason TEXT,
    b_suspended_by BIGINT,
    b_suspended_at TIMESTAMP,
    b_created_by BIGINT NOT NULL ,
    b_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    b_updated_by BIGINT,
    b_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.business_users (
    bu_id BIGSERIAL PRIMARY KEY,
    bu_b_id BIGINT NOT NULL REFERENCES atamlink.businesses(b_id) ON DELETE CASCADE,
    bu_up_id BIGINT NOT NULL REFERENCES atamlink.user_profiles(up_id) ON DELETE CASCADE,
    bu_role business_role NOT NULL DEFAULT 'viewer',
    bu_is_owner BOOLEAN NOT NULL DEFAULT false,
    bu_is_active BOOLEAN NOT NULL DEFAULT true,
    bu_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(bu_b_id, bu_up_id)
);

CREATE TABLE atamlink.business_invites (
    bi_id BIGSERIAL PRIMARY KEY,
    bi_b_id BIGINT NOT NULL REFERENCES atamlink.businesses(b_id) ON DELETE CASCADE,
    bi_token VARCHAR(100) NOT NULL UNIQUE,
    bi_role business_role NOT NULL DEFAULT 'viewer',
    bi_invited_by BIGINT NOT NULL,
    bi_is_used BOOLEAN NOT NULL DEFAULT false,
    bi_expires_at TIMESTAMP NOT NULL,
    bi_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE atamlink.business_subscriptions (
    bs_id BIGSERIAL PRIMARY KEY,
    bs_b_id BIGINT NOT NULL REFERENCES atamlink.businesses(b_id) ON DELETE CASCADE,
    bs_mp_id BIGINT NOT NULL REFERENCES atamlink.master_plans(mp_id),
    bs_status subscription_status NOT NULL DEFAULT 'pending',
    bs_starts_at TIMESTAMP NOT NULL,
    bs_expires_at TIMESTAMP NOT NULL,
    bs_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bs_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalogs (
    c_id BIGSERIAL PRIMARY KEY,
    c_b_id BIGINT NOT NULL REFERENCES atamlink.businesses(b_id) ON DELETE CASCADE,
    c_mt_id BIGINT NOT NULL REFERENCES atamlink.master_themes(mt_id),
    c_slug VARCHAR(100) NOT NULL UNIQUE,
    c_qr_url VARCHAR(500),
    c_title VARCHAR(200) NOT NULL,
    c_subtitle VARCHAR(300),
    c_is_active BOOLEAN NOT NULL DEFAULT true,
    c_settings JSONB DEFAULT '{}',
    c_created_by BIGINT NOT NULL,
    c_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    c_updated_by BIGINT,
    c_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_sections (
    cs_id BIGSERIAL PRIMARY KEY,
    cs_c_id BIGINT NOT NULL REFERENCES atamlink.catalogs(c_id) ON DELETE CASCADE,
    cs_type section_type NOT NULL,
    cs_is_visible BOOLEAN NOT NULL DEFAULT true,
    cs_config JSONB DEFAULT '{}',
    cs_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cs_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_cards (
    cc_id BIGSERIAL PRIMARY KEY,
    cc_cs_id BIGINT NOT NULL REFERENCES atamlink.catalog_sections(cs_id) ON DELETE CASCADE,
    cc_title VARCHAR(200) NOT NULL,
    cc_subtitle VARCHAR(300),
    cc_type card_type NOT NULL,
    cc_url VARCHAR(500),
    cc_is_visible BOOLEAN NOT NULL DEFAULT true,
    cc_has_detail BOOLEAN NOT NULL DEFAULT false,
    cc_price INTEGER,
    cc_discount INTEGER DEFAULT 0,
    cc_currency currency_type DEFAULT 'IDR',
    cc_created_by BIGINT NOT NULL,
    cc_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cc_updated_by BIGINT,
    cc_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_card_details (
    ccd_id BIGSERIAL PRIMARY KEY,
    ccd_cc_id BIGINT NOT NULL UNIQUE REFERENCES atamlink.catalog_cards(cc_id) ON DELETE CASCADE,
    ccd_slug VARCHAR(100) NOT NULL UNIQUE,
    ccd_description TEXT,
    ccd_is_visible BOOLEAN NOT NULL DEFAULT true,
    ccd_created_by BIGINT NOT NULL,
    ccd_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ccd_updated_by BIGINT,
    ccd_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_card_media (
    ccm_id BIGSERIAL PRIMARY KEY,
    ccm_cc_id BIGINT NOT NULL REFERENCES atamlink.catalog_cards(cc_id) ON DELETE CASCADE,
    ccm_type media_type NOT NULL,
    ccm_url VARCHAR(500) NOT NULL,
    ccm_created_by BIGINT NOT NULL,
    ccm_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ccm_updated_by BIGINT,
    ccm_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_card_links (
    ccl_id BIGSERIAL PRIMARY KEY,
    ccl_ccd_id BIGINT NOT NULL REFERENCES atamlink.catalog_card_details(ccd_id) ON DELETE CASCADE,
    ccl_type link_type NOT NULL,
    ccl_url VARCHAR(500) NOT NULL,
    ccl_is_visible BOOLEAN NOT NULL DEFAULT true,
    ccl_created_by BIGINT NOT NULL,
    ccl_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ccl_updated_by BIGINT,
    ccl_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_carousels (
    cr_id BIGSERIAL PRIMARY KEY,
    cr_cs_id BIGINT NOT NULL REFERENCES atamlink.catalog_sections(cs_id) ON DELETE CASCADE,
    cr_title VARCHAR(200),
    cr_is_visible BOOLEAN NOT NULL DEFAULT true,
    cr_created_by BIGINT NOT NULL,
    cr_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cr_updated_by BIGINT,
    cr_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_carousel_items (
    cci_id BIGSERIAL PRIMARY KEY,
    cci_cr_id BIGINT NOT NULL REFERENCES atamlink.catalog_carousels(cr_id) ON DELETE CASCADE,
    cci_image_url VARCHAR(500) NOT NULL,
    cci_caption VARCHAR(200),
    cci_description TEXT,
    cci_link_url VARCHAR(500),
    cci_created_by BIGINT NOT NULL,
    cci_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cci_updated_by BIGINT,
    cci_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_faqs (
    cf_id BIGSERIAL PRIMARY KEY,
    cf_cs_id BIGINT NOT NULL REFERENCES atamlink.catalog_sections(cs_id) ON DELETE CASCADE,
    cf_question TEXT NOT NULL,
    cf_answer TEXT NOT NULL,
    cf_is_visible BOOLEAN NOT NULL DEFAULT true,
    cf_created_by BIGINT NOT NULL,
    cf_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cf_updated_by BIGINT,
    cf_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_links (
    cl_id BIGSERIAL PRIMARY KEY,
    cl_cs_id BIGINT NOT NULL REFERENCES atamlink.catalog_sections(cs_id) ON DELETE CASCADE,
    cl_url VARCHAR(500) NOT NULL,
    cl_display_name VARCHAR(200) NOT NULL,
    cl_is_visible BOOLEAN NOT NULL DEFAULT true,
    cl_created_by BIGINT NOT NULL,
    cl_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cl_updated_by BIGINT,
    cl_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_socials (
    csoc_id BIGSERIAL PRIMARY KEY,
    csoc_cs_id BIGINT NOT NULL REFERENCES atamlink.catalog_sections(cs_id) ON DELETE CASCADE,
    csoc_platform social_platform NOT NULL,
    csoc_url VARCHAR(500) NOT NULL,
    csoc_is_visible BOOLEAN NOT NULL DEFAULT true,
    csoc_created_by BIGINT NOT NULL,
    csoc_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    csoc_updated_by BIGINT,
    csoc_updated_at TIMESTAMP 
);

CREATE TABLE atamlink.catalog_testimonials (
    ct_id BIGSERIAL PRIMARY KEY,
    ct_cs_id BIGINT NOT NULL REFERENCES atamlink.catalog_sections(cs_id) ON DELETE CASCADE,
    ct_message TEXT NOT NULL,
    ct_author VARCHAR(200) NOT NULL,
    ct_is_visible BOOLEAN NOT NULL DEFAULT true,
    ct_created_by BIGINT NOT NULL,
    ct_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ct_updated_by BIGINT,
    ct_updated_at TIMESTAMP 
);

-- Indexes
CREATE INDEX idx_businesses_is_active ON businesses(b_is_active);
CREATE INDEX idx_businesses_created_by ON businesses(b_created_by);
CREATE INDEX idx_businesses_type ON businesses(b_type);
CREATE INDEX idx_business_users_profile ON business_users(bu_up_id);
CREATE INDEX idx_business_users_role ON business_users(bu_role);
CREATE INDEX idx_business_invites_business ON business_invites(bi_b_id);
CREATE INDEX idx_business_invites_expires ON business_invites(bi_expires_at);
CREATE INDEX idx_business_subs_business ON business_subscriptions(bs_b_id);
CREATE INDEX idx_business_subs_status ON business_subscriptions(bs_status);
CREATE INDEX idx_business_subs_expires ON business_subscriptions(bs_expires_at);
CREATE INDEX idx_catalogs_business ON catalogs(c_b_id);
CREATE INDEX idx_catalogs_is_active ON catalogs(c_is_active);
CREATE INDEX idx_catalogs_theme ON catalogs(c_mt_id);
CREATE INDEX idx_catalog_sections_catalog ON catalog_sections(cs_c_id);
CREATE INDEX idx_catalog_sections_type ON catalog_sections(cs_type);
CREATE INDEX idx_catalog_sections_visible ON catalog_sections(cs_is_visible);
CREATE INDEX idx_catalog_cards_section ON catalog_cards(cc_cs_id);
CREATE INDEX idx_catalog_cards_type ON catalog_cards(cc_type);
CREATE INDEX idx_catalog_cards_visible ON catalog_cards(cc_is_visible);
CREATE INDEX idx_catalog_cards_price ON catalog_cards(cc_price);
CREATE INDEX idx_card_details_visible ON catalog_card_details(ccd_is_visible);
CREATE INDEX idx_card_media_card ON catalog_card_media(ccm_cc_id);
CREATE INDEX idx_card_media_type ON catalog_card_media(ccm_type);
CREATE INDEX idx_card_links_detail ON catalog_card_links(ccl_ccd_id);
CREATE INDEX idx_card_links_type ON catalog_card_links(ccl_type);
CREATE INDEX idx_carousels_section ON catalog_carousels(cr_cs_id);
CREATE INDEX idx_carousels_visible ON catalog_carousels(cr_is_visible);
CREATE INDEX idx_carousel_items_carousel ON catalog_carousel_items(cci_cr_id);
CREATE INDEX idx_faqs_section ON catalog_faqs(cf_cs_id);
CREATE INDEX idx_links_section ON catalog_links(cl_cs_id);
CREATE INDEX idx_socials_section ON catalog_socials(csoc_cs_id);
CREATE INDEX idx_testimonials_section ON catalog_testimonials(ct_cs_id);
CREATE UNIQUE INDEX uq_users_email ON atamlink.users USING btree (u_email);
CREATE UNIQUE INDEX uq_users_username ON atamlink.users USING btree (u_username);
CREATE INDEX idx_users_is_locked ON atamlink.users USING btree (u_is_locked);
CREATE INDEX ix_users_created_at ON atamlink.users USING btree (created_at);
CREATE UNIQUE INDEX ix_users_u_email ON atamlink.users USING btree (u_email);
CREATE INDEX ix_users_u_id ON atamlink.users USING btree (u_id);
CREATE INDEX ix_users_u_is_active ON atamlink.users USING btree (u_is_active);
CREATE INDEX ix_users_u_is_verified ON atamlink.users USING btree (u_is_verified);
CREATE INDEX ix_users_u_last_login_at ON atamlink.users USING btree (u_last_login_at);
CREATE INDEX ix_users_u_locked_until ON atamlink.users USING btree (u_locked_until);
CREATE UNIQUE INDEX ix_users_u_username ON atamlink.users USING btree (u_username);
CREATE INDEX idx_users_email_active ON atamlink.users USING btree (u_email) WHERE (u_is_active = true);
CREATE INDEX idx_users_username_active ON atamlink.users USING btree (u_username) WHERE (u_is_active = true);

-- ENUM untuk jenis aksi audit
CREATE TYPE audit_action_type AS ENUM (
    'CREATE',
    'UPDATE',
    'DELETE',
    'INVITE_SENT',
    'INVITE_USED'
);

-- Tabel utama untuk Audit Log
CREATE TABLE atamlink.audit_logs (
    al_id BIGSERIAL PRIMARY KEY,
    al_timestamp TIMESTAMPTZ NOT NULL DEFAULT now(),
    al_user_profile_id BIGINT REFERENCES atamlink.user_profiles(up_id) ON DELETE SET NULL,
    al_business_id BIGINT REFERENCES atamlink.businesses(b_id) ON DELETE SET NULL,
    al_action audit_action_type NOT NULL,
    al_table_name TEXT,
    al_record_id TEXT,
    al_old_data JSONB,
    al_new_data JSONB,
    al_context JSONB,
    al_reason TEXT
);

-- Indexes untuk mempercepat query pada tabel audit_logs
CREATE INDEX idx_audit_logs_user_profile_id ON atamlink.audit_logs(al_user_profile_id);
CREATE INDEX idx_audit_logs_business_id ON atamlink.audit_logs(al_business_id);
CREATE INDEX idx_audit_logs_timestamp ON atamlink.audit_logs(al_timestamp);
CREATE INDEX idx_audit_logs_record ON atamlink.audit_logs(al_table_name, al_record_id);
CREATE INDEX idx_audit_logs_action ON atamlink.audit_logs(al_action);

-- Index GIN pada kolom JSONB agar bisa melakukan query ke dalam datanya secara efisien
CREATE INDEX idx_audit_logs_context_gin ON atamlink.audit_logs USING GIN(al_context);