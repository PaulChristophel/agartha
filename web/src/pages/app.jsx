// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { AppView } from 'src/sections/overview';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function AppPage() {
  return (
    <>
      <Helmet>
        <title> Dashboard | Agartha </title>
      </Helmet>
      <AuthWrapper>
        <AppView />
      </AuthWrapper>
    </>
  );
}
