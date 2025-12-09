// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestChannelPermissionChecksForReactions(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// Save original environment variable
	originalValue := os.Getenv("INTEGRATION_ADMIN_USERNAME")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("INTEGRATION_ADMIN_USERNAME")
		} else {
			os.Setenv("INTEGRATION_ADMIN_USERNAME", originalValue)
		}
	}()

	// Create users
	user1 := th.BasicUser
	user2 := th.BasicUser2
	adminUser := th.SystemAdminUser 
	bot := th.CreateBot()
	botUser, appErr := th.App.GetUser(bot.UserId)
	require.Nil(t, appErr)

	// Create sessions
	user1Session, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: user1.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)

	botSession, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: botUser.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)

	adminSession, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: adminUser.Id,
		Roles:  model.SystemAdminRoleId,
	})
	require.Nil(t, err)

	// Create DM channel (with permissions)
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, user1.Id, user2.Id)
	require.Nil(t, appErr)
	require.NotNil(t, dmChannel)

	// Create GM channel (with permissions)
	user3 := th.CreateUser()
	gmChannel, appErr := th.App.createGroupChannel(th.Context, []string{user1.Id, user2.Id, user3.Id}, user1.Id)
	require.Nil(t, appErr)
	require.NotNil(t, gmChannel)

	// Create official channel
	channel1 := &model.Channel{
		DisplayName: "test channel",
		Name:		"test-channel-"+model.NewId(),
		Type:		model.ChannelTypePrivate,
		TeamId:		th.BasicTeam.Id,
		CreatorId:	adminUser.Id,
	}
	officialChannel, appErr := th.App.CreateChannel(th.Context, channel1, false)
	require.Nil(t, appErr)
	require.NotNil(t, officialChannel)

	// Create unofficial channel
	channel2 := &model.Channel{
		DisplayName: "test channel",
		Name:		"test-channel-"+model.NewId(),
		Type:		model.ChannelTypePrivate,
		TeamId:		th.BasicTeam.Id,
		CreatorId:	user2.Id,
	}
	unofficialChannel, appErr := th.App.CreateChannel(th.Context, channel2, true)
	require.Nil(t, appErr)
	require.NotNil(t, unofficialChannel)

	// Create posts in DM and GM for reaction tests
	dmPost := th.CreatePost(dmChannel)
	gmPost := th.CreatePost(gmChannel)
	oChannelPost := th.CreatePost(officialChannel)
	uoChannelPost := th.CreatePost(unofficialChannel)

	// Now remove permissions for testing
	th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

	th.RemovePermissionFromRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)

	th.RemovePermissionFromRole(model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)

	t.Run("SaveReactionForPost denied in DM without permission", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:    user1.Id,
			PostId:    dmPost.Id,
			EmojiName: "smile",
		}

		ctxWithSession := th.Context.WithSession(user1Session)
		_, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.NotNil(t, appErr)
		assert.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("SaveReactionForPost denied in GM without permission", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:    user1.Id,
			PostId:    gmPost.Id,
			EmojiName: "smile",
		}

		ctxWithSession := th.Context.WithSession(user1Session)
		_, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.NotNil(t, appErr)
		assert.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("Bot can add reaction in DM without permission", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:    botUser.Id,
			PostId:    dmPost.Id,
			EmojiName: "robot",
		}

		ctxWithSession := th.Context.WithSession(botSession)
		createdReaction, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.Nil(t, appErr)
		require.NotNil(t, createdReaction)
	})

	t.Run("System admin can add reaction in DM without permission", func(t *testing.T) {
		adminUser := th.SystemAdminUser

		reaction := &model.Reaction{
			UserId:    adminUser.Id,
			PostId:    dmPost.Id,
			EmojiName: "star",
		}

		ctxWithSession := th.Context.WithSession(adminSession)
		createdReaction, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.Nil(t, appErr)
		require.NotNil(t, createdReaction)
	})

	t.Run("SaveReactionForPost is performed within official channel", func(t *testing.T) {
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUser.Username)
		resetIntegrationAdminUsernameForTesting()

		reaction := &model.Reaction{
			UserId:		user1.Id,
			PostId:		oChannelPost.Id,
			EmojiName:	"smile",
		}

		ctxWithSession := th.Context.WithSession(user1Session)
		createdReaction, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.Nil(t, appErr)
		require.NotNil(t, createdReaction)
	})

	t.Run("SaveReactionForPost denied in unofficial channel without permisson", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:		user1.Id,
			PostId:		uoChannelPost.Id,
			EmojiName:	"smile",
		}

		ctxWithSession := th.Context.WithSession(user1Session)
		_, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.NotNil(t, appErr)
		require.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})}
