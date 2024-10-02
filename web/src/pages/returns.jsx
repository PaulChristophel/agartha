// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { ReturnsView } from 'src/sections/returns';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function ReturnsPage() {
  return (
    <>
      <Helmet>
        <title> Salt Returns | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <ReturnsView />
      </AuthWrapper>
    </>
  );
}
