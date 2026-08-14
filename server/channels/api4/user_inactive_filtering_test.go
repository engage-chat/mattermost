// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestUserInactiveFiltering(t *testing.T) {
	t.Run("mention autocomplete excludes deactivated users", func(t *testing.T) {
		mainHelper.Parallel(t)
		th := Setup(t).InitBasic(t)

		user := th.BasicUser2
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.EnableUserDeactivation = true })

		// Add user to channel
		th.LinkUserToTeam(t, user, th.BasicTeam)
		th.AddUserToChannel(t, user, th.BasicChannel)

		// Before user deactivation: verify user appears in autocomplete
		rusers, _, err := th.Client.AutocompleteUsersInChannel(
			context.Background(),
			th.BasicTeam.Id,
			th.BasicChannel.Id,
			user.Username[:3], // Search using first 3 characters of username
			model.UserSearchDefaultLimit,
			"",
		)
		require.NoError(t, err)

		// Before deactivation: verify user is included in results
		found := false
		for _, u := range rusers.Users {
			if u.Id == user.Id {
				found = true
				break
			}
		}
		require.True(t, found, "active user should appear in autocomplete results")

		// Deactivate user
		_, err = th.SystemAdminClient.UpdateUserActive(context.Background(), user.Id, false)
		require.NoError(t, err)

		// After deactivation: verify user does not appear in autocomplete
		rusers, _, err = th.Client.AutocompleteUsersInChannel(
			context.Background(),
			th.BasicTeam.Id,
			th.BasicChannel.Id,
			user.Username[:3],
			model.UserSearchDefaultLimit,
			"",
		)
		require.NoError(t, err)

		// After deactivation: verify user is not included in results
		found = false
		for _, u := range rusers.Users {
			if u.Id == user.Id {
				found = true
				break
			}
		}
		require.False(t, found, "deactivated user should not appear in autocomplete results")

		// Reactivate user
		_, err = th.SystemAdminClient.UpdateUserActive(context.Background(), user.Id, true)
		require.NoError(t, err)

		// After reactivation: verify user appears in autocomplete again
		rusers, _, err = th.Client.AutocompleteUsersInChannel(
			context.Background(),
			th.BasicTeam.Id,
			th.BasicChannel.Id,
			user.Username[:3],
			model.UserSearchDefaultLimit,
			"",
		)
		require.NoError(t, err)

		// After reactivation: verify user is included in results
		found = false
		for _, u := range rusers.Users {
			if u.Id == user.Id {
				found = true
				break
			}
		}
		require.True(t, found, "reactivated user should appear in autocomplete results again")
	})
}
