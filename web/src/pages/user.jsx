import { Helmet } from 'react-helmet-async';

import { UserView } from 'src/sections/user/view';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function UserPage() {
  return (
    <>
      <Helmet>
        <title> User | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <UserView />
      </AuthWrapper>
    </>
  );
}
