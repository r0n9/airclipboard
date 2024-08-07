var CACHE_NAME = 'snapdrop-cache-v2';
var urlsToCache = [
    'index.html',
    'css/board.css',
    'css/styles.css',
    'css/nprogress.min.css',
    'scripts/board.js',
    'scripts/network.js',
    'scripts/ui.js',
    'scripts/clipboard.js',
    'scripts/theme.js',
    'scripts/nprogress.min.js',
    'scripts/qrcode.min.js',
    'sounds/blop.mp3',
    'images/favicon-96x96.png',
    'images/donate.jpg',
    'css/board-en.json',
    'css/board-zh.json',
];

self.addEventListener('install', function (event) {
    // Perform install steps
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(function (cache) {
                console.log('Opened cache');
                return cache.addAll(urlsToCache);
            })
    );
});


self.addEventListener('fetch', function (event) {
    const url = new URL(event.request.url);
    // 如果请求 URL 包含特定路径，则跳过 Service Worker
    if (url.host.includes("googlesyndication.com")) {
        return;
    }

    event.respondWith(
        caches.match(event.request)
            .then(function (response) {
                    // Cache hit - return response
                    if (response) {
                        return response;
                    }
                    return fetch(event.request);
                }
            )
    );
});


self.addEventListener('activate', function (event) {
    console.log('Updating Service Worker...')
    event.waitUntil(
        caches.keys().then(function (cacheNames) {
            return Promise.all(
                cacheNames.filter(function (cacheName) {
                    // Return true if you want to remove this cache,
                    // but remember that caches are shared across
                    // the whole origin
                    return true
                }).map(function (cacheName) {
                    return caches.delete(cacheName);
                })
            );
        })
    );
});
