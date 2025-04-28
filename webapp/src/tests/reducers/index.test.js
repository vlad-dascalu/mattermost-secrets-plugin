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
                    viewed: true,
                    viewedAt: Date.now(),
                },
            };
            
            const expectedState = {
                secrets: {
                    'test-secret-id': true,
                },
            };
            
            expect(reducer(initialState, action)).toEqual(expectedState);
        });
        
        it('should handle multiple secrets', () => {
            const initialState = {
                secrets: {
                    'existing-secret-id': true,
                },
            };
            
            const action = {
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'new-secret-id',
                    viewed: true,
                    viewedAt: Date.now(),
                },
            };
            
            const expectedState = {
                secrets: {
                    'existing-secret-id': true,
                    'new-secret-id': true,
                },
            };
            
            expect(reducer(initialState, action)).toEqual(expectedState);
        });
        
        it('should override existing secret with same ID', () => {
            const initialState = {
                secrets: {
                    'test-secret-id': true,
                },
            };
            
            const action = {
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'test-secret-id',
                    viewed: true,
                    viewedAt: Date.now(),
                },
            };
            
            const expectedState = {
                secrets: {
                    'test-secret-id': true,
                },
            };
            
            expect(reducer(initialState, action)).toEqual(expectedState);
        });
        
        it('should ignore unknown actions', () => {
            const initialState = {
                secrets: {
                    'test-secret-id': true,
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