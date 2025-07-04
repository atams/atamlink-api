-- Drop Indexes
DROP INDEX IF EXISTS atamlink.idx_businesses_is_active;
DROP INDEX IF EXISTS atamlink.idx_businesses_created_by;
DROP INDEX IF EXISTS atamlink.idx_businesses_type;
DROP INDEX IF EXISTS atamlink.idx_business_users_profile;
DROP INDEX IF EXISTS atamlink.idx_business_users_role;
DROP INDEX IF EXISTS atamlink.idx_business_invites_business;
DROP INDEX IF EXISTS atamlink.idx_business_invites_expires;
DROP INDEX IF EXISTS atamlink.idx_business_subs_business;
DROP INDEX IF EXISTS atamlink.idx_business_subs_status;
DROP INDEX IF EXISTS atamlink.idx_business_subs_expires;
DROP INDEX IF EXISTS atamlink.idx_catalogs_business;
DROP INDEX IF EXISTS atamlink.idx_catalogs_is_active;
DROP INDEX IF EXISTS atamlink.idx_catalogs_theme;
DROP INDEX IF EXISTS atamlink.idx_catalog_sections_catalog;
DROP INDEX IF EXISTS atamlink.idx_catalog_sections_type;
DROP INDEX IF EXISTS atamlink.idx_catalog_sections_visible;
DROP INDEX IF EXISTS atamlink.idx_catalog_cards_section;
DROP INDEX IF EXISTS atamlink.idx_catalog_cards_type;
DROP INDEX IF EXISTS atamlink.idx_catalog_cards_visible;
DROP INDEX IF EXISTS atamlink.idx_catalog_cards_price;
DROP INDEX IF EXISTS atamlink.idx_card_details_visible;
DROP INDEX IF EXISTS atamlink.idx_card_media_card;
DROP INDEX IF EXISTS atamlink.idx_card_media_type;
DROP INDEX IF EXISTS atamlink.idx_card_links_detail;
DROP INDEX IF EXISTS atamlink.idx_card_links_type;
DROP INDEX IF EXISTS atamlink.idx_carousels_section;
DROP INDEX IF EXISTS atamlink.idx_carousels_visible;
DROP INDEX IF EXISTS atamlink.idx_carousel_items_carousel;
DROP INDEX IF EXISTS atamlink.idx_faqs_section;
DROP INDEX IF EXISTS atamlink.idx_links_section;
DROP INDEX IF EXISTS atamlink.idx_socials_section;
DROP INDEX IF EXISTS atamlink.idx_testimonials_section;
DROP INDEX IF EXISTS atamlink.uq_users_email;
DROP INDEX IF EXISTS atamlink.uq_users_username;
DROP INDEX IF EXISTS atamlink.idx_users_is_locked;
DROP INDEX IF EXISTS atamlink.ix_users_created_at;
DROP INDEX IF EXISTS atamlink.ix_users_u_email;
DROP INDEX IF EXISTS atamlink.ix_users_u_id;
DROP INDEX IF EXISTS atamlink.ix_users_u_is_active;
DROP INDEX IF EXISTS atamlink.ix_users_u_is_verified;
DROP INDEX IF EXISTS atamlink.ix_users_u_last_login_at;
DROP INDEX IF EXISTS atamlink.ix_users_u_locked_until;
DROP INDEX IF EXISTS atamlink.ix_users_u_username;
DROP INDEX IF EXISTS atamlink.idx_users_email_active;
DROP INDEX IF EXISTS atamlink.idx_users_username_active;

-- Drop Tables
DROP TABLE IF EXISTS
    atamlink.catalog_testimonials,
    atamlink.catalog_socials,
    atamlink.catalog_links,
    atamlink.catalog_faqs,
    atamlink.catalog_carousel_items,
    atamlink.catalog_carousels,
    atamlink.catalog_card_links,
    atamlink.catalog_card_media,
    atamlink.catalog_card_details,
    atamlink.catalog_cards,
    atamlink.catalog_sections,
    atamlink.catalogs,
    atamlink.business_subscriptions,
    atamlink.business_invites,
    atamlink.business_users,
    atamlink.businesses,
    atamlink.user_profiles,
    atamlink.master_themes,
    atamlink.master_plans,
    atamlink.users
CASCADE;

-- Drop ENUMs
DROP TYPE IF EXISTS business_type;
DROP TYPE IF EXISTS business_role;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS section_type;
DROP TYPE IF EXISTS card_type;
DROP TYPE IF EXISTS link_type;
DROP TYPE IF EXISTS media_type;
DROP TYPE IF EXISTS social_platform;
DROP TYPE IF EXISTS theme_type;
DROP TYPE IF EXISTS currency_type;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
