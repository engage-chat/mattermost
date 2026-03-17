// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestEnableTunagCustomRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	testGroup := model.CustomRolesUnofficial

	// Clean up any existing roles to ensure a clean slate for the test.
	_, err := th.SystemAdminClient.DisableTunagCustomRoles(context.Background(), testGroup)
	require.NoError(t, err)
	// And ensure it's cleaned up after the test finishes too.
	defer func() {
		_, err := th.SystemAdminClient.DisableTunagCustomRoles(context.Background(), testGroup)
		require.NoError(t, err)
	}()

	t.Run("as regular user", func(t *testing.T) {
		_, resp, err := th.Client.EnableTunagCustomRoles(context.Background(), testGroup)
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as admin user", func(t *testing.T) {
		returnedRoles, resp, err := th.SystemAdminClient.EnableTunagCustomRoles(context.Background(), testGroup)
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		expectedRolesMap := model.MakeTunagCustomRoles(testGroup)
		require.Len(t, returnedRoles, len(expectedRolesMap))

		for _, returnedRole := range returnedRoles {
			expectedRole, ok := expectedRolesMap[returnedRole.Name]
			require.True(t, ok)

			assert.Equal(t, expectedRole.DisplayName, returnedRole.DisplayName)
			assert.ElementsMatch(t, expectedRole.Permissions, returnedRole.Permissions)
		}
	})
}

func TestDisableTunagCustomRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	testGroup := model.CustomRolesUnofficial

	// Enable roles first to ensure we have something to disable.
	_, _, err := th.SystemAdminClient.EnableTunagCustomRoles(context.Background(), testGroup)
	require.NoError(t, err)

	t.Run("as regular user", func(t *testing.T) {
		resp, err := th.Client.DisableTunagCustomRoles(context.Background(), testGroup)
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as admin user", func(t *testing.T) {
		resp, err := th.SystemAdminClient.DisableTunagCustomRoles(context.Background(), testGroup)
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		// Verify they are soft-deleted in the DB
		roleNames := model.CustomRoleNamesForGroup(testGroup)
		dbRoles, appErr := th.App.GetRolesByNames(roleNames)
		require.Nil(t, appErr)
		require.Len(t, dbRoles, len(roleNames))

		for _, role := range dbRoles {
			require.NotEqual(t, int64(0), role.DeleteAt)
		}
	})
}

func TestGetTunagCustomRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	testGroup := model.CustomRolesUnofficial

	// Enable roles first to ensure we have something to get.
	_, _, err := th.SystemAdminClient.EnableTunagCustomRoles(context.Background(), testGroup)
	require.NoError(t, err)
	defer func() {
		_, err := th.SystemAdminClient.DisableTunagCustomRoles(context.Background(), testGroup)
		require.NoError(t, err)
	}()

	t.Run("as regular user", func(t *testing.T) {
		_, resp, err := th.Client.GetTunagCustomRoles(context.Background(), testGroup)
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as admin user", func(t *testing.T) {
		t.Run("get existing group", func(t *testing.T) {
			roles, resp, err := th.SystemAdminClient.GetTunagCustomRoles(context.Background(), testGroup)
			require.NoError(t, err)
			CheckOKStatus(t, resp)

			expectedRolesMap := model.MakeTunagCustomRoles(testGroup)
			require.Len(t, roles, len(expectedRolesMap))

			for _, returnedRole := range roles {
				expectedRole, ok := expectedRolesMap[returnedRole.Name]
				require.True(t, ok)

				assert.Equal(t, expectedRole.DisplayName, returnedRole.DisplayName)
				assert.ElementsMatch(t, expectedRole.Permissions, returnedRole.Permissions)
			}
		})

		t.Run("get non-existent group", func(t *testing.T) {
			roles, resp, err := th.SystemAdminClient.GetTunagCustomRoles(context.Background(), "non_existent_group")
			require.NoError(t, err)
			CheckOKStatus(t, resp)
			assert.Empty(t, roles)
		})
	})
}

func TestGetAllTunagCustomRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	testGroup := model.CustomRolesUnofficial

	// Enable roles first to ensure we have something to get.
	_, _, err := th.SystemAdminClient.EnableTunagCustomRoles(context.Background(), testGroup)
	require.NoError(t, err)
	defer func() {
		_, err := th.SystemAdminClient.DisableTunagCustomRoles(context.Background(), testGroup)
		require.NoError(t, err)
	}()

	t.Run("as regular user", func(t *testing.T) {
		_, resp, err := th.Client.GetAllTunagCustomRoles(context.Background())
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as admin user", func(t *testing.T) {
		roles, resp, err := th.SystemAdminClient.GetAllTunagCustomRoles(context.Background())
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		expectedRoles := model.MakeTunagCustomRoles(testGroup)
		assert.Len(t, roles, len(expectedRoles))
	})
}
