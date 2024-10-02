// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { JobDetailsView } from 'src/sections/jobDetails';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function JobsPage() {
  return (
    <>
      <Helmet>
        <title> Job Details | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <JobDetailsView />
      </AuthWrapper>
    </>
  );
}
