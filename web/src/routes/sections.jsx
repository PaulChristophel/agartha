import { lazy, Suspense } from 'react';
import { Outlet, Navigate, useRoutes } from 'react-router-dom';

import DashboardLayout from 'src/layouts/dashboard';

export const IndexPage = lazy(() => import('src/pages/app'));
export const MinionsPage = lazy(() => import('src/pages/minions'));
export const MinionDetailsPage = lazy(() => import('src/pages/minionDetails'));
export const JobsPage = lazy(() => import('src/pages/jobs'));
export const JobDetailsPage = lazy(() => import('src/pages/jobDetails'));
export const EventsPage = lazy(() => import('src/pages/events'));
export const EventDetailsPage = lazy(() => import('src/pages/eventDetails'));
export const LoginPage = lazy(() => import('src/pages/login'));
export const ReturnDetailsPage = lazy(() => import('src/pages/returnDetails'));
export const ConformityDetailsPage = lazy(() => import('src/pages/conformityDetails'));
export const ReturnsPage = lazy(() => import('src/pages/returns'));
export const ConformityPage = lazy(() => import('src/pages/conformity'));
export const DocsPage = lazy(() => import('src/pages/docs.tsx'));
export const RunPage = lazy(() => import('src/pages/run'));
export const KeysPage = lazy(() => import('src/pages/keys'));
export const Page404 = lazy(() => import('src/pages/page-not-found'));

// ----------------------------------------------------------------------

export default function Router() {
  const routes = useRoutes([
    {
      element: (
        <DashboardLayout>
          <Suspense>
            <Outlet />
          </Suspense>
        </DashboardLayout>
      ),
      children: [
        { element: <IndexPage />, index: true },
        { path: 'conformity', element: <ConformityPage /> },
        { path: 'conformity/:id', element: <ConformityDetailsPage /> },
        { path: 'minions', element: <MinionsPage /> },
        { path: 'minion/:id', element: <MinionDetailsPage /> },
        { path: 'jobs', element: <JobsPage /> },
        { path: 'job/:jid', element: <JobDetailsPage /> },
        { path: 'events', element: <EventsPage /> },
        { path: 'event/:id', element: <EventDetailsPage /> },
        { path: 'returns', element: <ReturnsPage /> },
        { path: 'return/:jid/:id', element: <ReturnDetailsPage /> },
        { path: 'docs', element: <DocsPage /> },
        { path: 'run', element: <RunPage /> },
        { path: 'keys', element: <KeysPage /> },
      ],
    },
    {
      path: 'login',
      element: <LoginPage />,
    },
    {
      path: '404',
      element: <Page404 />,
    },
    {
      path: '*',
      element: <Navigate to="/404" replace />,
    },
  ]);

  return routes;
}
