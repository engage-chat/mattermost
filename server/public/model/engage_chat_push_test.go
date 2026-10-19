// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngageChatVoIPDeviceIdPlatform(t *testing.T) {
	testCases := []struct {
		Description string
		Value       string
		VoIP        bool
		Standard    bool
	}{
		{"apple_voip token is a valid VoIP device id", PushNotifyAppleVoip + ":abcd", true, false},
		{"apple_voip without token is invalid", PushNotifyAppleVoip + ":", false, false},
		{"apple_voip without separator is invalid", PushNotifyAppleVoip, false, false},
		{"standard ids sent by the app remain valid", PushNotifyAppleReactNative + "-v2:abcd", true, true},
		{"beta standard ids sent by the app remain valid", PushNotifyAppleReactNative + "beta-v2:abcd", true, true},
		{"android standard ids remain standard only", PushNotifyAndroidReactNative + "-v2:abcd", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.VoIP, IsValidVoIPDeviceId(tc.Value))
			assert.Equal(t, tc.Standard, IsValidStandardDeviceId(tc.Value))
		})
	}
}
