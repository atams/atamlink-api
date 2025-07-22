// internal/constant/audit.go
package constant

// Audit Actions
const (
	// Common Actions
	AuditActionCreate = "CREATE"
	AuditActionRead   = "READ"
	AuditActionUpdate = "UPDATE"
	AuditActionDelete = "DELETE"
	
	// Business Specific Actions
	AuditActionBusinessSuspend   = "BUSINESS_SUSPEND"
	AuditActionBusinessUnsuspend = "BUSINESS_UNSUSPEND"
	
	// Business User Actions
	AuditActionUserAdd         = "USER_ADD"
	AuditActionUserRemove      = "USER_REMOVE"
	AuditActionUserRoleUpdate  = "USER_ROLE_UPDATE"
	AuditActionUserActivate    = "USER_ACTIVATE"
	AuditActionUserDeactivate  = "USER_DEACTIVATE"
	
	// Invite Actions
	AuditActionInviteCreate = "INVITE_CREATE"
	AuditActionInviteAccept = "INVITE_ACCEPT"
	AuditActionInviteExpire = "INVITE_EXPIRE"
	AuditActionInviteCancel = "INVITE_CANCEL"
	
	// Subscription Actions
	AuditActionSubscriptionCreate    = "SUBSCRIPTION_CREATE"
	AuditActionSubscriptionActivate  = "SUBSCRIPTION_ACTIVATE"
	AuditActionSubscriptionSuspend   = "SUBSCRIPTION_SUSPEND"
	AuditActionSubscriptionCancel    = "SUBSCRIPTION_CANCEL"
	AuditActionSubscriptionExpire    = "SUBSCRIPTION_EXPIRE"
	AuditActionSubscriptionRenew     = "SUBSCRIPTION_RENEW"
	AuditActionSubscriptionUpgrade   = "SUBSCRIPTION_UPGRADE"
	AuditActionSubscriptionDowngrade = "SUBSCRIPTION_DOWNGRADE"
	
	// Catalog Actions
	AuditActionCatalogPublish   = "CATALOG_PUBLISH"
	AuditActionCatalogUnpublish = "CATALOG_UNPUBLISH"
	AuditActionCatalogClone     = "CATALOG_CLONE"
	
	// Section Actions
	AuditActionSectionAdd      = "SECTION_ADD"
	AuditActionSectionRemove   = "SECTION_REMOVE"
	AuditActionSectionReorder  = "SECTION_REORDER"
	AuditActionSectionShow     = "SECTION_SHOW"
	AuditActionSectionHide     = "SECTION_HIDE"
	
	// Card Actions
	AuditActionCardAdd     = "CARD_ADD"
	AuditActionCardRemove  = "CARD_REMOVE"
	AuditActionCardReorder = "CARD_REORDER"
	AuditActionCardShow    = "CARD_SHOW"
	AuditActionCardHide    = "CARD_HIDE"
	
	// Media Actions
	AuditActionMediaAdd    = "MEDIA_ADD"
	AuditActionMediaRemove = "MEDIA_REMOVE"
	AuditActionMediaUpdate = "MEDIA_UPDATE"
)

// Audit Table Names
const (
	// Business Related Tables
	AuditTableBusinesses            = "businesses"
	AuditTableBusinessUsers         = "business_users"
	AuditTableBusinessInvites       = "business_invites"
	AuditTableBusinessSubscriptions = "business_subscriptions"
	
	// User Related Tables
	AuditTableUsers        = "users"
	AuditTableUserProfiles = "user_profiles"
	
	// Master Tables
	AuditTableMasterPlans  = "master_plans"
	AuditTableMasterThemes = "master_themes"
	
	// Catalog Related Tables
	AuditTableCatalogs               = "catalogs"
	AuditTableCatalogSections        = "catalog_sections"
	AuditTableCatalogCards           = "catalog_cards"
	AuditTableCatalogCardDetails     = "catalog_card_details"
	AuditTableCatalogCardLinks       = "catalog_card_links"
	AuditTableCatalogCardMedia       = "catalog_card_media"
	AuditTableCatalogCarousels       = "catalog_carousels"
	AuditTableCatalogCarouselItems   = "catalog_carousel_items"
	AuditTableCatalogFAQs            = "catalog_faqs"
	AuditTableCatalogLinks           = "catalog_links"
	AuditTableCatalogSocials         = "catalog_socials"
	AuditTableCatalogTestimonials    = "catalog_testimonials"
)

