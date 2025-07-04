package constant

// Business roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Role hierarchy untuk permission checking
var RoleHierarchy = map[string]int{
	RoleOwner:  4,
	RoleAdmin:  3,
	RoleEditor: 2,
	RoleViewer: 1,
}

// HasMinimumRole check apakah user role >= minimum role yang dibutuhkan
func HasMinimumRole(userRole, minimumRole string) bool {
	userLevel, exists := RoleHierarchy[userRole]
	if !exists {
		return false
	}

	minLevel, exists := RoleHierarchy[minimumRole]
	if !exists {
		return false
	}

	return userLevel >= minLevel
}

// IsValidRole check apakah role valid
func IsValidRole(role string) bool {
	_, exists := RoleHierarchy[role]
	return exists
}

// GetAllRoles mendapatkan semua roles
func GetAllRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleEditor, RoleViewer}
}

// Permission constants
const (
	// Business permissions
	PermBusinessView   = "business:view"
	PermBusinessCreate = "business:create"
	PermBusinessUpdate = "business:update"
	PermBusinessDelete = "business:delete"

	// Catalog permissions
	PermCatalogView   = "catalog:view"
	PermCatalogCreate = "catalog:create"
	PermCatalogUpdate = "catalog:update"
	PermCatalogDelete = "catalog:delete"

	// User management permissions
	PermUserView   = "user:view"
	PermUserInvite = "user:invite"
	PermUserUpdate = "user:update"
	PermUserRemove = "user:remove"

	// Subscription permissions
	PermSubscriptionView   = "subscription:view"
	PermSubscriptionUpdate = "subscription:update"
)

// RolePermissions mapping role ke permissions
var RolePermissions = map[string][]string{
	RoleOwner: {
		PermBusinessView, PermBusinessCreate, PermBusinessUpdate, PermBusinessDelete,
		PermCatalogView, PermCatalogCreate, PermCatalogUpdate, PermCatalogDelete,
		PermUserView, PermUserInvite, PermUserUpdate, PermUserRemove,
		PermSubscriptionView, PermSubscriptionUpdate,
	},
	RoleAdmin: {
		PermBusinessView, PermBusinessUpdate,
		PermCatalogView, PermCatalogCreate, PermCatalogUpdate, PermCatalogDelete,
		PermUserView, PermUserInvite, PermUserUpdate,
		PermSubscriptionView,
	},
	RoleEditor: {
		PermBusinessView,
		PermCatalogView, PermCatalogCreate, PermCatalogUpdate,
		PermUserView,
		PermSubscriptionView,
	},
	RoleViewer: {
		PermBusinessView,
		PermCatalogView,
		PermUserView,
		PermSubscriptionView,
	},
}

// HasPermission check apakah role memiliki permission tertentu
func HasPermission(role, permission string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}