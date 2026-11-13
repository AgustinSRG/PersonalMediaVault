// Self-destroying service worker (PWA plugin migration)

self.addEventListener("install", (e) => {
    self.skipWaiting();
});

self.addEventListener("activate", (e) => {
    e.waitUntil(
        (async () => {
            await self.registration.unregister();
            const cacheNames = await self.caches.keys();
            await Promise.all(cacheNames.map((name) => self.caches.delete(name)));
            const clients = await self.clients.matchAll({ type: "window" });
            await Promise.all(clients.map((client) => client.navigate(client.url).catch(() => {})));
        })(),
    );
});
