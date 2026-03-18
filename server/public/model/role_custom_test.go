// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllCustomRoleGroups(t *testing.T) {
	definedGroups := []string{CustomRolesUnofficial}
	groups := AllCustomRoleGroups()
	require.NotNil(t, groups)
	require.Len(t, groups, len(customRoleGroupFactories))
	require.ElementsMatch(t, definedGroups, groups)
}

func TestCustomRoleNamesForGroup(t *testing.T) {
	testCase := []struct {
		name          string
		groupName     string
		requiredNames []string
		isNil         bool
	}{
		{"should return role names for unofficial group", CustomRolesUnofficial, []string{SystemTunagAdmin, TeamTunagAdmin}, false},
		{"should return nil for invalid group", "invalid_group", nil, true},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			names := CustomRoleNamesForGroup(tc.groupName)

			if tc.isNil {
				require.Nil(t, names)
			} else {
				requiredRoles := MakeTunagCustomRoles(tc.groupName)
				require.NotNil(t, names)
				require.Len(t, names, len(requiredRoles))
				require.ElementsMatch(t, tc.requiredNames, names)
			}
		})
	}
}

func TestMakeTunagCustomRoles(t *testing.T) {
	t.Run("should return roles for unofficial group", func(t *testing.T) {
		roles := MakeTunagCustomRoles(CustomRolesUnofficial)
		require.NotNil(t, roles)
		require.Len(t, roles, len(makeTunagCustomRolesUnofficial()))

		role, ok := roles[SystemTunagAdmin]
		require.True(t, ok)
		require.Equal(t, SystemTunagAdmin, role.Name)
		require.Equal(t, SystemTunagAdmin, role.DisplayName)
		require.Equal(t, []string{PermissionCreatePrivateChannel.Id}, role.Permissions)

		role, ok = roles[TeamTunagAdmin]
		require.True(t, ok)
		require.Equal(t, TeamTunagAdmin, role.Name)
		require.Equal(t, TeamTunagAdmin, role.DisplayName)
		require.Equal(t, []string{PermissionCreateDirectChannel.Id, PermissionCreateGroupChannel.Id}, role.Permissions)
	})

	t.Run("should return nil for invalid group", func(t *testing.T) {
		roles := MakeTunagCustomRoles("invalid_group")
		require.Nil(t, roles)
	})
}
