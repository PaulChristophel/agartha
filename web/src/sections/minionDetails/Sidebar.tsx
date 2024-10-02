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
// const Sidebar: React.FC = () => (
//   <Box sx={{ padding: 2 }}>
//     <Typography variant="h6">Sidebar</Typography>
//     {/* Add sidebar content here */}
//   </Box>
// );

const Sidebar: React.FC<{
  value: number;
  handleChange: (event: React.SyntheticEvent, newValue: number) => void;
}> = ({ value, handleChange }) => (
  <Box sx={{ width: '100%', borderBottom: 1, borderColor: 'divider' }}>
    <Box>
      <CustomTabs value={value} onChange={handleChange} aria-label="minion details tabs">
        <Tab label="Common" />
        <Tab label="Salt" />
        <Tab label="Hardware" />
      </CustomTabs>
    </Box>
    <Box>
      <CustomTabs value={value} onChange={handleChange} aria-label="minion details tabs">
        <Tab label="Interface" />
        <Tab label="MAC" />
        <Tab label="DNS" />
      </CustomTabs>
    </Box>
  </Box>
);

export default Sidebar;
