import { handleViewSecret, handleCloseSecret } from '../../actions';

// Mock fetch
global.fetch = jest.fn();

describe('actions', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('handleViewSecret', () => {
        beforeEach(() => {
            global.fetch.mockClear();
        });
        
        it('should call API with correct parameters', async () => {
            // Mock successful API response
            global.fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ success: true }),
                })
            );
            
            const secretId = 'test-secret-id';
            await handleViewSecret(secretId)();
            
            // Check API call
            expect(global.fetch).toHaveBeenCalledWith(
                `/plugins/secrets-plugin/api/v1/secrets/view?secret_id=${secretId}`,
                expect.objectContaining({
                    method: 'GET',
                    headers: expect.objectContaining({
                        'Content-Type': 'application/json',
                    }),
                })
            );
        });
        
        it('should return data from API on success', async () => {
            const responseData = { success: true, message: 'Secret viewed' };
            
            // Mock successful API response
            global.fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve(responseData),
                })
            );
            
            const secretId = 'test-secret-id';
            const result = await handleViewSecret(secretId)();
            
            expect(result).toEqual(responseData);
        });
        
        it('should return error object when API call fails', async () => {
            // Mock API error
            const error = new Error('API error');
            global.fetch.mockImplementation(() => Promise.reject(error));
            
            const secretId = 'test-secret-id';
            const result = await handleViewSecret(secretId)();
            
            expect(result).toEqual({ error: 'API error' });
        });

        it('should return error object when response is not OK', async () => {
            // Mock non-OK response
            global.fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: false,
                    status: 404,
                    json: () => Promise.resolve({ message: 'Not found' }),
                })
            );
            
            const secretId = 'test-secret-id';
            const result = await handleViewSecret(secretId)();
            
            expect(result).toEqual({ error: 'HTTP error! status: 404' });
        });

        it('should handle JSON parsing errors', async () => {
            // Mock fetch to return a response that can't be parsed as JSON
            global.fetch.mockImplementation(() => Promise.resolve({
                ok: true,
                json: () => Promise.reject(new Error('Invalid JSON')),
            }));

            const secretId = 'test-secret-id';
            const result = await handleViewSecret(secretId)();
            
            expect(result).toEqual({ error: 'Invalid JSON' });
        });
    });

    describe('handleCloseSecret', () => {
        it('should call API with correct parameters', async () => {
            // Mock successful API response
            global.fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ success: true }),
                })
            );
            
            const secretId = 'test-secret-id';
            const postId = 'test-post-id';
            await handleCloseSecret(secretId, postId)();
            
            // Check API call
            expect(global.fetch).toHaveBeenCalledWith(
                `/plugins/secrets-plugin/api/v1/secrets/close?secret_id=${secretId}&post_id=${postId}`,
                expect.objectContaining({
                    method: 'POST',
                    headers: expect.objectContaining({
                        'Content-Type': 'application/json',
                    }),
                })
            );
        });
        
        it('should return data from API on success', async () => {
            const responseData = { success: true, message: 'Secret closed' };
            
            // Mock successful API response
            global.fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve(responseData),
                })
            );
            
            const secretId = 'test-secret-id';
            const postId = 'test-post-id';
            const result = await handleCloseSecret(secretId, postId)();
            
            expect(result).toEqual(responseData);
        });
        
        it('should return error object when API call fails', async () => {
            // Mock API error
            const error = new Error('API error');
            global.fetch.mockImplementation(() => Promise.reject(error));
            
            const secretId = 'test-secret-id';
            const postId = 'test-post-id';
            const result = await handleCloseSecret(secretId, postId)();
            
            expect(result).toEqual({ error: 'API error' });
        });

        it('should return error object when response is not OK', async () => {
            // Mock non-OK response
            global.fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: false,
                    status: 404,
                    json: () => Promise.resolve({ message: 'Not found' }),
                })
            );
            
            const secretId = 'test-secret-id';
            const postId = 'test-post-id';
            const result = await handleCloseSecret(secretId, postId)();
            
            expect(result).toEqual({ error: 'HTTP error! status: 404' });
        });
    });
}); 