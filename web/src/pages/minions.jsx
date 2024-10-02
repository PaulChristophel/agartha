import { Helmet } from 'react-helmet-async';

import { MinionsView } from 'src/sections/minions';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function MinionsPage() {
  return (
    <>
      <Helmet>
        <title> Minions | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <MinionsView />
      </AuthWrapper>
    </>
  );
}
