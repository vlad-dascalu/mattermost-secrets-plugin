import {id as pluginId} from './manifest';
import Root from './components/root';
import SecretPostType from './components/secret_post_type';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        try {
            // Register the root component
            registry.registerRootComponent(Root);
            
            // Register a custom post type for secret messages
            registry.registerPostTypeComponent('custom_secret', SecretPostType);
            
            // Note: registerPostAction is no longer supported in newer Mattermost versions
            // We'll handle view actions directly in the SecretPostType component
        } catch (error) {
            console.error('Error initializing secrets plugin:', error);
        }
    }
}

// Only register the plugin if window.registerPlugin is available
if (typeof window !== 'undefined' && window.registerPlugin) {
    window.registerPlugin(pluginId, new Plugin());
}

// Export the plugin id
export const id = pluginId; 