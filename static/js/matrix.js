// Matrix digital rain — full-viewport canvas backdrop behind the cards.
(function () {
    const canvas = document.getElementById('matrix');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const glyphs = 'アイウエオカキクケコサシスセソタチツテトナニヌネノﾊﾋﾌﾍﾎ0123456789'.split('');
    const fontSize = 16;
    let cols, drops;

    function reset() {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
        ctx.fillStyle = '#050506';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        cols = Math.ceil(canvas.width / fontSize);
        drops = Array.from({ length: cols }, () => Math.floor(Math.random() * -50));
    }

    function draw() {
        // translucent wash leaves a fading tail behind each glyph
        ctx.fillStyle = 'rgba(5, 5, 6, 0.08)';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.font = `${fontSize}px monospace`;

        for (let i = 0; i < cols; i++) {
            const g = glyphs[(Math.random() * glyphs.length) | 0];
            const x = i * fontSize;
            const y = drops[i] * fontSize;
            // bright leading glyph, green trail
            ctx.fillStyle = Math.random() > 0.975 ? '#caffca' : '#00ff00';
            ctx.fillText(g, x, y);
            if (y > canvas.height && Math.random() > 0.975) drops[i] = 0;
            drops[i]++;
        }
    }

    reset();
    window.addEventListener('resize', reset);
    setInterval(draw, 45);
})();
