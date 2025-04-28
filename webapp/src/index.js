import {id as pluginId} from './manifest';
import Root from './components/root';
import SecretPostType from './components/secret_post_type';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        // Register the root component
        registry.registerRootComponent(Root);
        
        console.log('Registering secret post type component');
        
        // Register a custom post type for secret messages
        registry.registerPostTypeComponent('custom_secret', SecretPostType);
        
        // Note: registerPostAction is no longer supported in newer Mattermost versions
        // We'll handle view actions directly in the SecretPostType component
    }
}

console.log(`Secrets plugin (${pluginId}) initialized`);
window.registerPlugin(pluginId, new Plugin()); 