import { useRef, useEffect } from 'react';

import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import type { Theme, SxProps } from '@mui/material/styles';

export interface FilterTypeaheadOption<T extends string> {
  value: T;
  label: string;
  keywords?: string;
}

interface FilterTypeaheadSelectProps<T extends string> {
  label: string;
  value: T;
  options: FilterTypeaheadOption<T>[];
  onChange: (value: T) => void;
  size?: 'small' | 'medium';
  shrinkLabel?: boolean;
  sx?: SxProps<Theme>;
}

const FilterTypeaheadSelect = <T extends string>({
  label,
  value,
  options,
  onChange,
  size = 'medium',
  shrinkLabel = false,
  sx,
}: FilterTypeaheadSelectProps<T>) => {
  const typeaheadBuffer = useRef('');
  const clearBufferTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(
    () => () => {
      if (clearBufferTimer.current) clearTimeout(clearBufferTimer.current);
    },
    []
  );

  const findOption = (query: string) =>
    options.find((option) => {
      const candidates = [option.label, option.value, option.keywords ?? ''].flatMap(
        (candidate) => [candidate, ...candidate.split(/\s+/)]
      );
      return candidates.some((candidate) => candidate.toLocaleLowerCase().startsWith(query));
    });

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key.length !== 1 || event.ctrlKey || event.metaKey || event.altKey) return;

    if (clearBufferTimer.current) clearTimeout(clearBufferTimer.current);
    typeaheadBuffer.current += event.key.toLocaleLowerCase();

    let match = findOption(typeaheadBuffer.current);
    if (!match && typeaheadBuffer.current.length > 1) {
      typeaheadBuffer.current = event.key.toLocaleLowerCase();
      match = findOption(typeaheadBuffer.current);
    }

    clearBufferTimer.current = setTimeout(() => {
      typeaheadBuffer.current = '';
    }, 700);

    if (match) {
      event.preventDefault();
      event.stopPropagation();
      onChange(match.value);
    }
  };

  return (
    <TextField
      select
      size={size}
      label={label}
      value={value}
      onChange={(event) => onChange(event.target.value as T)}
      onKeyDown={handleKeyDown}
      InputLabelProps={shrinkLabel ? { shrink: true } : undefined}
      sx={sx}
    >
      {options.map((option) => (
        <MenuItem key={option.value} value={option.value}>
          {option.label}
        </MenuItem>
      ))}
    </TextField>
  );
};

export default FilterTypeaheadSelect;
