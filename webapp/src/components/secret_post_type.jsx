import React from 'react';
import PropTypes from 'prop-types';
import {Tooltip, OverlayTrigger} from 'react-bootstrap';
import {Client4} from 'mattermost-redux/client';

import {id as pluginId} from '../manifest';

export default class SecretPostType extends React.PureComponent {
    static propTypes = {
        post: PropTypes.object.isRequired,
        theme: PropTypes.object.isRequired,
    };

    constructor(props) {
        super(props);

        this.state = {
            secret: null,
            viewed: false,
            error: null,
            loading: false,
            copied: false,
        };
    }

    viewSecret = async (secretId) => {
        this.setState({loading: true, error: null});

        try {
            // Fetch the secret content
            const response = await fetch(`${Client4.getUrl()}/plugins/${pluginId}/api/v1/secrets/view?secret_id=${secretId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
            });

            if (!response.ok) {
                throw new Error(`Failed to fetch secret: ${response.status}`);
            }

            const data = await response.json();
            this.setState({
                secret: data,
                viewed: true,
                loading: false,
            });

            // Mark the secret as viewed
            fetch(`${Client4.getUrl()}/plugins/${pluginId}/api/v1/secrets/viewed`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
                body: JSON.stringify({secret_id: secretId}),
            });
        } catch (error) {
            this.setState({
                error: error.message,
                loading: false,
            });
        }
    };

    copyToClipboard = () => {
        if (!this.state.secret || !this.state.secret.message) {
            return;
        }

        navigator.clipboard.writeText(this.state.secret.message)
            .then(() => {
                this.setState({copied: true});
                setTimeout(() => this.setState({copied: false}), 2000);
            })
            .catch((err) => {
                this.setState({error: `Failed to copy: ${err.message}`});
            });
    };

    render() {
        const {post, theme} = this.props;
        const {secret, viewed, error, loading, copied} = this.state;

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

        if (!viewed) {
            return (
                <div className='SecretPostType__container'>
                    <div className='SecretPostType__message'>
                        <span>This message contains a secret. View it once, then it disappears.</span>
                        <button 
                            className='btn btn-primary'
                            onClick={() => this.viewSecret(secretId)}
                        >
                            View Secret
                        </button>
                    </div>
                </div>
            );
        }

        // Secret has been viewed
        if (!secret || !secret.message) {
            return (
                <div className='SecretPostType__container'>
                    <div className='SecretPostType__viewed'>
                        Secret has been viewed and is no longer available.
                    </div>
                </div>
            );
        }

        const copyTooltip = (
            <Tooltip id='copy-tooltip'>
                {copied ? 'Copied!' : 'Copy to clipboard'}
            </Tooltip>
        );

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
                    <div className='SecretPostType__controls'>
                        <OverlayTrigger 
                            placement='top'
                            overlay={copyTooltip}
                        >
                            <button
                                className='btn btn-sm'
                                onClick={this.copyToClipboard}
                                style={{
                                    marginLeft: '8px',
                                    backgroundColor: theme.buttonBg,
                                    color: theme.buttonColor,
                                }}
                            >
                                <i className='icon fa fa-copy'/>
                            </button>
                        </OverlayTrigger>
                    </div>
                </div>
                <div 
                    className='SecretPostType__content'
                    style={{
                        backgroundColor: `${theme.centerChannelColor}10`,
                        padding: '8px',
                        borderRadius: '4px',
                        marginTop: '8px',
                        whiteSpace: 'pre-wrap',
                    }}
                >
                    <pre style={{
                        backgroundColor: `${theme.centerChannelColor}20`,
                        padding: '12px',
                        borderRadius: '4px',
                        margin: '0',
                        overflowX: 'auto',
                        fontFamily: 'monospace'
                    }}>
                        <code>{secret.message}</code>
                    </pre>
                </div>
                <div 
                    className='SecretPostType__footer'
                    style={{
                        fontSize: '12px',
                        color: theme.centerChannelColor,
                        marginTop: '8px',
                        fontStyle: 'italic',
                    }}
                >
                    This message will disappear when you navigate away. Make sure to copy any important information.
                </div>
            </div>
        );
    }
} 