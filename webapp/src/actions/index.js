// Handle the view secret action when a user clicks the button
export const handleViewSecret = (secretId) => async () => {
    try {
        const response = await fetch(`/plugins/secrets-plugin/api/v1/secrets/view?secret_id=${secretId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        return data;
    } catch (error) {
        // Log error to server instead of console
        return {error: error.message};
    }
};

export const handleCloseSecret = (secretId, postId) => async () => {
    try {
        const response = await fetch(`/plugins/secrets-plugin/api/v1/secrets/close?secret_id=${secretId}&post_id=${postId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        return data;
    } catch (error) {
        // Log error to server instead of console
        return {error: error.message};
    }
}; 