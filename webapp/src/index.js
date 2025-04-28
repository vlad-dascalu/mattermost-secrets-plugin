import {id as pluginId} from './manifest';
import Root from './components/root';
import SecretPostType from './components/secret_post_type';

// Register the plugin
export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    async initialize(registry, store) {
        try {
            // Register the root component
            registry.registerRootComponent(Root);
            
            // Register a custom post type for secret messages
            registry.registerPostTypeComponent('custom_secret', SecretPostType);
            
            // Register the message watcher
            registry.registerMessageWillBePostedHook(
                (post) => {
                    try {
                        // Process the message
                        return {post, error: null};
                    } catch (error) {
                        return {
                            post,
                            error: {
                                message: 'Failed to process message',
                                id: pluginId,
                            },
                        };
                    }
                }
            );
        } catch (error) {
            // Handle error appropriately
        }
    }
}

// Only register the plugin if window.registerPlugin is available
if (typeof window !== 'undefined' && window.registerPlugin) {
    window.registerPlugin(pluginId, new Plugin());
}

// Export the plugin
export const id = pluginId; 