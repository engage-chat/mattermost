// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';
import type {Dispatch} from 'redux';

import type {ChannelCategory} from '@mattermost/types/channel_categories';

import {setCategoryCollapsed, setCategorySorting} from 'mattermost-redux/actions/channel_categories';
import {savePreferences} from 'mattermost-redux/actions/preferences';
import {getCurrentUser, getCurrentUserId} from 'mattermost-redux/selectors/entities/users';
import {isAdmin} from 'mattermost-redux/utils/user_utils';

import {getDraggingState, makeGetFilteredChannelIdsForCategory} from 'selectors/views/channel_sidebar';

import type {GlobalState} from 'types/store';

import { getUser, getUsers } from 'mattermost-redux/selectors/entities/users';
import SidebarCategory from './sidebar_category';
import { CategoryTypes } from 'mattermost-redux/constants/channel_categories';
import { getAllChannels } from 'mattermost-redux/selectors/entities/channels';
import { createSelector } from 'mattermost-redux/selectors/create_selector';

type OwnProps = {
    category: ChannelCategory;
}

function makeMapStateToProps() {
    const getChannelIdsForCategory = makeGetFilteredChannelIdsForCategory();

    const selectAllCreatorsLoaded = createSelector(
        'selectAllCreatorsLoaded',
        (state: GlobalState, ownProps: OwnProps) => getChannelIdsForCategory(state, ownProps.category),
        getAllChannels,
        getUsers,
        (channelIds, allChannels, users) => {
            for (const channelId of channelIds) {
                const channel = allChannels[channelId];
                if (!channel || channel.type === 'D' || channel.type === 'G') {
                    continue;
                }
                if (channel.creator_id) {
                    const creator = users[channel.creator_id];
                    if (!creator || !creator.username) {
                        return false;
                    }
                }
            }
            return true;
        }
    );

    return (state: GlobalState, ownProps: OwnProps) => {
        const allCreatorsLoaded = ownProps.category.type === CategoryTypes.DIRECT_MESSAGES ? true : selectAllCreatorsLoaded(state, ownProps);

        return {
            channelIds: getChannelIdsForCategory(state, ownProps.category),
            draggingState: getDraggingState(state),
            currentUserId: getCurrentUserId(state),
            isAdmin: isAdmin(getCurrentUser(state).roles),
            allCreatorsLoaded,
        };
    };
}

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators({
            setCategoryCollapsed,
            setCategorySorting,
            savePreferences,
        }, dispatch),
    };
}

export default connect(makeMapStateToProps, mapDispatchToProps)(SidebarCategory);
