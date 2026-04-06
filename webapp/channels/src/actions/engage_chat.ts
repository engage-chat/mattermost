// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Client4} from 'mattermost-redux/client';

import type {ActionFuncAsync} from 'types/store';

import {RECEIVED_CHANNEL_ACCESSIBLE} from 'reducers/engage_chat';

// Prevent duplicate in-flight requests for the same channelId
const pendingRequests = new Set<string>();

export function fetchChannelAccessible(channelId: string): ActionFuncAsync<void> {
    return async (dispatch) => {
        if (pendingRequests.has(channelId)) {
            return;
        }
        pendingRequests.add(channelId);
        try {
            const response = await fetch(
                `${Client4.getUrl()}/api/v4/engage_chat/channels/${channelId}/accessible`,
                {
                    method: 'GET',
                    headers: {
                        Authorization: `Bearer ${Client4.getToken()}`,
                        'X-Requested-With': 'XMLHttpRequest',
                    },
                },
            );
            if (!response.ok) {
                return;
            }
            const data: {is_accessible: boolean} = await response.json();
            dispatch({
                type: RECEIVED_CHANNEL_ACCESSIBLE,
                channelId,
                accessible: data.is_accessible,
            });
        } catch {
            // On API failure, cache false so the channel remains inaccessible
            // until the user reloads the page (no automatic retry).
            dispatch({
                type: RECEIVED_CHANNEL_ACCESSIBLE,
                channelId,
                accessible: false,
            });
        } finally {
            pendingRequests.delete(channelId);
        }
    };
}
