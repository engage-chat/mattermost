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

func TestEnableCustomRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	roleNames := model.AllCustomRoleNames()

	// Clean up any existing roles to ensure a clean slate for the test.
	_, err := th.SystemAdminClient.DisableCustomRoles(context.Background(), roleNames)
	require.NoError(t, err)
	// And ensure it's cleaned up after the test finishes too.
	defer func() {
		_, err := th.SystemAdminClient.DisableCustomRoles(context.Background(), roleNames)
		require.NoError(t, err)
	}()

	t.Run("as regular user", func(t *testing.T) {
		_, resp, err := th.Client.EnableCustomRoles(context.Background(), roleNames)
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as admin user", func(t *testing.T) {
		returnedRoles, resp, err := th.SystemAdminClient.EnableCustomRoles(context.Background(), roleNames)
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		expectedRolesMap := model.MakeAllCustomRoleTemplates()
		require.Len(t, returnedRoles, len(expectedRolesMap))

		for _, returnedRole := range returnedRoles {
			expectedRole, ok := expectedRolesMap[returnedRole.Name]
			require.True(t, ok)

			assert.Equal(t, expectedRole.DisplayName, returnedRole.DisplayName)
			assert.ElementsMatch(t, expectedRole.Permissions, returnedRole.Permissions)
		}
	})
}

func TestDisableCustomRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	roleNames := model.AllCustomRoleNames()

	// Enable roles first to ensure we have something to disable.
	_, _, err := th.SystemAdminClient.EnableCustomRoles(context.Background(), roleNames)
	require.NoError(t, err)

	t.Run("as regular user", func(t *testing.T) {
		resp, err := th.Client.DisableCustomRoles(context.Background(), roleNames)
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as admin user", func(t *testing.T) {
		resp, err := th.SystemAdminClient.DisableCustomRoles(context.Background(), roleNames)
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		// Verify they are soft-deleted in the DB
		dbRoles, appErr := th.App.GetRolesByNames(roleNames)
		require.Nil(t, appErr)
		require.Len(t, dbRoles, len(roleNames))

		for _, role := range dbRoles {
			require.NotEqual(t, int64(0), role.DeleteAt)
		}
	})
}
