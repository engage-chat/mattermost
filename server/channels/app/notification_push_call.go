// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// withLegacyCallsTransport routes a Calls ringing push over the VoIP
// transport when the sender did not choose one.
//
// The Calls plugin only sets PushNotification.Transport from v1.12.0 on.
// Older versions (Engage Chat still bundles v1.11) send ringing pushes with
// the standard transport, which upstream routing never delivers to the VoIP
// token, so CallKit would not ring on iOS. Pushes that already carry a
// transport are left untouched, as are non-message pushes: iOS terminates an
// app that receives a VoIP push without reporting a call, so clear pushes
// must stay on the standard transport.
func withLegacyCallsTransport(msg *model.PushNotification) *model.PushNotification {
	if msg == nil ||
		msg.Type != model.PushTypeMessage ||
		msg.SubType != model.PushSubTypeCalls ||
		msg.Transport != model.PushTransportStandard {
		return msg
	}

	voIPMsg := msg.DeepCopy()
	voIPMsg.Transport = model.PushTransportVoIP
	return voIPMsg
}
