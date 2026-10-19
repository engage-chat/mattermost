// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

// The Engage Chat mobile app attaches its PushKit token as "apple_voip:<token>"
// next to the standard "apple_rn[beta]-v2:<token>" device id.
func TestEngageChatAttachVoIPDeviceId(t *testing.T) {
	mainHelper.Parallel(t)
	th := Setup(t).InitBasic(t)

	client := th.CreateClient()
	th.LoginBasicWithClient(t, client)

	resetBothTokens := func(session *model.Session) {
		require.NoError(t, th.Server.Store().Session().UpdateDeviceId(session.Id, "", "", session.ExpiresAt))
		th.App.ClearSessionCacheForUser(session.UserId)
	}

	for _, platform := range []string{model.PushNotifyAppleReactNative, model.PushNotifyAppleReactNative + "beta"} {
		t.Run("iOS app tokens are stored in both columns for "+platform, func(t *testing.T) {
			session, _ := th.App.GetSession(client.AuthToken)
			defer resetBothTokens(session)

			deviceID := platform + "-v2:standard-token"
			voIPID := model.PushNotifyAppleVoip + ":voip-token"
			_, err := client.AttachDeviceProps(context.Background(), map[string]string{
				"device_id":      deviceID,
				"voip_device_id": voIPID,
			})
			require.NoError(t, err)

			storeSession, err := th.Server.Store().Session().Get(th.Context, session.Id)
			require.NoError(t, err)
			assert.Equal(t, deviceID, storeSession.DeviceId)
			assert.Equal(t, voIPID, storeSession.VoIPDeviceId)
		})
	}

	t.Run("Android app token leaves the VoIP column empty", func(t *testing.T) {
		session, _ := th.App.GetSession(client.AuthToken)
		defer resetBothTokens(session)

		deviceID := model.PushNotifyAndroidReactNative + "-v2:standard-token"
		_, err := client.AttachDeviceProps(context.Background(), map[string]string{"device_id": deviceID})
		require.NoError(t, err)

		storeSession, err := th.Server.Store().Session().Get(th.Context, session.Id)
		require.NoError(t, err)
		assert.Equal(t, deviceID, storeSession.DeviceId)
		assert.Empty(t, storeSession.VoIPDeviceId)
	})

	t.Run("Attaching apple_voip alone preserves an existing device_id", func(t *testing.T) {
		session, _ := th.App.GetSession(client.AuthToken)
		defer resetBothTokens(session)

		existingDeviceID := model.PushNotifyAppleReactNative + "-v2:existing-standard"
		require.NoError(t, th.Server.Store().Session().UpdateDeviceId(session.Id, existingDeviceID, "", session.ExpiresAt))
		th.App.ClearSessionCacheForUser(session.UserId)

		voIPID := model.PushNotifyAppleVoip + ":new-voip"
		_, err := client.AttachDeviceProps(context.Background(), map[string]string{"voip_device_id": voIPID})
		require.NoError(t, err)

		storeSession, err := th.Server.Store().Session().Get(th.Context, session.Id)
		require.NoError(t, err)
		assert.Equal(t, existingDeviceID, storeSession.DeviceId)
		assert.Equal(t, voIPID, storeSession.VoIPDeviceId)
	})

	t.Run("apple_voip is rejected as a standard device_id", func(t *testing.T) {
		session, _ := th.App.GetSession(client.AuthToken)
		defer resetBothTokens(session)

		res, err := client.AttachDeviceProps(context.Background(), map[string]string{"device_id": model.PushNotifyAppleVoip + ":voip-token"})
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		storeSession, err := th.Server.Store().Session().Get(th.Context, session.Id)
		require.NoError(t, err)
		assert.Empty(t, storeSession.DeviceId)
		assert.Empty(t, storeSession.VoIPDeviceId)
	})

	t.Run("Attaching apple_voip revokes another session holding the same VoIP token", func(t *testing.T) {
		currentSession, _ := th.App.GetSession(client.AuthToken)
		defer resetBothTokens(currentSession)

		voIPID := model.PushNotifyAppleVoip + ":shared-voip"
		priorSession, appErr := th.App.CreateSession(th.Context, &model.Session{
			UserId:       currentSession.UserId,
			VoIPDeviceId: voIPID,
			ExpiresAt:    currentSession.ExpiresAt,
		})
		require.Nil(t, appErr)

		_, err := client.AttachDeviceProps(context.Background(), map[string]string{"voip_device_id": voIPID})
		require.NoError(t, err)

		_, appErr = th.App.GetSessionById(th.Context, priorSession.Id)
		require.NotNil(t, appErr, "prior session with the same VoIP token must be revoked")

		_, appErr = th.App.GetSessionById(th.Context, currentSession.Id)
		require.Nil(t, appErr, "current session must survive its own attach")
	})
}
