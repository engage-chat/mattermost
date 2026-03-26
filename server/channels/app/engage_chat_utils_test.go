// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestIsChannelAccessible(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// 1. --- Roles and Test Users Setup ---
	_, err := th.App.EnableCustomRoles(th.Context, []string{model.SystemEngageAdmin, model.TeamEngageAdmin})
	require.Nil(t, err)

	restrictedUser := th.CreateUser()
	th.LinkUserToTeam(restrictedUser, th.BasicTeam)

	botRec := th.CreateBot()
	botUser, err := th.App.GetUser(botRec.UserId)
	require.Nil(t, err)

	// User with SystemEngageAdmin role
	systemExceptionUser := th.CreateUser()
	th.LinkUserToTeam(systemExceptionUser, th.BasicTeam)
	_, err = th.App.UpdateUserRoles(th.Context, systemExceptionUser.Id, model.SystemUserRoleId+" "+model.SystemEngageAdmin, false)
	require.Nil(t, err)

	// User with TeamEngageAdmin role
	teamExceptionUser := th.CreateUser()
	th.LinkUserToTeam(teamExceptionUser, th.BasicTeam)
	_, err = th.App.UpdateTeamMemberRoles(th.Context, th.BasicTeam.Id, teamExceptionUser.Id, model.TeamUserRoleId+" "+model.TeamEngageAdmin)
	require.Nil(t, err)

	// User with no special roles
	regularUser := th.CreateUser()
	th.LinkUserToTeam(regularUser, th.BasicTeam)

	session, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: restrictedUser.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)
	ctxWithSession := th.Context.WithSession(session)

	// 2. --- Channels Setup ---
	// Setup for official channel test
	th.App.ResetIntegrationAdminUsernameCache()
	t.Setenv("INTEGRATION_ADMIN_USERNAME", th.SystemAdminUser.Username)
	defer th.App.ResetIntegrationAdminUsernameCache()

	officialChannel := th.CreateChannel(th.Context, th.BasicTeam)
	officialChannel.CreatorId = th.SystemAdminUser.Id // Make it official
	_, err = th.App.UpdateChannel(th.Context, officialChannel)
	require.Nil(t, err)
	th.AddUserToChannel(restrictedUser, officialChannel)

	publicChannel, err := th.App.CreateChannel(th.Context, &model.Channel{
		TeamId: th.BasicTeam.Id, DisplayName: "Public Channel", Name: "public-channel-" + model.NewId(), Type: model.ChannelTypeOpen,
	}, false)
	require.Nil(t, err)
	th.AddUserToChannel(restrictedUser, publicChannel)

	// --- Channels for permission tests ---
	dmChannel, _ := th.App.GetOrCreateDirectChannel(ctxWithSession, restrictedUser.Id, regularUser.Id)
	dmWithException, _ := th.App.GetOrCreateDirectChannel(ctxWithSession, restrictedUser.Id, systemExceptionUser.Id)
	gmChannel := th.CreateGroupChannel(th.Context, restrictedUser, regularUser)
	gmWithException := th.CreateGroupChannel(th.Context, restrictedUser, systemExceptionUser)
	privateChannel := th.CreatePrivateChannel(th.Context, th.BasicTeam)
	th.AddUserToChannel(restrictedUser, privateChannel)
	th.AddUserToChannel(regularUser, privateChannel)
	privateWithException := th.CreatePrivateChannel(th.Context, th.BasicTeam)
	th.AddUserToChannel(restrictedUser, privateWithException)
	th.AddUserToChannel(teamExceptionUser, privateWithException)

	// 3. --- Run Test Cases ---
	t.Run("General Access Scenarios", func(t *testing.T) {
		testCases := []struct {
			name          string
			channelID     string
			userID        string
			expected      bool
			expectedError bool
		}{
			{"Invalid channel ID should return an error", "invalid-channel-id", restrictedUser.Id, false, true},
			{"Invalid user ID should return an error", publicChannel.Id, "invalid-user-id", false, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				accessible, err := th.App.IsChannelAccessible(ctxWithSession, tc.channelID, tc.userID)
				if tc.expectedError {
					require.NotNil(t, err)
				} else {
					require.Nil(t, err)
				}
				require.Equal(t, tc.expected, accessible)
			})
		}
	})

	t.Run("Situation no restriction", func(t *testing.T) {
		testCases := []struct {
			name     string
			channel  *model.Channel
			userID   string
			expected bool
		}{
			{"DM Channel", dmChannel, restrictedUser.Id, true},
			{"Group Channel", gmChannel, restrictedUser.Id, true},
			{"Private Channel", privateChannel, restrictedUser.Id, true},
			{"Bot user should always be accessible", publicChannel, botUser.Id, true},
			{"Official channel should always be accessible", officialChannel, restrictedUser.Id, true},
			{"Public channel should always be accessible", publicChannel, restrictedUser.Id, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				accessible, err := th.App.IsChannelAccessible(ctxWithSession, tc.channel.Id, tc.userID)
				require.Nil(t, err)
				require.Equal(t, tc.expected, accessible)
			})
		}
	})

	t.Run("Situation under restriction", func(t *testing.T) {
		// Remove permissions for the restricted user (setup once for all sub-scenarios)
		th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		th.RemovePermissionFromRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
		th.RemovePermissionFromRole(model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
		defer func() {
			th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
			th.AddPermissionToRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
			th.AddPermissionToRole(model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
		}()

		t.Run("without exception", func(t *testing.T) {
			testCases := []struct {
				name     string
				channel  *model.Channel
				userID   string
				expected bool
			}{
				{"DM without exception should NOT be accessible", dmChannel, restrictedUser.Id, false},
				{"Group without exception should NOT be accessible", gmChannel, restrictedUser.Id, false},
				{"Private without exception should NOT be accessible", privateChannel, restrictedUser.Id, false},
				{"Official channel should always be accessible", officialChannel, restrictedUser.Id, true},
				{"Public channel should always be accessible", publicChannel, restrictedUser.Id, true},
				{"Bot user should always be accessible", publicChannel, botUser.Id, true},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					accessible, err := th.App.IsChannelAccessible(ctxWithSession, tc.channel.Id, tc.userID)
					require.Nil(t, err)
					require.Equal(t, tc.expected, accessible)
				})
			}
		})

		t.Run("with exception", func(t *testing.T) {
			testCases := []struct {
				name     string
				channel  *model.Channel
				userID   string
				expected bool
			}{
				{"DM with exception member should be accessible", dmWithException, restrictedUser.Id, true},
				{"Group with exception member should be accessible", gmWithException, restrictedUser.Id, true},
				{"Private with exception member should be accessible", privateWithException, restrictedUser.Id, true},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					accessible, err := th.App.IsChannelAccessible(ctxWithSession, tc.channel.Id, tc.userID)
					require.Nil(t, err)
					require.Equal(t, tc.expected, accessible)
				})
			}
		})
	})
}
