// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestWithLegacyCallsTransport(t *testing.T) {
	t.Run("nil is passed through", func(t *testing.T) {
		assert.Nil(t, withLegacyCallsTransport(nil))
	})

	t.Run("calls ringing push without transport is promoted to VoIP on a copy", func(t *testing.T) {
		msg := &model.PushNotification{Type: model.PushTypeMessage, SubType: model.PushSubTypeCalls}

		got := withLegacyCallsTransport(msg)

		assert.Equal(t, model.PushTransportVoIP, got.Transport)
		assert.Equal(t, model.PushTransportStandard, msg.Transport, "the caller's notification must not be modified")
	})

	for _, tc := range []struct {
		name string
		msg  *model.PushNotification
	}{
		{"calls push that already chose VoIP", &model.PushNotification{Type: model.PushTypeMessage, SubType: model.PushSubTypeCalls, Transport: model.PushTransportVoIP}},
		{"calls clear push", &model.PushNotification{Type: model.PushTypeClear, SubType: model.PushSubTypeCalls}},
		{"chat message push", &model.PushNotification{Type: model.PushTypeMessage}},
	} {
		t.Run(tc.name+" is left untouched", func(t *testing.T) {
			before := *tc.msg
			got := withLegacyCallsTransport(tc.msg)
			assert.Same(t, tc.msg, got)
			assert.Equal(t, before, *got)
		})
	}
}

// Calls < v1.12 sends ringing pushes through the plugin API without a
// transport. They must reach the Engage Chat iOS app's "apple_voip" token,
// while Android keeps receiving them on its standard token.
func TestPluginAPISendPushNotificationLegacyCallsRouting(t *testing.T) {
	mainHelper.Parallel(t)

	const (
		iOSStandardToken = model.PushNotifyAppleReactNative + "-v2:standardtoken"
		iOSVoIPToken     = model.PushNotifyAppleVoip + ":voiptoken"
		androidToken     = model.PushNotifyAndroidReactNative + "-v2:androidtoken"
	)

	for _, tc := range []struct {
		name              string
		deviceId          string
		voIPDeviceId      string
		sessionProps      map[string]string
		expectedPlatform  string
		expectedDeviceID  string
		expectedTransport model.PushTransport
	}{
		{
			name:              "iOS with a VoIP token rings through apple_voip",
			deviceId:          iOSStandardToken,
			voIPDeviceId:      iOSVoIPToken,
			expectedPlatform:  model.PushNotifyAppleVoip,
			expectedDeviceID:  "voiptoken",
			expectedTransport: model.PushTransportVoIP,
		},
		{
			name:              "iOS with an invalidated VoIP token falls back to the standard token",
			deviceId:          iOSStandardToken,
			voIPDeviceId:      iOSVoIPToken,
			sessionProps:      map[string]string{model.SessionPropLastRemovedVoIPDeviceId: iOSVoIPToken},
			expectedPlatform:  model.PushNotifyAppleReactNative + "-v2",
			expectedDeviceID:  "standardtoken",
			expectedTransport: model.PushTransportStandard,
		},
		{
			name:              "Android rings through its standard token",
			deviceId:          androidToken,
			expectedPlatform:  model.PushNotifyAndroidReactNative + "-v2",
			expectedDeviceID:  "androidtoken",
			expectedTransport: model.PushTransportStandard,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			handler := &testPushNotificationHandler{t: t, behavior: "simple"}
			pushServer := httptest.NewServer(http.HandlerFunc(handler.handleReq))
			defer pushServer.Close()

			th.App.UpdateConfig(func(cfg *model.Config) {
				*cfg.EmailSettings.PushNotificationServer = pushServer.URL
			})

			_, appErr := th.App.CreateSession(th.Context, &model.Session{
				UserId:       th.BasicUser.Id,
				DeviceId:     tc.deviceId,
				VoIPDeviceId: tc.voIPDeviceId,
				Props:        tc.sessionProps,
				ExpiresAt:    model.GetMillis() + 100000,
			})
			require.Nil(t, appErr)

			api := th.SetupPluginAPI()
			appErr = api.SendPushNotification(&model.PushNotification{
				Type:    model.PushTypeMessage,
				SubType: model.PushSubTypeCalls,
			}, th.BasicUser.Id)
			require.Nil(t, appErr)

			require.Eventually(t, func() bool {
				return len(handler.notifications()) == 1
			}, 2*time.Second, 10*time.Millisecond, "expected exactly one push notification")

			sent := handler.notifications()[0]
			assert.Equal(t, tc.expectedPlatform, sent.Platform)
			assert.Equal(t, tc.expectedDeviceID, sent.DeviceId)
			assert.Equal(t, tc.expectedTransport, sent.Transport)
			assert.Equal(t, model.PushSubTypeCalls, sent.SubType)
		})
	}

	t.Run("a removed apple_voip token is tracked separately from the standard token", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		handler := &testPushNotificationHandler{t: t, behavior: "always_remove"}
		pushServer := httptest.NewServer(http.HandlerFunc(handler.handleReq))
		defer pushServer.Close()

		th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.EmailSettings.PushNotificationServer = pushServer.URL
		})

		session, appErr := th.App.CreateSession(th.Context, &model.Session{
			UserId:       th.BasicUser.Id,
			DeviceId:     iOSStandardToken,
			VoIPDeviceId: iOSVoIPToken,
			ExpiresAt:    model.GetMillis() + 100000,
		})
		require.Nil(t, appErr)

		api := th.SetupPluginAPI()
		appErr = api.SendPushNotification(&model.PushNotification{
			Type:    model.PushTypeMessage,
			SubType: model.PushSubTypeCalls,
		}, th.BasicUser.Id)
		require.Nil(t, appErr)

		require.Eventually(t, func() bool {
			return len(handler.notifications()) == 1
		}, 2*time.Second, 10*time.Millisecond)

		updated, appErr := th.App.GetSession(session.Token)
		require.Nil(t, appErr)
		assert.Equal(t, iOSVoIPToken, updated.Props[model.SessionPropLastRemovedVoIPDeviceId])
		assert.Empty(t, updated.Props[model.SessionPropLastRemovedDeviceId], "chat pushes must keep using the standard token")
	})
}