// Audit Types untuk menentukan tabel audit mana yang digunakan
const (
	AuditTypeBusiness = "business"
	AuditTypeCatalog  = "catalog"
)

// Audit Messages
const (
	// Business Messages
	AuditMsgBusinessCreated   = "Bisnis baru telah dibuat"
	AuditMsgBusinessUpdated   = "Data bisnis telah diperbarui"
	AuditMsgBusinessDeleted   = "Bisnis telah dihapus"
	AuditMsgBusinessSuspended = "Bisnis telah disuspend"
	AuditMsgBusinessResumed   = "Bisnis telah direaktivasi"
	
	// User Messages
	AuditMsgUserAdded        = "User baru telah ditambahkan ke bisnis"
	AuditMsgUserRemoved      = "User telah dihapus dari bisnis"
	AuditMsgUserRoleUpdated  = "Role user telah diperbarui"
	AuditMsgUserActivated    = "User telah diaktifkan"
	AuditMsgUserDeactivated  = "User telah dinonaktifkan"
	
	// Invite Messages
	AuditMsgInviteCreated  = "Undangan baru telah dibuat"
	AuditMsgInviteAccepted = "Undangan telah diterima"
	AuditMsgInviteExpired  = "Undangan telah kadaluarsa"
	AuditMsgInviteCanceled = "Undangan telah dibatalkan"
	
	// Subscription Messages
	AuditMsgSubscriptionCreated    = "Subscription baru telah dibuat"
	AuditMsgSubscriptionActivated  = "Subscription telah diaktifkan"
	AuditMsgSubscriptionSuspended  = "Subscription telah disuspend"
	AuditMsgSubscriptionCanceled   = "Subscription telah dibatalkan"
	AuditMsgSubscriptionExpired    = "Subscription telah kadaluarsa"
	AuditMsgSubscriptionRenewed    = "Subscription telah diperpanjang"
	AuditMsgSubscriptionUpgraded   = "Subscription telah diupgrade"
	AuditMsgSubscriptionDowngraded = "Subscription telah didowngrade"
	
	// Catalog Messages
	AuditMsgCatalogCreated    = "Katalog baru telah dibuat"
	AuditMsgCatalogUpdated    = "Data katalog telah diperbarui"
	AuditMsgCatalogDeleted    = "Katalog telah dihapus"
	AuditMsgCatalogPublished  = "Katalog telah dipublikasikan"
	AuditMsgCatalogUnpublished = "Katalog telah di-unpublish"
	AuditMsgCatalogCloned     = "Katalog telah diduplikasi"
	
	// Section Messages
	AuditMsgSectionAdded    = "Section baru telah ditambahkan"
	AuditMsgSectionUpdated  = "Data section telah diperbarui"
	AuditMsgSectionRemoved  = "Section telah dihapus"
	AuditMsgSectionReordered = "Urutan section telah diubah"
	AuditMsgSectionShown    = "Section telah ditampilkan"
	AuditMsgSectionHidden   = "Section telah disembunyikan"
	
	// Card Messages
	AuditMsgCardAdded     = "Card baru telah ditambahkan"
	AuditMsgCardUpdated   = "Data card telah diperbarui"
	AuditMsgCardRemoved   = "Card telah dihapus"
	AuditMsgCardReordered = "Urutan card telah diubah"
	AuditMsgCardShown     = "Card telah ditampilkan"
	AuditMsgCardHidden    = "Card telah disembunyikan"
	
	// Media Messages
	AuditMsgMediaAdded   = "Media baru telah ditambahkan"
	AuditMsgMediaUpdated = "Media telah diperbarui"
	AuditMsgMediaRemoved = "Media telah dihapus"
)

