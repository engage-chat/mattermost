// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Channel} from '@mattermost/types/channels';

import {getUser} from 'mattermost-redux/selectors/entities/users';

import store from 'stores/redux_store';

import type {GlobalState} from 'types/store';

/**
 * Regex pattern for official tunag integration admin usernames.
 * Pattern: tunag-{5digits}-{lowercase_alphanumeric_hyphens}-admin
 * Subdomain rules: lowercase letters, numbers, hyphens
 * Example: tunag-00002-stmn-admin
 */
const OFFICIAL_INTEGRATION_ADMIN_PATTERN = /^tunag-\d{5}-[a-z0-9-]+-admin$/;

// Cache to store IDs of official creators. We use a Set to support multiple official creators.
// We do not cache non-official creator IDs since there is typically only one official user per tenant,
// making a non-official cache unnecessary, and looking up users in the Redux store is fast enough.
const officialCreatorIdsCache = new Set<string>();
const pendingFetchCreatorIds = new Set<string>();

/**
 * Check if a channel is an official tunag channel based on its creator's username.
 * Official channels are created by integration admin users with usernames matching the pattern:
 * tunag-{company_id}-{subdomain}-admin
 *
 * @param {Channel | string | null | undefined} channel - Channel object (string input not supported for creator validation)
 * @param {GlobalState} [state] - Optional Redux state. If not provided, it uses the global store state.
 * @returns {boolean} - true if channel is an official tunag channel, false otherwise
 */
export function isOfficialTunagChannel(channel: Channel | string | null | undefined, state?: GlobalState): boolean {
    // If it's a string, we cannot validate creator, so return false
    if (typeof channel === 'string' || !channel) {
        return false;
    }

    // Check if channel has creator_id
    if (!channel.creator_id) {
        return false;
    }

    // Check cache first
    if (officialCreatorIdsCache.has(channel.creator_id)) {
        return true;
    }

    // Get the creator user from Redux store
    const currentState = state || store.getState();
    const creator = getUser(currentState, channel.creator_id);

    if (!creator || !creator.username) {
        // Fetch creator user if we haven't already
        if (!pendingFetchCreatorIds.has(channel.creator_id)) {
            pendingFetchCreatorIds.add(channel.creator_id);
            setTimeout(() => {
                const {getUser: fetchUserAction} = require('mattermost-redux/actions/users');
                store.dispatch(fetchUserAction(channel.creator_id) as any);
            }, 0);
        }
        return false;
    }

    // Check if creator's username matches the integration admin pattern
    const isOfficial = OFFICIAL_INTEGRATION_ADMIN_PATTERN.test(creator.username);
    if (isOfficial) {
        officialCreatorIdsCache.add(channel.creator_id);
    }

    return isOfficial;
}

export function clearOfficialCreatorIdsCache() {
    officialCreatorIdsCache.clear();
    pendingFetchCreatorIds.clear();
}
