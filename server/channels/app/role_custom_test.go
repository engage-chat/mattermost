// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"
)

func TestGetCustomRolesForGroup(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	testGroup := model.CustomRolesUnofficial

	t.Run("get a specific group", func(t *testing.T) {
		_, err := th.App.EnableCustomRoles(th.Context, testGroup)
		require.Nil(t, err)
		roles, err := th.App.GetCustomRolesForGroup(th.Context, testGroup)
		require.Nil(t, err)
		expectedRoles := model.MakeTunagCustomRoles(testGroup)
		actualRoles := make(map[string]*model.Role)
		for _, role := range roles {
			actualRoles[role.Name] = role
		}
		require.NotNil(t, roles)
		require.Len(t, roles, len(expectedRoles))
		for name, expectedRole := range expectedRoles {
			actualRole, ok := actualRoles[name]
			require.True(t, ok)
			require.Equal(t, expectedRole.Name, actualRole.Name)
			require.Equal(t, expectedRole.DisplayName, actualRole.DisplayName)
			require.Equal(t, expectedRole.Description, actualRole.Description)
			require.Equal(t, expectedRole.Permissions, actualRole.Permissions)
		}
	})

	t.Run("get all custom roles", func(t *testing.T) {
		// Enable all known custom role groups
		var expectedLen int
		for _, group := range model.AllCustomRoleGroups() {
			_, err := th.App.EnableCustomRoles(th.Context, group)
			expectedLen += len(model.MakeTunagCustomRoles(group))
			require.Nil(t, err)
		}

		roles, err := th.App.GetCustomRolesForGroup(th.Context, "")
		require.Nil(t, err)
		require.NotNil(t, roles)
		require.Len(t, roles, expectedLen)
	})

	t.Run("get non-existent group", func(t *testing.T) {
		roles, err := th.App.GetCustomRolesForGroup(th.Context, "non_existent_group")
		require.Nil(t, err)
		require.Empty(t, roles)
	})
}

func TestEnableCustomRoles(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	assertRolesMatch := func(t *testing.T, actualRoles []*model.Role, group string) {
		t.Helper()

		expectedRoleTemplates := model.MakeTunagCustomRoles(group)
		require.Len(t, actualRoles, len(expectedRoleTemplates))

		actualRolesMap := make(map[string]*model.Role)
		for _, role := range actualRoles {
			actualRolesMap[role.Name] = role
		}

		for name, expectedRole := range expectedRoleTemplates {
			actualRole, ok := actualRolesMap[name]
			require.True(t, ok)
			require.Equal(t, int64(0), actualRole.DeleteAt)
			require.Equal(t, expectedRole.Name, actualRole.Name)
			require.Equal(t, expectedRole.DisplayName, actualRole.DisplayName)
			require.ElementsMatch(t, expectedRole.Permissions, actualRole.Permissions)
		}
	}

	testGroup := model.CustomRolesUnofficial

	t.Run("create roles for the first time", func(t *testing.T) {
		roles, err := th.App.EnableCustomRoles(th.Context, testGroup)
		require.Nil(t, err)
		require.Len(t, roles, len(model.MakeTunagCustomRoles(testGroup)))

		// Verify they exist in the database
		dbRoles, err := th.App.GetRolesByNames(model.CustomRoleNamesForGroup(testGroup))
		require.Nil(t, err)
		require.Len(t, dbRoles, len(roles))
		assertRolesMatch(t, roles, testGroup)
	})

	t.Run("roles already exist and are active", func(t *testing.T) {
		_, err := th.App.EnableCustomRoles(th.Context, testGroup)
		require.Nil(t, err)

		// Call a second time
		roles, err := th.App.EnableCustomRoles(th.Context, testGroup)
		require.Nil(t, err)

		require.Len(t, roles, len(model.MakeTunagCustomRoles(testGroup)))
		assertRolesMatch(t, roles, testGroup)
	})

	t.Run("restore soft-deleted roles", func(t *testing.T) {
		// Enable roles
		roles, err := th.App.EnableCustomRoles(th.Context, testGroup)
		require.Nil(t, err)

		for _, role := range roles {
			// Soft-delete them
			_, err := th.App.DeleteRole(role.Id)
			require.Nil(t, err)

			// Verify it's deleted
			deletedRole, err := th.App.GetRole(role.Id)
			require.Nil(t, err)
			require.NotEqual(t, int64(0), deletedRole.DeleteAt)
		}

		// Call EnableCustomRoles again to trigger restore
		_, err = th.App.EnableCustomRoles(th.Context, testGroup)
		require.Nil(t, err)

		// Verify it's restored
		for _, role := range roles {
			restoredRole, err := th.App.GetRole(role.Id)
			require.Nil(t, err)
			require.Equal(t, int64(0), restoredRole.DeleteAt)
		}
	})
}

func TestDisableCustomRoles(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	testGroup := model.CustomRolesUnofficial

	// Enable the roles to ensure they exist
	_, err := th.App.EnableCustomRoles(th.Context, testGroup)
	require.Nil(t, err)

	// Confirm they exist and are active
	rolesBefore, err := th.App.GetRolesByNames(model.CustomRoleNamesForGroup(testGroup))
	require.Nil(t, err)
	for _, role := range rolesBefore {
		require.Equal(t, int64(0), role.DeleteAt)
	}

	err = th.App.DisableCustomRoles(th.Context, testGroup)
	require.Nil(t, err)

	// Verify they are soft-deleted
	rolesAfter, err := th.App.GetRolesByNames(model.CustomRoleNamesForGroup(testGroup))
	require.Nil(t, err)
	require.Len(t, rolesAfter, len(rolesBefore))
	for _, role := range rolesAfter {
		require.NotEqual(t, int64(0), role.DeleteAt)
	}
}
