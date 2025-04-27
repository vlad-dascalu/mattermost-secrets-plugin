import React from 'react';
import PropTypes from 'prop-types';
import {Client4} from 'mattermost-redux/client';

import {id as pluginId} from '../manifest';

export default class SecretPostType extends React.PureComponent {
    static propTypes = {
        post: PropTypes.object.isRequired,
        theme: PropTypes.object.isRequired,
    };

    constructor(props) {
        super(props);

        // Check localStorage to see if this secret has been viewed
        const secretId = props.post.props && props.post.props.secret_id;
        const viewedKey = `secret_viewed_${secretId}`;
        const viewedData = localStorage.getItem(viewedKey);
        const viewed = viewedData !== null;

        this.state = {
            error: null,
            loading: false,
            viewed: viewed,
            viewedAt: viewedData ? parseInt(viewedData, 10) : null,
        };
    }

    viewSecret = async (secretId) => {
        this.setState({loading: true, error: null});

        try {
            console.log(`Attempting to view secret: ${secretId}`);
            
            // Fetch the secret content - the server will respond with an ephemeral message
            const response = await fetch(`${Client4.getUrl()}/plugins/${pluginId}/api/v1/secrets/view?secret_id=${secretId}`, {
                method: 'POST', // Use POST as required by the server
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
            });

            const responseData = await response.json();
            console.log('Response from server:', responseData);

            if (!response.ok) {
                const errorMessage = responseData.message || `Status: ${response.status}`;
                console.error(`Failed to view secret: ${errorMessage}`);
                throw new Error(`Failed to fetch secret: ${errorMessage}`);
            }

            // Mark this secret as viewed in localStorage so it persists across refreshes
            const viewedAt = Date.now();
            localStorage.setItem(`secret_viewed_${secretId}`, viewedAt.toString());

            // The server handles displaying the secret in an ephemeral message
            this.setState({
                loading: false,
                viewed: true,
                viewedAt: viewedAt,
            });
            
            console.log('Secret viewed successfully, ephemeral message should appear above this post');
        } catch (error) {
            console.error('Error viewing secret:', error);
            this.setState({
                error: error.message,
                loading: false,
            });
        }
    };

    render() {
        const {post, theme} = this.props;
        const {error, loading, viewed, viewedAt} = this.state;

        // Extract the secret ID from the post props
        const secretId = post.props && post.props.secret_id;
        
        if (!secretId) {
            return (
                <div className='SecretPostType__container'>
                    <div className='SecretPostType__error'>Invalid secret message</div>
                </div>
            );
        }

        if (error) {
            return (
                <div className='SecretPostType__container'>
                    <div className='SecretPostType__error'>{error}</div>
                </div>
            );
        }

        if (loading) {
            return (
                <div className='SecretPostType__container'>
                    <div className='SecretPostType__loading'>Loading secret message...</div>
                </div>
            );
        }

        return (
            <div 
                className='SecretPostType__container'
                style={{
                    backgroundColor: theme.centerChannelBg,
                    border: `1px solid ${theme.centerChannelColor}20`,
                    borderRadius: '4px',
                    padding: '12px',
                    marginTop: '8px',
                }}
            >
                <div className='SecretPostType__header'>
                    <i className='icon fa fa-lock' style={{color: theme.linkColor}}/>
                    <span style={{marginLeft: '8px', fontWeight: 'bold'}}>Secret Message</span>
                </div>
                <div 
                    className='SecretPostType__message'
                    style={{
                        padding: '8px',
                        marginTop: '8px',
                    }}
                >
                    {viewed ? (
                        <div>
                            <p style={{fontWeight: 'bold'}}>You have already viewed this secret message.</p>
                            <p>This secret can only be viewed once per user.</p>
                            <p style={{fontStyle: 'italic', color: '#888'}}>
                                The secret content was shown in a temporary message that should 
                                disappear when you navigate away from this channel.
                            </p>
                            {viewedAt && (
                                <p style={{fontSize: '12px', color: theme.centerChannelColor, fontStyle: 'italic'}}>
                                    Viewed on {new Date(viewedAt).toLocaleString()}
                                </p>
                            )}
                        </div>
                    ) : (
                        <>
                            <p>This message contains a secret. View it once, then it disappears.</p>
                            <p><em>The secret will be shown only to you in a temporary message that will disappear when you navigate to another channel.</em></p>
                            <button 
                                className='btn btn-primary'
                                onClick={() => this.viewSecret(secretId)}
                                style={{
                                    backgroundColor: theme.buttonBg,
                                    color: theme.buttonColor,
                                    marginTop: '8px',
                                }}
                            >
                                View Secret
                            </button>
                        </>
                    )}
                </div>
            </div>
        );
    }
} 