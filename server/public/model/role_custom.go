// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

// Custom role groups
const (
	CustomRolesUnofficial = "unofficial_channel"
	CustomRolesTest       = "test_group"
)

// Custom roles
const (
	SystemTunagAdmin = "system_tunag_admin"
	TeamTunagAdmin   = "team_tunag_admin"

	TestTunagAdmin = "test_tunag_admin"
)

// customRoleGroupFactories is the master list of all custom role group definitions.
// It maps a group name to a factory function that returns the roles for that group.
// This factory pattern is used to avoid initialization-order errors with permissions.
var customRoleGroupFactories = map[string]func() map[string]Role{
	CustomRolesUnofficial: makeTunagCustomRolesUnofficial,
	CustomRolesTest:       makeTunagCustomRolesTest,
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

func makeTunagCustomRolesTest() map[string]Role {
	return map[string]Role{
		TestTunagAdmin: {
			Name:        TestTunagAdmin,
			DisplayName: TestTunagAdmin,
			Description: "",
			Permissions: []string{
				PermissionCreateBot.Id,
			},
		},
	}
}

func makeTunagCustomRolesUnofficial() map[string]Role {
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
