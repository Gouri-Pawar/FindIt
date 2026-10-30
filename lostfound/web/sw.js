const CACHE = 'findit-v1';
const SHELL = ['/', '/manifest.json', '/icons/icon-192.png', '/icons/icon-512.png'];

// Save the app files when the service worker installs
self.addEventListener('install', e => {
  e.waitUntil(caches.open(CACHE).then(c => c.addAll(SHELL)));
  self.skipWaiting();
});

// Delete old caches when a new version starts
self.addEventListener('activate', e => {
  e.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k)))
    )
  );
  self.clients.claim();
});

self.addEventListener('fetch', e => {
  const req = e.request;
  const url = new URL(req.url);

  // Only handle normal page and file loads from our own site
  if (req.method !== 'GET' || url.origin !== location.origin) return;

  // The page: try the internet first, use the saved copy if offline
  if (req.mode === 'navigate') {
    e.respondWith(
      fetch(req)
        .then(res => {
          const copy = res.clone();
          caches.open(CACHE).then(c => c.put('/', copy));
          return res;
        })
        .catch(() => caches.match('/'))
    );
    return;
  }

  // Icons and manifest: saved copy first, because they rarely change
  if (url.pathname.startsWith('/icons/') || url.pathname === '/manifest.json') {
    e.respondWith(caches.match(req).then(hit => hit || fetch(req)));
    return;
  }

  // Everything else (items, photos, login, claims) always comes live from the server
});