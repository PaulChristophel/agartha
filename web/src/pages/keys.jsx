import { Helmet } from 'react-helmet-async';

import { KeysView } from 'src/sections/keys';

// ----------------------------------------------------------------------

export default function KeysPage() {
  return (
    <>
      <Helmet>
        <title> Keys | Agartha </title>
      </Helmet>

      <KeysView />
    </>
  );
}
