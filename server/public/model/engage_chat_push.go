// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

// PushNotifyAppleVoip is the platform prefix the Engage Chat mobile app uses
// for its PushKit (VoIP) token, e.g. "apple_voip:<token>". Our push proxy
// picks the VoIP APNs path from this platform rather than from the upstream
// Transport field, so the app keeps sending it instead of upstream's
// "apple_rn[beta]-v2:" form. It is accepted as a VoIPDeviceId only, never
// as a standard DeviceId, so chat pushes are never routed to a VoIP token.
const PushNotifyAppleVoip = "apple_voip"
