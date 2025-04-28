import { handleViewSecret } from '../../actions';
import {id as pluginId} from '../../manifest';

// Mock fetchs
global.fetch = jest.fn();

// Mock localStorage
const localStorageMock = {
    getItem: jest.fn(),
    setItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

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
            localStorageMock.getItem.mockClear();
            localStorageMock.setItem.mockClear();
            dispatch = jest.fn();
            getState = jest.fn();
        });
        
        it('should return error if no secret ID is provided', async () => {
            const post = { props: {} };
            const result = handleViewSecret(post);
            
            expect(result.error).toBeDefined();
            expect(result.error.message).toBe('Invalid secret message');
        });
        
        it('should return already viewed data if secret was previously viewed', async () => {
            const post = { props: { secret_id: 'test-secret-id' } };
            const viewedAt = Date.now();
            localStorageMock.getItem.mockReturnValue(viewedAt.toString());
            
            const result = handleViewSecret(post);
            
            expect(result.data).toEqual({
                secretId: 'test-secret-id',
                alreadyViewed: true,
                viewedAt: viewedAt
            });
            expect(fetch).not.toHaveBeenCalled();
        });
        
        it('should make API call and dispatch action on success', async () => {
            const post = { props: { secret_id: 'test-secret-id' } };
            localStorageMock.getItem.mockReturnValue(null);
            
            // Mock successful API response
            fetch.mockImplementation(() => 
                Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({}),
                })
            );
            
            const action = handleViewSecret(post);
            const result = await action(dispatch, getState);
            
            // Check API call
            expect(fetch).toHaveBeenCalledTimes(1);
            expect(fetch).toHaveBeenCalledWith(
                `http://localhost:8065/plugins/${pluginId}/api/v1/secrets/view?secret_id=test-secret-id`,
                expect.any(Object)
            );
            
            // Check localStorage was updated
            expect(localStorageMock.setItem).toHaveBeenCalledWith(
                'secret_viewed_test-secret-id',
                expect.any(String)
            );
            
            // Check dispatch was called with correct action
            expect(dispatch).toHaveBeenCalledWith({
                type: 'RECEIVED_SECRET',
                data: {
                    secretId: 'test-secret-id',
                    viewed: true,
                    viewedAt: expect.any(Number),
                },
            });
            
            // Check result
            expect(result.data).toEqual({
                secretId: 'test-secret-id',
                viewed: true,
                viewedAt: expect.any(Number),
            });
        });
        
        it('should handle API errors', async () => {
            const post = { props: { secret_id: 'test-secret-id' } };
            localStorageMock.getItem.mockReturnValue(null);
            const error = new Error('API error');
            
            // Mock failed API response
            fetch.mockImplementation(() => Promise.reject(error));
            
            const action = handleViewSecret(post);
            const result = await action(dispatch, getState);
            
            // Check no dispatch called
            expect(dispatch).not.toHaveBeenCalled();
            
            // Check localStorage was not updated
            expect(localStorageMock.setItem).not.toHaveBeenCalled();
            
            // Check error returned
            expect(result.error).toBe(error);
        });

        it('should handle JSON parsing errors', async () => {
            // Mock fetch to return a response that can't be parsed as JSON
            global.fetch = jest.fn().mockImplementation(() => Promise.resolve({
                ok: true,
                json: () => {
                    throw new Error('Invalid JSON');
                },
            }));

            const post = {
                props: {
                    secret_id: 'test-secret-id',
                },
            };

            const dispatch = jest.fn();
            const result = await handleViewSecret(post)(dispatch);
            
            expect(result.error).toBeDefined();
            expect(result.error.message).toContain('Failed to fetch secret');
            expect(dispatch).not.toHaveBeenCalled();
            expect(localStorageMock.setItem).not.toHaveBeenCalled();
        });

        it('should handle non-OK responses with error message', async () => {
            // Mock fetch to return a non-OK response with an error message
            global.fetch = jest.fn().mockImplementation(() => Promise.resolve({
                ok: false,
                status: 404,
                json: () => Promise.resolve({
                    message: 'Secret not found',
                }),
            }));

            const post = {
                props: {
                    secret_id: 'test-secret-id',
                },
            };

            const result = await handleViewSecret(post)(jest.fn());
            expect(result.error).toBeDefined();
            expect(result.error.message).toContain('Failed to fetch secret: Secret not found');
        });

        it('should handle non-OK responses without error message', async () => {
            // Mock fetch to return a non-OK response without an error message
            global.fetch = jest.fn().mockImplementation(() => Promise.resolve({
                ok: false,
                status: 500,
                json: () => Promise.resolve({}),
            }));

            const post = {
                props: {
                    secret_id: 'test-secret-id',
                },
            };

            const result = await handleViewSecret(post)(jest.fn());
            expect(result.error).toBeDefined();
            expect(result.error.message).toContain('Failed to fetch secret: Status: 500');
        });
    });
}); 