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

func TestGetChannelAccessible(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// Enable custom roles
	_, appErr := th.App.EnableCustomRoles(th.Context, []string{model.SystemEngageAdmin, model.TeamEngageAdmin})
	require.Nil(t, appErr)

	// Setup exception user with SystemEngageAdmin
	exceptionUser := th.CreateUser()
	th.LinkUserToTeam(exceptionUser, th.BasicTeam)
	_, appErr = th.App.UpdateUserRoles(th.Context, exceptionUser.Id, model.SystemUserRoleId+" "+model.SystemEngageAdmin, false)
	require.Nil(t, appErr)

	// DM without exception member
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, th.BasicUser.Id, th.BasicUser2.Id)
	require.Nil(t, appErr)

	// DM with exception member
	dmWithException, appErr := th.App.GetOrCreateDirectChannel(th.Context, th.BasicUser.Id, exceptionUser.Id)
	require.Nil(t, appErr)

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		client := th.CreateClient()
		_, resp, err := client.GetChannelAccessible(context.Background(), th.BasicChannel.Id)
		require.Error(t, err)
		CheckUnauthorizedStatus(t, resp)
	})

	t.Run("invalid channel ID returns 400", func(t *testing.T) {
		_, resp, err := th.Client.GetChannelAccessible(context.Background(), "invalidid")
		require.Error(t, err)
		CheckBadRequestStatus(t, resp)
	})

	t.Run("non-existent channel ID returns error", func(t *testing.T) {
		_, resp, err := th.Client.GetChannelAccessible(context.Background(), model.NewId())
		require.Error(t, err)
		CheckNotFoundStatus(t, resp)
	})

	t.Run("public channel is always accessible", func(t *testing.T) {
		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), th.BasicChannel.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.True(t, accessible)
	})

	t.Run("DM channel is accessible when no restriction", func(t *testing.T) {
		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), dmChannel.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.True(t, accessible)
	})

	t.Run("under restriction: DM without exception is not accessible", func(t *testing.T) {
		th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), dmChannel.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.False(t, accessible)
	})

	t.Run("under restriction: DM with exception member is accessible", func(t *testing.T) {
		th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), dmWithException.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.True(t, accessible)
	})
}
