import {id as pluginId} from './manifest';
import Root from './components/root';
import SecretPostType from './components/secret_post_type';
import {handleViewSecret} from './actions';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        registry.registerRootComponent(Root);
        
        // Register a custom post type for secret messages
        registry.registerPostTypeComponent('custom_secret', SecretPostType);
        
        // Register a handler for the view secret action
        registry.registerPostAction('view_secret', handleViewSecret);
    }
}

window.registerPlugin(pluginId, new Plugin()); 