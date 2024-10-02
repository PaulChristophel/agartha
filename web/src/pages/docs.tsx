import { useEffect } from 'react';
import { Helmet } from 'react-helmet-async';
import { useNavigate } from 'react-router-dom';

// ----------------------------------------------------------------------

export default function DocsPage() {
  const navigate = useNavigate();

  useEffect(() => {
    window.location.href = '/docs/index.html';
  }, [navigate]);

  return (
    <>
      <Helmet>
        <title> Documentation | Agartha </title>
      </Helmet>
      <div>Redirecting to Documentation...</div>
    </>
  );
}
