import { Helmet } from 'react-helmet-async';

import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';
import { MinionDetailsView } from 'src/sections/minionDetails';

// ----------------------------------------------------------------------

export default function MinionDetailsPage() {
  return (
    <>
      <Helmet>
        <title> Minion Details | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <MinionDetailsView />
      </AuthWrapper>
    </>
  );
}
