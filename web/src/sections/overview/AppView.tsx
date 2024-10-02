import React from 'react';

import Grid from '@mui/material/Grid';
import Container from '@mui/material/Container';

import KeysWidget from './KeysWidget.tsx';
import StatusWidget from './StatusWidget.tsx';
import ConformityWidget from './ConformityWidget.tsx';

const AppView: React.FC = () => (
  <Container maxWidth="xl">
    <Grid container spacing={3}>
      <Grid item xs={12} sm={6} md={3}>
        <ConformityWidget />
      </Grid>

      <Grid item xs={12} sm={6} md={3}>
        <KeysWidget />
      </Grid>

      <Grid item xs={12} sm={6} md={3.4}>
        <StatusWidget />
      </Grid>
    </Grid>
  </Container>
);

export default AppView;
