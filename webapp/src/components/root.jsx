import React from 'react';

/**
 * Root component required by the Mattermost plugin framework.
 * This component doesn't render anything visible but is registered as the root component
 * to satisfy the plugin architecture requirements.
 */
export default class Root extends React.PureComponent {
    render() {
        // This component doesn't render anything visible
        return null;
    }
} 