// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import { EventsView } from 'src/sections/events';
import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';

// ----------------------------------------------------------------------

export default function EventsPage() {
  return (
    <>
      <Helmet>
        <title> Events | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <EventsView />
      </AuthWrapper>
    </>
  );
}
