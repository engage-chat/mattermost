// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

// Custom roles
const (
	SystemTunagAdmin = "system_tunag_admin"
	TeamTunagAdmin   = "team_tunag_admin"
)

// GetAllCustomRoleTemplates returns a map of all defined custom role templates.
func GetAllCustomRoleTemplates() map[string]Role {
	return map[string]Role{
		SystemTunagAdmin: {
			Name:        SystemTunagAdmin,
			DisplayName: SystemTunagAdmin,
			Description: "",
			Permissions: []string{
				PermissionCreatePrivateChannel.Id,
			},
		},
		TeamTunagAdmin: {
			Name:        TeamTunagAdmin,
			DisplayName: TeamTunagAdmin,
			Description: "",
			Permissions: []string{
				PermissionCreateDirectChannel.Id,
				PermissionCreateGroupChannel.Id,
			},
		},
	}
}

// AllCustomRoleNames returns a slice of all defined custom role names.
func AllCustomRoleNames() []string {
	templates := GetAllCustomRoleTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}
