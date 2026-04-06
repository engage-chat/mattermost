// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Permissions} from 'mattermost-redux/constants';
import {getChannel} from 'mattermost-redux/selectors/entities/channels';
import {haveIChannelPermission, haveICurrentTeamPermission} from 'mattermost-redux/selectors/entities/roles';

import store from 'stores/redux_store';

import {fetchChannelAccessible} from 'actions/engage_chat';

export const isAvailableUnofficialChannel = (channelId: string): boolean => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const state: any = store.getState();

    // If the API result is cached in Redux, use it in preference to the permission-based check.
    if (state.engageChat && channelId in state.engageChat.channelAccessible) {
        return state.engageChat.channelAccessible[channelId];
    }

    // Check the local permission first as a fast path.
    // If permission is granted, return true immediately without calling the API.
    const channel = getChannel(state, channelId);
    if (channel) {
        let permission: string | undefined;

        switch (channel.type) {
        case 'P':
            permission = Permissions.CREATE_PRIVATE_CHANNEL;
            break;
        case 'D':
            permission = Permissions.CREATE_DIRECT_CHANNEL;
            break;
        case 'G':
            permission = Permissions.CREATE_GROUP_CHANNEL;
            break;
        default:
            return true;
        }

        if (haveIChannelPermission(state, channel.team_id, channel.id, permission)) {
            return true;
        }
    }

    // Permission denied or channel not found locally — fire an API request to get the
    // authoritative result. Return false (inaccessible) until the API responds.
    // On API failure, the channel remains inaccessible until the user reloads the page.
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    store.dispatch(fetchChannelAccessible(channelId) as any);
    return false;
};

export const isAvailableDMGMChannel = (): boolean => {
    const state = store.getState();
    return haveICurrentTeamPermission(state, Permissions.CREATE_DIRECT_CHANNEL) && haveICurrentTeamPermission(state, Permissions.CREATE_GROUP_CHANNEL);
};
