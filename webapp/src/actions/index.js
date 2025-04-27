import {Client4} from 'mattermost-redux/client';
import {id as pluginId} from '../manifest';

// Handle the view secret action when a user clicks the button
export function handleViewSecret(post, context) {
    const secretId = post.props && post.props.secret_id;
    
    if (!secretId) {
        return {error: {message: 'Invalid secret message'}};
    }

    return async (dispatch, getState) => {
        try {
            // Get the secret content
            const response = await fetch(`${Client4.getUrl()}/plugins/${pluginId}/api/v1/secrets/view?secret_id=${secretId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
            });

            if (!response.ok) {
                throw new Error(`Failed to fetch secret: ${response.status}`);
            }

            const secret = await response.json();

            // Mark the secret as viewed
            await fetch(`${Client4.getUrl()}/plugins/${pluginId}/api/v1/secrets/viewed`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
                body: JSON.stringify({secret_id: secretId}),
            });

            // Dispatch an action to update the UI
            dispatch({
                type: 'RECEIVED_SECRET',
                data: {
                    secretId,
                    message: secret.message,
                },
            });

            return {data: secret};
        } catch (error) {
            return {error};
        }
    };
} 