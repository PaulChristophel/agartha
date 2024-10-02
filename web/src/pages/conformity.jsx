// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { ConformityView } from 'src/sections/conformity';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function ConformityPage() {
  return (
    <>
      <Helmet>
        <title> Salt Conformity | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <ConformityView />
      </AuthWrapper>
    </>
  );
}
