import { getChannel } from "mattermost-redux/selectors/entities/channels";
import { isOfficialTunagChannel } from "./official_channel_utils";
import {Permissions} from 'mattermost-redux/constants';
import { haveIChannelPermission } from "mattermost-redux/selectors/entities/roles";
import store from 'stores/redux_store';

export const isAvailableUnofficialChannel = (channelId: string): boolean => {
    const state = store.getState();
    const channel = getChannel(state, channelId);
    
    if (!channel) {
        return false;
    }

    if (isOfficialTunagChannel(channel)) {
        return true;
    }

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

    return haveIChannelPermission(state, channel.team_id, channel.id, permission); 
}
