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

	roleNames := []string{model.SystemEngageAdmin, model.TeamEngageAdmin}

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

	t.Run("via local client", func(t *testing.T) {
		returnedRoles, resp, err := th.LocalClient.EnableCustomRoles(context.Background(), roleNames)
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
