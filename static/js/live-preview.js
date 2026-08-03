// Live preview — renders ASCII as the user types and clears it as they
// delete, by POSTing to /ascii-art and injecting the returned fragment
// without a page reload. Also owns the Live/Manual render-mode switch and
// the auto-growing textarea. Degrades gracefully: with JS off the form posts
// normally and the server returns a full page.
(function () {
    const textInput = document.getElementById('text');
    const previewForm = document.querySelector('.form');
    if (!textInput || !previewForm) return;

    const containerElement = document.querySelector('.container');

    // Color-picker fields live in color-picker.js; re-queried here by id so
    // this file stands on its own. Same DOM elements, separate references.
    const canvas = document.getElementById('color-palette');
    const hexInput = document.getElementById('hex-input');
    const colorInput = document.getElementById('color');

    // The result card is server-rendered on a no-JS POST; on a fresh GET
    // it does not exist yet and is created on demand below.
    let resultCard = document.querySelector('.result-card');
    let debounceTimer = null;
    let inFlightRequest = null;

    // Builds (or reuses) the result card and returns its <pre> element.
    function ensureResultElement() {
        if (!resultCard) {
            resultCard = document.createElement('section');
            resultCard.className = 'glass result-card';
            resultCard.innerHTML =
                '<header class="result-header">' +
                '<span class="dot"></span>' +
                '<span class="result-label">Output</span>' +
                '</header>' +
                '<pre class="result"></pre>' +
                '<form action="/download" method="POST" class="download-form">' +
                // .hidden-field rather than style="display:none": this markup is
                // parsed by the browser, so an inline style here is blocked by
                // the Content-Security-Policy. Mirrors templates/index.html.
                '<textarea name="asciiText" class="hidden-field"></textarea>' +
                '<input type="hidden" name="color" value="">' +
                '<select name="format" class="format-dropdown" aria-label="Download file format">' +
                '<option value="txt">.txt</option>' +
                '<option value="html">.html</option>' +
                '</select>' +
                '<button type="submit" class="download-btn">' +
                '<svg width="15" height="15" viewBox="0 0 15 15" fill="none" aria-hidden="true"><path d="M7.5 2 L7.5 9.5 M4.5 6.5 L7.5 9.5 L10.5 6.5 M3 12.5 L12 12.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>' +
                '<span class="download-label">Download .txt</span>' +
                '</button>' +
                '</form>';
            containerElement.appendChild(resultCard);
            // format-dropdown.js enhances the format <select> and owns the button
            // label ("Download .txt" / ".html"); it watches for this card.
        }
        return resultCard.querySelector('.result');
    }

    // Keeps the download form's hidden textarea in sync with the plain-text
    // (non-colorized) version of whatever is currently rendered.
    function updateDownloadPayload(rawText) {
        if (!resultCard) return;
        const downloadTextarea = resultCard.querySelector('.download-form textarea');
        if (downloadTextarea) {
            downloadTextarea.value = rawText;
        }
    }

    // Removes the result card entirely — used when the input is empty.
    function clearResult() {
        if (resultCard) {
            resultCard.remove();
            resultCard = null;
        }
    }

    async function renderPreview() {
        const text = textInput.value;

        // Empty input → the art disappears, matching "disappear as I delete".
        if (text === '') {
            clearResult();
            return;
        }

        const banner = document.querySelector('input[name="banner"]:checked').value;
        const color = colorInput.value;

        // Cancel any request still in flight so responses cannot arrive out
        // of order and paint stale output over newer keystrokes.
        if (inFlightRequest) {
            inFlightRequest.abort();
        }
        inFlightRequest = new AbortController();

        const requestBody = new URLSearchParams({ text: text, banner: banner, color: color });

        try {
            const response = await fetch('/ascii-art', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    'X-Requested-With': 'fetch'
                },
                body: requestBody.toString(),
                signal: inFlightRequest.signal
            });

            // On any error, leave the last good render in place.
            if (!response.ok) {
                return;
            }

            const rendered = await response.text();
            if (rendered === '') {
                clearResult();
                return;
            }

            const resultElement = ensureResultElement();
            // Hand the picked color to the card so the CSS plate can derive a
            // contrasting backing from it; empty clears it for plain output.
            resultCard.style.setProperty('--art-color', color || 'transparent');
            // Colored output is intentional <span> HTML and must be parsed.
            // Plain output is literal text — some glyphs contain real < and >
            // that must NOT be parsed as tags, so it goes in as textContent.
            if (color) {
                resultElement.innerHTML = rendered;
            } else {
                resultElement.textContent = rendered;
            }
            // textContent strips any HTML tags, giving the plain-text
            // version needed for the downloaded .txt file either way.
            updateDownloadPayload(resultElement.textContent);
            // Keep the HTML export's color in sync with the current pick.
            const colorField = resultCard.querySelector('.download-form input[name="color"]');
            if (colorField) {
                colorField.value = color;
            }
        } catch (error) {
            if (error.name !== 'AbortError') {
                console.error(error);
            }
        }
    }

    // Debounce typing so we render shortly after a pause, not per keystroke.
    function schedulePreview() {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(renderPreview, 120);
    }

    // ---------------------------------------------------------------
    // Render mode: the switch beside the banner pills decides whether
    // typing renders live, or whether rendering waits for the button.
    //   checked  → live preview, Generate button disabled (redundant)
    //   unchecked → manual, Generate button is the only trigger
    // ---------------------------------------------------------------
    const liveToggle = document.getElementById('live-toggle');
    const generateButton = document.querySelector('.generate-btn');
    const modeText = document.querySelector('.mode-text');

    // Applies the current switch state: toggles the button and label, and
    // in live mode immediately syncs the output to whatever is typed.
    function applyMode() {
        const isLive = liveToggle.checked;
        generateButton.disabled = isLive;
        modeText.textContent = isLive ? 'Live' : 'Manual';
        if (isLive) {
            renderPreview();
        }
    }

    // Auto-render wrappers — these only fire while live mode is on.
    function liveSchedule() {
        if (liveToggle.checked) {
            schedulePreview();
        }
    }
    function liveRender() {
        if (liveToggle.checked) {
            renderPreview();
        }
    }

    // Auto-grow the textarea to fit its content instead of scrolling.
    // border-box means scrollHeight excludes the border, so add it back to
    // avoid clipping the last line. min-height in CSS keeps the floor.
    const textareaBorder = (function () {
        const computed = getComputedStyle(textInput);
        return parseFloat(computed.borderTopWidth) + parseFloat(computed.borderBottomWidth);
    })();
    function autoGrowTextarea() {
        textInput.style.height = 'auto';
        textInput.style.height = (textInput.scrollHeight + textareaBorder) + 'px';
    }

    textInput.addEventListener('input', autoGrowTextarea);
    textInput.addEventListener('input', liveSchedule);

    // Banner and color changes are discrete events — re-render immediately,
    // but still only when live mode is on.
    document.querySelectorAll('input[name="banner"]').forEach(function (radio) {
        radio.addEventListener('change', liveRender);
    });
    canvas.addEventListener('click', liveRender);
    hexInput.addEventListener('input', liveRender);

    liveToggle.addEventListener('change', applyMode);

    // The Generate button (form submit) always renders — it is the manual
    // trigger, and harmless in live mode (where the button is disabled).
    // Intercepting the submit also stops the page reload that used to reset
    // the background and wipe the text. Without JS the form posts normally
    // and the server returns a full page, so the feature degrades gracefully.
    previewForm.addEventListener('submit', function (event) {
        event.preventDefault();
        renderPreview();
    });

    // Set the initial button state and paint any text already present
    // (e.g. after a no-JS POST fallback or browser-restored form state).
    applyMode();
    autoGrowTextarea();
})();
