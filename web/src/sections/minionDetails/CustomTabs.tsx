import Tabs from '@mui/material/Tabs';
import { styled } from '@mui/material/styles';

const CustomTabs = styled(Tabs)({
  '& .MuiTabs-indicator': {
    backgroundColor: '#00e676', // Adjust the color to match your design
  },
});

export default CustomTabs;
