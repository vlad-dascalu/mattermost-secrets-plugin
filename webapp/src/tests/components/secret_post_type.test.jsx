import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';

import SecretPostType from '../../components/secret_post_type';

// Mock fetchs
global.fetch = jest.fn();

// Mock Client4.getUrl()
jest.mock('mattermost-redux/client', () => ({
    Client4: {
        getUrl: jest.fn(() => 'http://localhost:8065'),
    },
}));

describe('SecretPostType', () => {
    const defaultProps = {
        post: {
            props: {
                secret_id: 'test-secret-id',
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
        fetch.mockClear();
        fetch.mockImplementation(() => 
            Promise.resolve({
                ok: true,
                json: () => Promise.resolve({ message: 'This is a secret message', allow_copy: true }),
            })
        );
    });

    it('should render initial view with button', () => {
        render(<SecretPostType {...defaultProps} />);
        
        // Check that the button is rendered
        expect(screen.getByText('View Secret')).toBeInTheDocument();
        expect(screen.getByText('This message contains a secret. View it once, then it disappears.')).toBeInTheDocument();
    });

    it('should show loading state when button is clicked', async () => {
        render(<SecretPostType {...defaultProps} />);
        
        // Click the button
        fireEvent.click(screen.getByText('View Secret'));
        
        // Should show loading state
        expect(screen.getByText('Loading secret message...')).toBeInTheDocument();
        
        // Wait for the fetch to complete
        await waitFor(() => {
            expect(fetch).toHaveBeenCalledTimes(1);
        });
    });

    it('should display secret content after loading', async () => {
        render(<SecretPostType {...defaultProps} />);
        
        // Click the button
        fireEvent.click(screen.getByText('View Secret'));
        
        // Wait for the secret content to be displayed
        await waitFor(() => {
            expect(screen.getByText('This is a secret message')).toBeInTheDocument();
        });
        
        // Check that the copy button is displayed
        expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should handle errors', async () => {
        // Mock a failed fetch
        fetch.mockImplementationOnce(() => 
            Promise.resolve({
                ok: false,
                status: 404,
            })
        );
        
        render(<SecretPostType {...defaultProps} />);
        
        // Click the button
        fireEvent.click(screen.getByText('View Secret'));
        
        // Wait for the error message to be displayed
        await waitFor(() => {
            expect(screen.getByText(/Failed to fetch secret/)).toBeInTheDocument();
        });
    });

    it('should handle missing secret ID', () => {
        const propsWithoutSecretId = {
            ...defaultProps,
            post: {
                props: {},
            },
        };
        
        render(<SecretPostType {...propsWithoutSecretId} />);
        
        // Should show an error message
        expect(screen.getByText('Invalid secret message')).toBeInTheDocument();
    });
}); 