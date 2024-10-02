// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';
import { ConformityDetailsView } from 'src/sections/conformityDetails';

// ----------------------------------------------------------------------

export default function ConformityPage() {
  return (
    <>
      <Helmet>
        <title> Conformity Details | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <ConformityDetailsView />
      </AuthWrapper>
    </>
  );
}
