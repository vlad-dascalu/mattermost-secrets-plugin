import {Client4} from 'mattermost-redux/client';
import {id as pluginId} from '../manifest';

// Handle the view secret action when a user clicks the button
export function handleViewSecret(post) {
    const secretId = post.props && post.props.secret_id;
    
    if (!secretId) {
        return {error: {message: 'Invalid secret message'}};
    }

    // Check if this secret has already been viewed
    const viewedKey = `secret_viewed_${secretId}`;
    const viewedData = localStorage.getItem(viewedKey);
    if (viewedData) {
        console.log(`Secret ${secretId} has already been viewed at ${new Date(parseInt(viewedData, 10)).toLocaleString()}`);
        return {
            data: {
                secretId,
                alreadyViewed: true,
                viewedAt: parseInt(viewedData, 10)
            }
        };
    }

    return async (dispatch) => {
        try {
            console.log(`Action: Attempting to view secret: ${secretId}`);
            
            // Get the secret content directly via API endpoint
            const response = await fetch(`${Client4.getUrl()}/plugins/${pluginId}/api/v1/secrets/view?secret_id=${secretId}`, {
                method: 'POST', // Using POST as required by the server
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
            });

            let responseData;
            try {
                responseData = await response.json();
                console.log('Action: Response from server:', responseData);
            } catch (e) {
                console.error('Action: Failed to parse JSON response:', e);
                throw new Error('Failed to fetch secret: Invalid JSON response');
            }

            if (!response.ok) {
                const errorMessage = responseData.message || `Status: ${response.status}`;
                console.error(`Action: Failed to view secret: ${errorMessage}`);
                throw new Error(`Failed to fetch secret: ${errorMessage}`);
            }

            // Mark this secret as viewed in localStorage
            const viewedAt = Date.now();
            localStorage.setItem(viewedKey, viewedAt.toString());
            console.log('Action: Secret marked as viewed in localStorage');

            console.log('Action: Secret viewed successfully, look for ephemeral message');

            // Dispatch to update UI components
            dispatch({
                type: 'RECEIVED_SECRET',
                data: {
                    secretId,
                    viewed: true,
                    viewedAt,
                },
            });

            return {
                data: {
                    secretId,
                    viewed: true,
                    viewedAt,
                }
            };
        } catch (error) {
            console.error('Action: Error viewing secret:', error);
            return {error};
        }
    };
} 