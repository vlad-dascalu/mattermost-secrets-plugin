import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import SecretPostType from '../../components/secret_post_type';
import { Client4 } from 'mattermost-redux/client';

// Mock Client4
jest.mock('mattermost-redux/client', () => ({
    Client4: {
        getUrl: jest.fn().mockReturnValue('http://localhost:8065'),
    },
}));

// Mock localStorage
const localStorageMock = (function() {
    let store = {};
    return {
        getItem: jest.fn(key => store[key] || null),
        setItem: jest.fn((key, value) => {
            store[key] = value.toString();
        }),
        clear: jest.fn(() => {
            store = {};
        }),
    };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock fetch
global.fetch = jest.fn();

describe('components/SecretPostType', () => {
    const baseProps = {
        post: {
            props: {
                secret_id: 'test-secret-id',
                message: 'This is a test secret',
                expires_at: Date.now() + 3600000, // 1 hour from now
            },
        },
        theme: {
            centerChannelBg: '#ffffff',
            centerChannelColor: '#333333',
            linkColor: '#2389d7',
            buttonBg: '#166de0',
            buttonColor: '#ffffff',
        },
    };

    beforeEach(() => {
        localStorageMock.clear();
        global.fetch.mockClear();
    });

    it('should render correctly', () => {
        render(<SecretPostType {...baseProps} />);
        expect(screen.getByText('View Secret')).toBeInTheDocument();
    });

    it('should show loading state while fetching secret', async () => {
        // Mock fetch to never resolve
        global.fetch.mockImplementation(() => new Promise(() => {}));
        
        render(<SecretPostType {...baseProps} />);
        fireEvent.click(screen.getByText('View Secret'));
        
        expect(screen.getByText('Loading secret message...')).toBeInTheDocument();
    });

    it('should show error message when fetch fails', async () => {
        // Mock fetch to reject with an error
        global.fetch.mockImplementation(() => Promise.reject(new Error('API error')));
        
        render(<SecretPostType {...baseProps} />);
        fireEvent.click(screen.getByText('View Secret'));
        
        await waitFor(() => {
            expect(screen.getByText('API error')).toBeInTheDocument();
        });
    });

    it('should show error message when response is not OK', async () => {
        // Mock fetch to return a non-OK response
        global.fetch.mockImplementation(() => Promise.resolve({
            ok: false,
            status: 404,
            json: () => Promise.resolve({ message: 'Secret not found' })
        }));
        
        render(<SecretPostType {...baseProps} />);
        fireEvent.click(screen.getByText('View Secret'));
        
        await waitFor(() => {
            expect(screen.getByText(/Failed to fetch secret/)).toBeInTheDocument();
        });
    });

    it('should show expired message when secret has expired', () => {
        const expiredProps = {
            ...baseProps,
            post: {
                props: {
                    ...baseProps.post.props,
                    expired: true,
                },
            },
        };
        
        render(<SecretPostType {...expiredProps} />);
        expect(screen.getByText('This secret has expired and is no longer available.')).toBeInTheDocument();
    });

    it('should show already viewed message when secret was previously viewed', () => {
        const secretId = baseProps.post.props.secret_id;
        const viewedAt = Date.now() - 1000;
        localStorageMock.setItem(`secret_viewed_${secretId}`, viewedAt.toString());
        
        render(<SecretPostType {...baseProps} />);
        expect(screen.getByText('You have already viewed this secret message.')).toBeInTheDocument();
    });

    it('should show initial message before viewing secret', () => {
        render(<SecretPostType {...baseProps} />);
        expect(screen.getByText('This message contains a secret. View it once, then it disappears.')).toBeInTheDocument();
    });

    it('should show viewed message after successful fetch', async () => {
        // Mock fetch to return a successful response
        global.fetch.mockImplementation(() => Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
                ephemeralText: 'Secret content',
            }),
        }));
        
        render(<SecretPostType {...baseProps} />);
        fireEvent.click(screen.getByText('View Secret'));
        
        await waitFor(() => {
            expect(screen.getByText('You have already viewed this secret message.')).toBeInTheDocument();
        });
    });

    it('should update state when post props change', () => {
        const {rerender} = render(<SecretPostType {...baseProps} />);
        
        // Update props to mark the secret as expired
        const updatedProps = {
            ...baseProps,
            post: {
                props: {
                    ...baseProps.post.props,
                    expired: true,
                },
            },
        };
        
        rerender(<SecretPostType {...updatedProps} />);
        
        expect(screen.getByText('This secret has expired and is no longer available.')).toBeInTheDocument();
    });

    it('should handle invalid secret ID', () => {
        const invalidProps = {
            ...baseProps,
            post: {
                props: {},
            },
        };
        
        render(<SecretPostType {...invalidProps} />);
        expect(screen.getByText('Invalid secret message')).toBeInTheDocument();
    });

    it('should handle expired secret in response', async () => {
        // Mock fetch to return a response indicating the secret has expired
        global.fetch.mockImplementation(() => Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
                ephemeralText: 'This secret has expired',
            }),
        }));
        
        render(<SecretPostType {...baseProps} />);
        fireEvent.click(screen.getByText('View Secret'));
        
        await waitFor(() => {
            expect(screen.getByText('This secret has expired and is no longer available.')).toBeInTheDocument();
        });
    });
}); 