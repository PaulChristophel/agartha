// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { RunView } from 'src/sections/run';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function RunPage() {
  return (
    <>
      <Helmet>
        <title> Run | Sylvite </title>
      </Helmet>

      <AuthWrapper>
        <RunView />
      </AuthWrapper>
    </>
  );
}
