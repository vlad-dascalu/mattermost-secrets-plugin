import {combineReducers} from 'redux';

// Reducer for secrets viewed by the current user
const secrets = (state = {}, action) => {
    switch (action.type) {
    case 'RECEIVED_SECRET':
        return {
            ...state,
            [action.data.secretId]: action.data.message,
        };
    default:
        return state;
    }
};

export default combineReducers({
    secrets,
}); 