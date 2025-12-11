import { getChannel } from "mattermost-redux/selectors/entities/channels";
import { useSelector } from "react-redux";
import { GlobalState } from "types/store";
import { isOfficialTunagChannel } from "./official_channel_utils";
import {Permissions} from 'mattermost-redux/constants';
import { haveIChannelPermission } from "mattermost-redux/selectors/entities/roles";


export const useIsAvailableUnofficialChannel = (channelId: string): boolean => {
    return useSelector((state: GlobalState) => {
        const channel = getChannel(state, channelId);

        if (!channel) {
            return false;
        }

        const isOfficial = isOfficialTunagChannel(channel);

        if (isOfficial) {
            return true;
        }

        let permission: string;

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
    });
}
