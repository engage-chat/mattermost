// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

// Custom role groups
const (
	CustomRolesUnofficial = "unofficial"
)

// Custom roles
const (
	SystemTunagUnofficial = "system_tunag_unofficial"
	TeamTunagUnofficial   = "team_tunag_unofficial"
)

// customRoleGroupFactories is the master list of all custom role group definitions.
// It maps a group name to a factory function that returns the roles for that group.
// This factory pattern is used to avoid initialization-order errors with permissions.
var customRoleGroupFactories = map[string]func() map[string]Role{
	CustomRolesUnofficial: makeTunagCustomRolesUnofficial,
}

func AllCustomRoleGroups() []string {
	groups := make([]string, 0, len(customRoleGroupFactories))
	for groupName := range customRoleGroupFactories {
		groups = append(groups, groupName)
	}
	return groups
}

func CustomRoleNamesForGroup(customRoleGroup string) []string {
	roles := MakeTunagCustomRoles(customRoleGroup)
	if roles == nil {
		return nil
	}

	names := make([]string, 0, len(roles))
	for roleKey := range roles {
		names = append(names, roleKey)
	}
	return names
}

// MakeTunagCustomRoles returns a map of custom roles for a given group.
// It acts as a public accessor to the customRoleGroupFactories map.
func MakeTunagCustomRoles(customRoleGroup string) map[string]Role {
	if factory, ok := customRoleGroupFactories[customRoleGroup]; ok {
		return factory()
	}
	return nil
}

func makeTunagCustomRolesUnofficial() map[string]Role {
	return map[string]Role{
		SystemTunagUnofficial: {
			Name:        SystemTunagUnofficial,
			DisplayName: SystemTunagUnofficial,
			Description: "",
			Permissions: []string{
				PermissionCreatePrivateChannel.Id,
			},
		},
		TeamTunagUnofficial: {
			Name:        TeamTunagUnofficial,
			DisplayName: TeamTunagUnofficial,
			Description: "",
			Permissions: []string{
				PermissionCreateDirectChannel.Id,
				PermissionCreateGroupChannel.Id,
			},
		},
	}
}
