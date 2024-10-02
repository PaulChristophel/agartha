import React from 'react';

import Tab from '@mui/material/Tab';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import { styled } from '@mui/material/styles';

const CustomTabs = styled(Tabs)({
  '& .MuiTabs-indicator': {
    backgroundColor: '#00e676', // Adjust the color to match your design
  },
});

const TabsNavigation: React.FC<{
  value: number;
  handleChange: (event: React.SyntheticEvent, newValue: number) => void;
}> = ({ value, handleChange }) => (
  <Box sx={{ width: '100%', borderBottom: 1, borderColor: 'divider' }}>
    <CustomTabs value={value} onChange={handleChange} aria-label="minion details tabs">
      <Tab label="Grains" />
      <Tab label="Pillar" />
      <Tab label="History" />
      <Tab label="Graph" />
    </CustomTabs>
  </Box>
);

export default TabsNavigation;
