// Cursor-following gradient border.
// "One event listener powers it all" — a single mousemove handler on the
// container writes each .glass card's cursor-relative position into the
// --x / --y custom properties the CSS spotlight reads.
(function () {
    const scope = document.querySelector('.container') || document.body;

    scope.addEventListener('mousemove', (e) => {
        // Re-query every move so cards added after load (the result card is
        // created dynamically while typing) are included. One listener still
        // powers it all; the work is trivial for a handful of cards.
        const cards = document.querySelectorAll('.glass');
        for (const card of cards) {
            const r = card.getBoundingClientRect();
            card.style.setProperty('--x', `${e.clientX - r.left}px`);
            card.style.setProperty('--y', `${e.clientY - r.top}px`);
        }
    });

    // Static-screenshot aid: load any page with ?pose to freeze the glow on the
    // first card so a headless capture can show the effect. No-op otherwise.
    if (location.search.includes('pose')) {
        const card = document.querySelector('.glass');
        if (!card) return;
        // The .glow-posed rule lives in style.css. Building a <style> element
        // here instead would be blocked by the Content-Security-Policy, which no
        // longer allows inline styles. Setting custom properties through the
        // CSSOM below is unaffected by that policy.
        card.classList.add('glow-posed');
        card.style.setProperty('--x', '92%');
        card.style.setProperty('--y', '24%');
    }
})();