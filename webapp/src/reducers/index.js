import {combineReducers} from 'redux';

// Simple reducer just to track that secrets were viewed
const secrets = (state = {}, action) => {
    switch (action.type) {
    case 'RECEIVED_SECRET':
        return {
            ...state,
            [action.data.secretId]: true,
        };
    default:
        return state;
    }
};

export default combineReducers({
    secrets,
}); 