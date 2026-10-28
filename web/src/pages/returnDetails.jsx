// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';
import { ReturnDetailsView } from 'src/sections/returnDetails';

// ----------------------------------------------------------------------

export default function ReturnDetailsPage() {
  return (
    <>
      <Helmet>
        <title> Return Details | Sylvite </title>
      </Helmet>

      <AuthWrapper>
        <ReturnDetailsView />
      </AuthWrapper>
    </>
  );
}
