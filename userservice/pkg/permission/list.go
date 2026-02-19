package permission

import (
	"fmt"
	"strings"
)

func init() {
	if permEOF >= 64 {
		panic("too many permissions (max count is 64)")
	}
}

const (
	// count of internal permissions (they start from 0)
	InternalPermsCnt = 1
)

const (
	PerformHardResetWrite = iota

	SystemSettingsRead
	SystemSettingsWrite

	UserRolesManagementRead
	UserRolesManagementWrite

	UsersManagementRead
	UsersManagementWrite

	AnimalsManagementRead
	AnimalsManagementWrite

	AnimalsMilkingBlockRead
	AnimalsMilkingBlockWrite

	MilkingMonitorRead
	MilkingMonitorWrite

	// Animal Attentions
	AnimalAttentionsRead

	// Widgets
	WidgetsRead
	WidgetsWrite

	// Information Panel
	InformationPanelRead
	InformationPanelWrite

	// Error messages
	ErrorMessagesRead

	// Animal Events (excluding Milking blocks and Animal separations)
	AnimalEventsRead
	AnimalEventsWrite

	// Animal Separation
	AnimalSeparationRead
	AnimalSeparationWrite

	// Diagnostic Analysis Section
	DiagnosticAnalysisRead

	// Herds Analysis Section
	HerdsAnalysisRead

	// Animal Reports
	AnimalReportsRead
	AnimalReportsInteractionRead
	AnimalReportsInteractionWrite

	// Milking Reports
	MilkingReportsRead

	// Heat Reports
	HeatReportsRead

	// Diagnostic Reports
	DiagnosticReportsRead
	DiagnosticReportsWrite

	// Farm Settings
	FarmSettingsRead
	FarmSettingsWrite

	// Milking Settings
	MilkingSettingsRead
	MilkingSettingsWrite

	// Farm Settings - Locations
	FarmSettingsLocationsRead
	FarmSettingsLocationsWrite

	// Farm Settings - Milking Blocks
	FarmSettingsMilkingBlocksRead
	FarmSettingsMilkingBlocksWrite

	// Farm Settings - Milking Attentions
	FarmSettingsMilkingAttentionsRead
	FarmSettingsMilkingAttentionsWrite

	// KPI Settings
	KPISettingsRead
	KPISettingsWrite

	// Diagnostic Settings
	DiagnosticSettingsRead
	DiagnosticSettingsWrite

	// Data Import/Export Settings
	DataImportExportSettingsRead
	DataImportExportSettingsWrite

	// Parameters Settings
	ParametersSettingsRead
	ParametersSettingsWrite

	// Maintenance Section (excluding Hard Reset)
	MaintenanceRead
	MaintenanceWrite

	// Maintenance Errors
	MaintenanceErrorsRead
	MaintenanceErrorsWrite

	permEOF
)