// GetAuditMessage returns appropriate message for action
func GetAuditMessage(action string) string {
	messages := map[string]string{
		AuditActionCreate: "Data telah dibuat",
		AuditActionUpdate: "Data telah diperbarui",
		AuditActionDelete: "Data telah dihapus",
		
		// Business
		AuditActionBusinessSuspend:   AuditMsgBusinessSuspended,
		AuditActionBusinessUnsuspend: AuditMsgBusinessResumed,
		
		// User
		AuditActionUserAdd:        AuditMsgUserAdded,
		AuditActionUserRemove:     AuditMsgUserRemoved,
		AuditActionUserRoleUpdate: AuditMsgUserRoleUpdated,
		AuditActionUserActivate:   AuditMsgUserActivated,
		AuditActionUserDeactivate: AuditMsgUserDeactivated,
		
		// Invite
		AuditActionInviteCreate: AuditMsgInviteCreated,
		AuditActionInviteAccept: AuditMsgInviteAccepted,
		AuditActionInviteExpire: AuditMsgInviteExpired,
		AuditActionInviteCancel: AuditMsgInviteCanceled,
		
		// Subscription
		AuditActionSubscriptionCreate:    AuditMsgSubscriptionCreated,
		AuditActionSubscriptionActivate:  AuditMsgSubscriptionActivated,
		AuditActionSubscriptionSuspend:   AuditMsgSubscriptionSuspended,
		AuditActionSubscriptionCancel:    AuditMsgSubscriptionCanceled,
		AuditActionSubscriptionExpire:    AuditMsgSubscriptionExpired,
		AuditActionSubscriptionRenew:     AuditMsgSubscriptionRenewed,
		AuditActionSubscriptionUpgrade:   AuditMsgSubscriptionUpgraded,
		AuditActionSubscriptionDowngrade: AuditMsgSubscriptionDowngraded,
		
		// Catalog
		AuditActionCatalogPublish:   AuditMsgCatalogPublished,
		AuditActionCatalogUnpublish: AuditMsgCatalogUnpublished,
		AuditActionCatalogClone:     AuditMsgCatalogCloned,
		
		// Section
		AuditActionSectionAdd:     AuditMsgSectionAdded,
		AuditActionSectionRemove:  AuditMsgSectionRemoved,
		AuditActionSectionReorder: AuditMsgSectionReordered,
		AuditActionSectionShow:    AuditMsgSectionShown,
		AuditActionSectionHide:    AuditMsgSectionHidden,
		
		// Card
		AuditActionCardAdd:     AuditMsgCardAdded,
		AuditActionCardRemove:  AuditMsgCardRemoved,
		AuditActionCardReorder: AuditMsgCardReordered,
		AuditActionCardShow:    AuditMsgCardShown,
		AuditActionCardHide:    AuditMsgCardHidden,
		
		// Media
		AuditActionMediaAdd:    AuditMsgMediaAdded,
		AuditActionMediaUpdate: AuditMsgMediaUpdated,
		AuditActionMediaRemove: AuditMsgMediaRemoved,
	}
	
	if msg, exists := messages[action]; exists {
		return msg
	}
	return "Aktivitas telah dilakukan"
}

// GetAuditType determines which audit table to use based on table name
func GetAuditType(tableName string) string {
	businessTables := map[string]bool{
		AuditTableBusinesses:            true,
		AuditTableBusinessUsers:         true,
		AuditTableBusinessInvites:       true,
		AuditTableBusinessSubscriptions: true,
		AuditTableUsers:                 true,
		AuditTableUserProfiles:          true,
		AuditTableMasterPlans:           true,
		AuditTableMasterThemes:          true,
	}
	
	if businessTables[tableName] {
		return AuditTypeBusiness
	}
	return AuditTypeCatalog
}