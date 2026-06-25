// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useCallback} from 'react';
import {useIntl} from 'react-intl';
import {useSelector} from 'react-redux';

import type {Channel} from '@mattermost/types/channels';

import {getChannel} from 'mattermost-redux/selectors/entities/channels';
import {getUser} from 'mattermost-redux/selectors/entities/users';

import {trackEvent} from 'actions/telemetry_actions';

import LeaveChannelModal from 'components/leave_channel_modal';
import SidebarChannelLink from 'components/sidebar/sidebar_channel/sidebar_channel_link';
import BuildingIcon from 'components/widgets/icons/building_icon';

import Constants, {ModalIdentifiers} from 'utils/constants';
import {isOfficialTunagChannel} from 'utils/official_channel_utils';

import type {GlobalState} from 'types/store';

import SidebarBaseChannelIcon from './sidebar_base_channel_icon';

import type {PropsFromRedux} from './index';
import { Visibility } from '@tanstack/react-table';

export interface Props extends PropsFromRedux {
    channel: Channel;
    currentTeamName: string;
}

const SidebarBaseChannel = ({
    channel,
    currentTeamName,
    actions,
}: Props) => {
    const intl = useIntl();

    const handleLeavePublicChannel = useCallback((callback: () => void) => {
        actions.leaveChannel(channel.id);
        trackEvent('ui', 'ui_public_channel_x_button_clicked');
        callback();
    }, [channel.id, actions.leaveChannel]);

    const handleLeavePrivateChannel = useCallback((callback: () => void) => {
        actions.openModal({modalId: ModalIdentifiers.LEAVE_PRIVATE_CHANNEL_MODAL, dialogType: LeaveChannelModal, dialogProps: {channel}});
        trackEvent('ui', 'ui_private_channel_x_button_clicked');
        callback();
    }, [channel, actions.openModal]);

    let channelLeaveHandler = null;
    if (channel.type === Constants.OPEN_CHANNEL && channel.name !== Constants.DEFAULT_CHANNEL) {
        channelLeaveHandler = handleLeavePublicChannel;
    } else if (channel.type === Constants.PRIVATE_CHANNEL) {
        channelLeaveHandler = handleLeavePrivateChannel;
    }

    const channelIconState = useSelector((state: GlobalState) => {
        const ch = getChannel(state, channel.id) ?? channel;
        if (!ch || !ch.creator_id || ch.type === Constants.DM_CHANNEL || ch.type === Constants.GM_CHANNEL) {
            return 'unofficial';
        }

        const creator = getUser(state, ch.creator_id);
        if (!creator || !creator.username) {
            return 'pending';
        }

        return isOfficialTunagChannel(ch, state) ? 'official' : 'unofficial';
    });

    let channelIcon = null;
    if (channelIconState === 'pending') {
        const iconClass = channel.type === Constants.OPEN_CHANNEL ? 'icon-globe' : 'icon-lock-outline';
        channelIcon = <i className={`icon ${iconClass}`} style={{visibility: 'hidden'}}/>;
    } else if (channelIconState === 'official') {
        channelIcon = <BuildingIcon/>;
    } else if (channelIconState === 'unofficial') {
        channelIcon = (
            <SidebarBaseChannelIcon
                channelType={channel.type}
            />
        );
    }

    let ariaLabelPrefix;
    if (channel.type === Constants.OPEN_CHANNEL) {
        ariaLabelPrefix = intl.formatMessage({id: 'accessibility.sidebar.types.public', defaultMessage: 'public channel'});
    } else if (channel.type === Constants.PRIVATE_CHANNEL) {
        ariaLabelPrefix = intl.formatMessage({id: 'accessibility.sidebar.types.private', defaultMessage: 'private channel'});
    }

    return (
        <SidebarChannelLink
            channel={channel}
            link={`/${currentTeamName}/channels/${channel.name}`}
            label={channel.display_name}
            ariaLabelPrefix={ariaLabelPrefix}
            channelLeaveHandler={channelLeaveHandler!}
            icon={channelIcon}
            isSharedChannel={channel.shared}
        />
    );
};

export default SidebarBaseChannel;
