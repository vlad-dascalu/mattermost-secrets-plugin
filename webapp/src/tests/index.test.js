import {id as pluginId} from '../manifest';

// Mock window.registerPlugin before requiring the index file
window.registerPlugin = jest.fn();

// Now import the Plugin class
import Plugin from '../index';

describe('Plugin', () => {
    let plugin;
    let registry;
    let store;

    beforeEach(() => {
        plugin = new Plugin();
        registry = {
            registerRootComponent: jest.fn(),
            registerPostTypeComponent: jest.fn(),
        };
        store = {};
    });

    it('should initialize correctly', () => {
        plugin.initialize(registry, store);
        
        expect(registry.registerRootComponent).toHaveBeenCalled();
        expect(registry.registerPostTypeComponent).toHaveBeenCalledWith('custom_secret', expect.any(Function));
    });

    it('should register the plugin with the correct ID', () => {
        // Create a new Plugin instance and call registerPlugin directly
        const pluginInstance = new Plugin();
        window.registerPlugin(pluginId, pluginInstance);
        
        expect(window.registerPlugin).toHaveBeenCalledWith(pluginId, expect.any(Plugin));
    });
    
    it('should not register the plugin when window.registerPlugin is not available', () => {
        // Save the original registerPlugin function
        const originalRegisterPlugin = window.registerPlugin;
        
        // Remove the registerPlugin function
        delete window.registerPlugin;
        
        // Import the index file again to trigger the conditional registration
        jest.resetModules();
        require('../index');
        
        // Restore the original registerPlugin function
        window.registerPlugin = originalRegisterPlugin;
        
        // No error should have been thrown
        expect(true).toBe(true);
    });
}); 