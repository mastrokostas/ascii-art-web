// Color palette picker — an HSL gradient canvas paired with a hex input,
// kept in sync both ways. Clicking the canvas samples a color; typing a valid
// hex moves the marker to the nearest matching pixel. The chosen value feeds
// the hidden #color input the form submits.
(function () {
    const canvas = document.getElementById('color-palette');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const hexInput = document.getElementById('hex-input');
    const colorInput = document.getElementById('color');

    // Draws the full color palette onto the canvas.
    // Three overlapping gradients: hue left-to-right, white top-to-middle, black middle-to-bottom.
    function drawPalette() {
        const hueGradient = ctx.createLinearGradient(0, 0, canvas.width, 0);
        hueGradient.addColorStop(0/6, 'hsl(0,   100%, 50%)');
        hueGradient.addColorStop(1/6, 'hsl(60,  100%, 50%)');
        hueGradient.addColorStop(2/6, 'hsl(120, 100%, 50%)');
        hueGradient.addColorStop(3/6, 'hsl(180, 100%, 50%)');
        hueGradient.addColorStop(4/6, 'hsl(240, 100%, 50%)');
        hueGradient.addColorStop(5/6, 'hsl(300, 100%, 50%)');
        hueGradient.addColorStop(6/6, 'hsl(360, 100%, 50%)');
        ctx.fillStyle = hueGradient;
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        const whiteGradient = ctx.createLinearGradient(0, 0, 0, canvas.height);
        whiteGradient.addColorStop(0,   'rgba(255,255,255,1)');
        whiteGradient.addColorStop(0.5, 'rgba(255,255,255,0)');
        ctx.fillStyle = whiteGradient;
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        const blackGradient = ctx.createLinearGradient(0, 0, 0, canvas.height);
        blackGradient.addColorStop(0.5, 'rgba(0,0,0,0)');
        blackGradient.addColorStop(1,   'rgba(0,0,0,1)');
        ctx.fillStyle = blackGradient;
        ctx.fillRect(0, 0, canvas.width, canvas.height);
    }

    // Converts r, g, b (0-255) to a #rrggbb hex string.
    function rgbToHex(r, g, b) {
        return '#' + [r, g, b].map(v => v.toString(16).padStart(2, '0')).join('');
    }

    // Redraws the palette and draws a circular selection marker at (x, y).
    // White outer ring + black inner ring so the marker is visible on any color.
    function drawMarker(x, y) {
        drawPalette();
        ctx.beginPath();
        ctx.arc(x, y, 7, 0, Math.PI * 2);
        ctx.strokeStyle = 'white';
        ctx.lineWidth = 2;
        ctx.stroke();
        ctx.beginPath();
        ctx.arc(x, y, 7, 0, Math.PI * 2);
        ctx.strokeStyle = 'black';
        ctx.lineWidth = 1;
        ctx.stroke();
    }

    // Converts a #rrggbb string to its {r, g, b} integer components.
    function hexToRgb(hex) {
        return {
            r: parseInt(hex.slice(1, 3), 16),
            g: parseInt(hex.slice(3, 5), 16),
            b: parseInt(hex.slice(5, 7), 16)
        };
    }

    // Clean palette pixels, cached once for the reverse hex lookup below.
    // Captured before any marker is drawn so the marker never pollutes it.
    let palettePixels = null;
    function cachePalettePixels() {
        palettePixels = ctx.getImageData(0, 0, canvas.width, canvas.height).data;
    }

    // Finds the palette coordinate whose color is closest to the target RGB,
    // by smallest squared distance. This lets the hex input drive the marker,
    // and degrades gracefully for colors the palette cannot reproduce exactly.
    function nearestPalettePoint(targetRed, targetGreen, targetBlue) {
        if (!palettePixels) cachePalettePixels();
        let bestDistance = Infinity;
        let bestX = 0;
        let bestY = 0;
        for (let y = 0; y < canvas.height; y++) {
            for (let x = 0; x < canvas.width; x++) {
                const index = (y * canvas.width + x) * 4;
                const deltaRed = palettePixels[index] - targetRed;
                const deltaGreen = palettePixels[index + 1] - targetGreen;
                const deltaBlue = palettePixels[index + 2] - targetBlue;
                const distance = deltaRed * deltaRed + deltaGreen * deltaGreen + deltaBlue * deltaBlue;
                if (distance < bestDistance) {
                    bestDistance = distance;
                    bestX = x;
                    bestY = y;
                }
            }
        }
        return { x: bestX, y: bestY };
    }

    // Canvas click — sample pixel, update hex input and hidden input, draw marker.
    canvas.addEventListener('click', function(e) {
        const rect = canvas.getBoundingClientRect();
        const x = Math.round(e.clientX - rect.left);
        const y = Math.round(e.clientY - rect.top);
        const pixel = ctx.getImageData(x, y, 1, 1).data;
        const hex = rgbToHex(pixel[0], pixel[1], pixel[2]);
        hexInput.value = hex;
        colorInput.value = hex;
        drawMarker(x, y);
    });

    // Hex input — validate and update hidden input. On a valid value also
    // move the palette marker to match, so the sync is bidirectional.
    // If the value is not a valid #rrggbb, hidden input stays empty (plain render fallback).
    hexInput.addEventListener('input', function() {
        const value = hexInput.value.trim();
        if (/^#[0-9a-fA-F]{6}$/.test(value)) {
            colorInput.value = value;
            const rgb = hexToRgb(value);
            const point = nearestPalettePoint(rgb.r, rgb.g, rgb.b);
            drawMarker(point.x, point.y);
        } else {
            colorInput.value = '';
        }
    });

    drawPalette();
    // Cache the clean palette now, before any marker is drawn, for the
    // reverse hex lookup. If a hex value was restored (no-JS POST fallback),
    // reflect it on the palette so the marker matches the input on load.
    cachePalettePixels();
    const initialHex = hexInput.value.trim();
    if (/^#[0-9a-fA-F]{6}$/.test(initialHex)) {
        const initialRgb = hexToRgb(initialHex);
        const initialPoint = nearestPalettePoint(initialRgb.r, initialRgb.g, initialRgb.b);
        drawMarker(initialPoint.x, initialPoint.y);
    }
})();