var permsNames = map[Permission]string{
	SystemSettingsRead:  "system_settings_read",
	SystemSettingsWrite: "system_settings_write",

	UserRolesManagementRead:  "user_roles_management_read",
	UserRolesManagementWrite: "user_roles_management_write",

	UsersManagementRead:  "users_management_read",
	UsersManagementWrite: "users_management_write",

	AnimalsManagementRead:  "animals_management_read",
	AnimalsManagementWrite: "animals_management_write",

	AnimalsMilkingBlockRead:  "animals_milking_block_read",
	AnimalsMilkingBlockWrite: "animals_milking_block_write",

	MilkingMonitorRead:  "milking_monitor_read",
	MilkingMonitorWrite: "milking_monitor_write",

	PerformHardResetWrite: "perform_hard_reset_write",

	// Animal Attentions
	AnimalAttentionsRead: "animal_attentions_read",

	// Widgets
	WidgetsRead:  "widgets_read",
	WidgetsWrite: "widgets_write",

	// Information Panel
	InformationPanelRead:  "information_panel_read",
	InformationPanelWrite: "information_panel_write",

	// Error messages
	ErrorMessagesRead: "error_messages_read",

	// Animal Events (excluding Milking blocks and Animal separations)
	AnimalEventsRead:  "animal_events_read",
	AnimalEventsWrite: "animal_events_write",

	// Animal Separation
	AnimalSeparationRead:  "animal_separation_read",
	AnimalSeparationWrite: "animal_separation_write",

	// Diagnostic Analysis Section
	DiagnosticAnalysisRead: "diagnostic_analysis_read",

	// Herds Analysis Section
	HerdsAnalysisRead: "herds_analysis_read",

	// Animal Reports
	AnimalReportsRead:             "animal_reports_read",
	AnimalReportsInteractionRead:  "animal_reports_interaction_read",
	AnimalReportsInteractionWrite: "animal_reports_interaction_write",

	// Milking Reports
	MilkingReportsRead: "milking_reports_read",

	// Heat Reports
	HeatReportsRead: "heat_reports_read",

	// Diagnostic Reports
	DiagnosticReportsRead:  "diagnostic_reports_read",
	DiagnosticReportsWrite: "diagnostic_reports_write",

	// Farm Settings
	FarmSettingsRead:  "farm_settings_read",
	FarmSettingsWrite: "farm_settings_write",

	// Milking Settings
	MilkingSettingsRead:  "milking_settings_read",
	MilkingSettingsWrite: "milking_settings_write",

	// Farm Settings - Locations
	FarmSettingsLocationsRead:  "farm_settings_locations_read",
	FarmSettingsLocationsWrite: "farm_settings_locations_write",

	// Farm Settings - Milking Blocks
	FarmSettingsMilkingBlocksRead:  "farm_settings_milking_blocks_read",
	FarmSettingsMilkingBlocksWrite: "farm_settings_milking_blocks_write",

	// Farm Settings - Milking Attentions
	FarmSettingsMilkingAttentionsRead:  "farm_settings_milking_attentions_read",
	FarmSettingsMilkingAttentionsWrite: "farm_settings_milking_attentions_write",

	// KPI Settings
	KPISettingsRead:  "kpi_settings_read",
	KPISettingsWrite: "kpi_settings_write",

	// Diagnostic Settings
	DiagnosticSettingsRead:  "diagnostic_settings_read",
	DiagnosticSettingsWrite: "diagnostic_settings_write",

	// Data Import/Export Settings
	DataImportExportSettingsRead:  "data_import_export_settings_read",
	DataImportExportSettingsWrite: "data_import_export_settings_write",

	// Parameters Settings
	ParametersSettingsRead:  "parameters_settings_read",
	ParametersSettingsWrite: "parameters_settings_write",

	// Maintenance Section (excluding Hard Reset)
	MaintenanceRead:  "maintenance_read",
	MaintenanceWrite: "maintenance_write",

	// Maintenance Errors
	MaintenanceErrorsRead:  "maintenance_errors_read",
	MaintenanceErrorsWrite: "maintenance_errors_write",
}

func (p Permission) Value() int {
	return int(p)
}

func (p Permission) String() string {
	name, ok := permsNames[p]
	if !ok {
		return "unknown permission"
	}

	return name
}

func (p Permission) IsValid() bool {
	return p < permEOF
}

func (p Permission) IsExternal() bool {
	return p >= InternalPermsCnt
}

func GetPermissionByName(name string) (Permission, error) {
	for perm, strPerm := range permsNames {
		if strings.EqualFold(strPerm, name) {
			return perm, nil
		}
	}

	return 0, fmt.Errorf("%w: %s", ErrInvalidPermission, name)
}

func GetPermissionsByNames(names ...string) ([]Permission, error) {
	var permissions []Permission

	for _, perm := range names {
		perm, err := GetPermissionByName(perm)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func GetPermissionById(id uint8) (Permission, error) {
	if id >= uint8(permEOF) {
		return 0, fmt.Errorf("%w: %d", ErrInvalidPermission, id)
	}

	return Permission(id), nil
}

func GetPermissionsByIds(ids ...uint8) ([]Permission, error) {
	var permissions []Permission

	for _, id := range ids {
		perm, err := GetPermissionById(id)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func GetList() []Permission {
	var perms []Permission

	for perm := Permission(0); perm < permEOF; perm++ {
		perms = append(perms, perm)
	}

	return perms
}

func GetListOfExternal() []Permission {
	var perms []Permission

	for perm := Permission(InternalPermsCnt); perm < permEOF; perm++ {
		perms = append(perms, perm)
	}

	return perms
}
