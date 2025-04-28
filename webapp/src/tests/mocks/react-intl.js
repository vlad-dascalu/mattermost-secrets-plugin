// Mock for react-intl
export const FormattedMessage = ({defaultMessage, values}) => {
    if (values) {
        let message = defaultMessage;
        Object.keys(values).forEach(key => {
            message = message.replace(`{${key}}`, values[key]);
        });
        return message;
    }
    return defaultMessage;
};

export const IntlProvider = ({children}) => children;

export const useIntl = () => ({
    formatMessage: ({defaultMessage, values}) => {
        if (values) {
            let message = defaultMessage;
            Object.keys(values).forEach(key => {
                message = message.replace(`{${key}}`, values[key]);
            });
            return message;
        }
        return defaultMessage;
    },
    locale: 'en',
}); 