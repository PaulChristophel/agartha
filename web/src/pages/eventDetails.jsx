// Make sure to import PropTypes
import { Helmet } from 'react-helmet-async';

import AuthWrapper from 'src/sections/login/AuthWrapper.tsx';
import { EventDetailsView } from 'src/sections/eventDetails';

// ----------------------------------------------------------------------

export default function EventsPage() {
  return (
    <>
      <Helmet>
        <title> Event Details | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <EventDetailsView />
      </AuthWrapper>
    </>
  );
}
