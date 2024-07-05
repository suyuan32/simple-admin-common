package entenum

const (
	// TenantDefaultId is the default id of tenant
	TenantDefaultId uint64 = 1
)

const (
	// DataPermAll is the data permission of all data
	DataPermAll = 1

	// DataPermCustomDept is the data permission of custom department data
	DataPermCustomDept = 2

	// DataPermOwnDeptAndSub is the data permission of users's own department and sub departments data
	DataPermOwnDeptAndSub = 3

	// DataPermOwnDept is the data permission of users's own department data
	DataPermOwnDept = 4

	// DataPermSelf is the data permission of your own data
	DataPermSelf = 5
)
