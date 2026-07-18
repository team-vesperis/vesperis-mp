package key

type MaintenanceKey string

func (mk MaintenanceKey) String() string {
	return string(mk)
}

const (
	MaintenanceKey_InMaintenance MaintenanceKey = "maintenance.inMaintenance"
	MaintenanceKey_Reason        MaintenanceKey = "maintenance.reason"
	MaintenanceKey_Expire        MaintenanceKey = "maintenance.expire"
	MaintenanceKey_Expiration    MaintenanceKey = "maintenance.expiration"
)
