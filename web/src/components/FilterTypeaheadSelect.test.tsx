import { useState } from 'react';
import { it, expect, describe } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen } from '@testing-library/react';

import FilterTypeaheadSelect, { type FilterTypeaheadOption } from './FilterTypeaheadSelect.tsx';

type TestValue = 'string' | 'bool' | 'gte';

const options: FilterTypeaheadOption<TestValue>[] = [
  { value: 'string', label: 'string' },
  { value: 'bool', label: 'bool', keywords: 'boolean' },
  { value: 'gte', label: 'greater or equal' },
];

const TestSelect = () => {
  const [value, setValue] = useState<TestValue>('string');

  return <FilterTypeaheadSelect label="Type" value={value} options={options} onChange={setValue} />;
};

describe('FilterTypeaheadSelect', () => {
  it('remains a dropdown', async () => {
    const user = userEvent.setup();
    render(<TestSelect />);

    await user.click(screen.getByRole('combobox', { name: 'Type' }));

    expect(screen.getByRole('option', { name: 'bool' })).toBeVisible();
  });

  it('selects an option by typing its label', async () => {
    const user = userEvent.setup();
    render(<TestSelect />);

    await user.tab();
    await user.keyboard('bool');

    expect(screen.getByRole('combobox', { name: 'Type' })).toHaveTextContent('bool');
  });

  it('also matches compact comparison values', async () => {
    const user = userEvent.setup();
    render(<TestSelect />);

    await user.tab();
    await user.keyboard('gte');

    expect(screen.getByRole('combobox', { name: 'Type' })).toHaveTextContent('greater or equal');
  });
});
