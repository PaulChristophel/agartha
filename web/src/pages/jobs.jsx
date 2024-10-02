// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { JobsView } from 'src/sections/jobs';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function JobsPage() {
  return (
    <>
      <Helmet>
        <title> Jobs | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <JobsView />
      </AuthWrapper>
    </>
  );
}
