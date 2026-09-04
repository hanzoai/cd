import {createBrowserHistory} from 'history';

// The browser history and base href are shared singletons: app.tsx feeds them
// to the router, and the router-compat adapters read them when they synthesize
// v5-shaped route props. Keeping them in a leaf module lets both import without
// a cycle back through the app component.
const bases = document.getElementsByTagName('base');
export const base = bases.length > 0 ? bases[0].getAttribute('href') || '/' : '/';
export const history = createBrowserHistory({basename: base});
