import React from 'react';
import {render} from '@testing-library/react';
import Root from '../../components/root';

describe('components/Root', () => {
    it('should render nothing (null)', () => {
        const {container} = render(<Root />);
        expect(container.firstChild).toBeNull();
    });
}); 