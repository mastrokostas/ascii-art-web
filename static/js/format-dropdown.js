// Custom file-format dropdown — upgrades the native <select class="format-dropdown">
// into a glass, rounded menu that matches the cards. Native select popups can't
// be styled (they are drawn by the OS), so we hide the native control and drive
// a custom listbox instead. The native <select> stays in the DOM so the form
// still submits its value and the control works with JavaScript disabled.
//
// Applies to selects present at load and to the result card built on the fly
// while typing (via a MutationObserver), so this file needs no help from the
// other scripts.
(function () {
    const CARET = '<svg class="format-caret" width="10" height="10" viewBox="0 0 10 10" fill="none" aria-hidden="true">' +
        '<path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>';

    function enhance(select) {
        if (select.dataset.enhanced) return; // idempotent — the observer can fire twice
        select.dataset.enhanced = 'true';

        const options = Array.from(select.options);
        const form = select.closest('form');
        const downloadLabel = form ? form.querySelector('.download-label') : null;

        // Build the custom control: a trigger pill and a listbox menu.
        const wrap = document.createElement('div');
        wrap.className = 'format-select';

        const trigger = document.createElement('button');
        trigger.type = 'button'; // never submits the form
        trigger.className = 'format-trigger';
        trigger.setAttribute('aria-haspopup', 'listbox');
        trigger.setAttribute('aria-expanded', 'false');
        trigger.setAttribute('aria-label', select.getAttribute('aria-label') || 'File format');
        trigger.innerHTML = '<span class="format-value"></span>' + CARET;

        const menu = document.createElement('ul');
        menu.className = 'format-menu';
        menu.setAttribute('role', 'listbox');

        options.forEach(function (opt) {
            const li = document.createElement('li');
            li.className = 'format-option';
            li.setAttribute('role', 'option');
            li.dataset.value = opt.value;
            li.textContent = opt.textContent;
            li.tabIndex = -1;
            menu.appendChild(li);
        });

        wrap.appendChild(trigger);
        wrap.appendChild(menu);
        select.parentNode.insertBefore(wrap, select);
        select.hidden = true; // still submits; just not shown

        const valueLabel = trigger.querySelector('.format-value');

        // Reflect the native select's value into the trigger, the options, and the
        // download button label — the single source of truth stays the <select>.
        function sync() {
            const current = options.find(function (o) { return o.value === select.value; }) || options[0];
            valueLabel.textContent = current.textContent;
            menu.querySelectorAll('.format-option').forEach(function (li) {
                li.setAttribute('aria-selected', li.dataset.value === select.value ? 'true' : 'false');
            });
            if (downloadLabel) {
                downloadLabel.textContent = 'Download .' + select.value;
            }
        }

        function isOpen() { return wrap.classList.contains('open'); }

        function open() {
            wrap.classList.add('open');
            trigger.setAttribute('aria-expanded', 'true');
            const active = menu.querySelector('.format-option[aria-selected="true"]') || menu.firstElementChild;
            if (active) active.focus();
        }

        function close(focusTrigger) {
            wrap.classList.remove('open');
            trigger.setAttribute('aria-expanded', 'false');
            if (focusTrigger) trigger.focus();
        }

        function choose(value) {
            select.value = value;
            sync();
            select.dispatchEvent(new Event('change', { bubbles: true }));
            close(true);
        }

        trigger.addEventListener('click', function () {
            isOpen() ? close(false) : open();
        });

        trigger.addEventListener('keydown', function (e) {
            if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                open();
            }
        });

        menu.addEventListener('click', function (e) {
            const li = e.target.closest('.format-option');
            if (li) choose(li.dataset.value);
        });

        menu.addEventListener('keydown', function (e) {
            const items = Array.from(menu.querySelectorAll('.format-option'));
            const idx = items.indexOf(document.activeElement);
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                items[(idx + 1) % items.length].focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                items[(idx - 1 + items.length) % items.length].focus();
            } else if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                if (document.activeElement.dataset.value) choose(document.activeElement.dataset.value);
            } else if (e.key === 'Escape') {
                e.preventDefault();
                close(true);
            } else if (e.key === 'Tab') {
                close(false);
            }
        });

        // Any click outside the control closes it.
        document.addEventListener('click', function (e) {
            if (isOpen() && !wrap.contains(e.target)) close(false);
        });

        sync();
    }

    function enhanceAll(root) {
        (root || document).querySelectorAll('select.format-dropdown').forEach(enhance);
    }

    // Enhance selects already in the page...
    enhanceAll();

    // ...and any added later (the result card is created on the fly while typing).
    const observer = new MutationObserver(function (mutations) {
        for (const mutation of mutations) {
            for (const node of mutation.addedNodes) {
                if (node.nodeType !== 1) continue;
                if (node.matches && node.matches('select.format-dropdown')) enhance(node);
                if (node.querySelectorAll) enhanceAll(node);
            }
        }
    });
    observer.observe(document.body, { childList: true, subtree: true });
})();
