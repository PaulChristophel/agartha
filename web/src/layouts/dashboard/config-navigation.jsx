import SvgColor from 'src/components/svg-color';

// ----------------------------------------------------------------------

const icon = (name) => (
  <SvgColor src={`/assets/icons/navbar/${name}.svg`} sx={{ width: 1, height: 1 }} />
);

const navConfig = [
  {
    title: 'dashboard',
    path: '/',
    icon: icon('ic_analytics'),
  },
  {
    title: 'conformity',
    path: '/conformity',
    icon: icon('ic_lock'),
  },
  {
    title: 'events',
    path: '/events',
    icon: icon('ic_events'),
  },
  {
    title: 'jobs',
    path: '/jobs',
    icon: icon('ic_jobs'),
  },
  {
    title: 'keys',
    path: '/keys',
    icon: icon('ic_key'),
  },
  {
    title: 'minions',
    path: '/minions',
    icon: icon('ic_minion'),
  },
  {
    title: 'returns',
    path: '/returns',
    icon: icon('ic_return'),
  },
  {
    title: 'run',
    path: '/run',
    icon: icon('ic_run'),
  },
  {
    type: 'divider', // Custom type to represent the horizontal line
  },
  {
    title: 'API docs',
    external: true,
    path: '/docs',
    icon: icon('ic_docs'),
  },
];

export default navConfig;
