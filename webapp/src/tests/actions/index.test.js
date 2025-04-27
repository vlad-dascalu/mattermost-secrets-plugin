import { handleViewSecret } from '../../actions';

// Mock fetchs
global.fetch = jest.fn();

// Mock Client4.getUrl()
jest.mock('mattermost-redux/client', () => ({
    Client4: {
        getUrl: jest.fn(() => 'http://localhost:8065'),
    },
}));

describe('actions', () => {
    describe('handleViewSecret', () => {
        let dispatch;
        let getState;
        
        beforeEach(() => {
            fetch.mockClear();
            dispatch = jest.fn();
            getState = jest.fn();
        });
        
        it('should return error if no secret ID is provided', async () => {
            const post = { props: {} };
            const result = handleViewSecret(post);
            
            expect(result.error).toBeDefined();
            expect(result.error.message).toBe('Invalid secret message');
        });
        
        it('should make API calls and dispatch action on success', async () => {
            const post = { props: { secret_id: 'test-secret-id' } };
            const secretData = { message: 'This is a test secret' };
            
            // Mock successful API responses
            fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve(secretData),
                })
            );
            
            const action = handleViewSecret(post);
            const result = await action(dispatch, getState);
            
            // Check API calls
            expect(fetch).toHaveBeenCalledTimes(2);
            
            // Check dispatch was called with correct action
            expect(dispatch).toHaveBeenCalledWith({
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'test-secret-id',
                    message: secretData.message,
                },
            });
            
            // Check result
            expect(result.data).toEqual(secretData);
        });
        
        it('should handle API errors', async () => {
            const post = { props: { secret_id: 'test-secret-id' } };
            const error = new Error('API error');
            
            // Mock failed API response
            fetch.mockImplementation(() => Promise.reject(error));
            
            const action = handleViewSecret(post);
            const result = await action(dispatch, getState);
            
            // Check no dispatch called
            expect(dispatch).not.toHaveBeenCalled();
            
            // Check error returned
            expect(result.error).toBe(error);
        });
    });
}); 