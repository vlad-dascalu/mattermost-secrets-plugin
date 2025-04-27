import reducer from '../../reducers';

describe('reducers', () => {
    it('should have the correct initial state', () => {
        const initialState = reducer(undefined, {});
        expect(initialState).toEqual({
            secrets: {},
        });
    });
    
    describe('secrets reducer', () => {
        it('should handle RECEIVED_SECRET', () => {
            const initialState = {
                secrets: {},
            };
            
            const action = {
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'test-secret-id',
                    message: 'This is a test secret',
                },
            };
            
            const expectedState = {
                secrets: {
                    'test-secret-id': 'This is a test secret',
                },
            };
            
            expect(reducer(initialState, action)).toEqual(expectedState);
        });
        
        it('should handle multiple secrets', () => {
            const initialState = {
                secrets: {
                    'existing-secret-id': 'Existing secret message',
                },
            };
            
            const action = {
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'new-secret-id',
                    message: 'New secret message',
                },
            };
            
            const expectedState = {
                secrets: {
                    'existing-secret-id': 'Existing secret message',
                    'new-secret-id': 'New secret message',
                },
            };
            
            expect(reducer(initialState, action)).toEqual(expectedState);
        });
        
        it('should override existing secret with same ID', () => {
            const initialState = {
                secrets: {
                    'test-secret-id': 'Old secret message',
                },
            };
            
            const action = {
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'test-secret-id',
                    message: 'Updated secret message',
                },
            };
            
            const expectedState = {
                secrets: {
                    'test-secret-id': 'Updated secret message',
                },
            };
            
            expect(reducer(initialState, action)).toEqual(expectedState);
        });
        
        it('should ignore unknown actions', () => {
            const initialState = {
                secrets: {
                    'test-secret-id': 'Secret message',
                },
            };
            
            const action = {
                type: 'UNKNOWN_ACTION',
                data: {
                    something: 'else',
                },
            };
            
            expect(reducer(initialState, action)).toEqual(initialState);
        });
    });
}); 