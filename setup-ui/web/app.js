(() => {
  const deviceSelect = document.querySelector('#device-uri');
  if (!deviceSelect) return;

  // Texts for the printer search. The page language comes from <html lang>.
  const texts = {
    en: {
      searchFailed: 'Search failed',
      noPrinter: 'No network printer found',
      openManual: 'Open "My printer isn\'t listed" below and enter its IP address.',
      choose: 'Choose printer …',
      found: (count) => (count === 1 ? '1 network printer found.' : `${count} network printers found.`),
      unavailable: 'Printer discovery not available',
      useIP: 'You can set up the printer by its IP address instead.',
    },
    de: {
      searchFailed: 'Suche fehlgeschlagen',
      noPrinter: 'Kein Netzwerkdrucker gefunden',
      openManual: 'Öffne unten „Mein Drucker wird nicht gefunden“ und trage seine IP-Adresse ein.',
      choose: 'Drucker auswählen …',
      found: (count) => `${count} Netzwerkdrucker gefunden.`,
      unavailable: 'Druckersuche nicht verfügbar',
      useIP: 'Du kannst den Drucker stattdessen über seine IP-Adresse einrichten.',
    },
  };
  const t = texts[document.documentElement.lang] || texts.en;

  const selectWrap = document.querySelector('#device-select-wrap');
  const status = document.querySelector('#discovery-status');
  const manualBox = document.querySelector('#manual-box');
  const connectionMode = document.querySelector('#connection-mode');
  const printerIP = document.querySelector('#printer-ip');
  const modelSelect = document.querySelector('#model-uri');
  const description = document.querySelector('#description');
  const queueName = document.querySelector('#queue-name');

  const queueSafe = (value) => value
    .trim()
    .replace(/[^a-zA-Z0-9_-]+/g, '_')
    .replace(/^[_-]+|[_-]+$/g, '')
    .slice(0, 64);

  const modelKey = (value) => value
    .toLowerCase()
    .replace(/brother/g, '')
    .replace(/[^a-z0-9]+/g, '');

  const chooseMatchingModel = (deviceName) => {
    const deviceKey = modelKey(deviceName);
    let best = null;
    let bestLength = 0;
    [...modelSelect.options].forEach((option) => {
      if (!option.value) return;
      const optionKey = modelKey(option.textContent);
      if (optionKey && (deviceKey.includes(optionKey) || optionKey.includes(deviceKey)) && optionKey.length > bestLength) {
        best = option;
        bestLength = optionKey.length;
      }
    });
    if (best) modelSelect.value = best.value;
  };

  const applyDevice = () => {
    const option = deviceSelect.selectedOptions[0];
    if (!option || !option.value) return;
    const name = option.dataset.name || option.textContent;
    description.value = name;
    queueName.value = queueSafe(name);
    chooseMatchingModel(name);
  };

  manualBox?.addEventListener('toggle', () => {
    const manual = manualBox.open;
    connectionMode.value = manual ? 'manual' : 'discovered';
    printerIP.required = manual;
    deviceSelect.required = !manual;
  });

  deviceSelect.addEventListener('change', applyDevice);

  fetch('/api/devices', {headers: {'Accept': 'application/json'}})
    .then((response) => {
      if (!response.ok) throw new Error(t.searchFailed);
      return response.json();
    })
    .then(({devices}) => {
      deviceSelect.replaceChildren();
      selectWrap.classList.remove('loading');
      deviceSelect.disabled = false;

      if (!devices?.length) {
        const option = new Option(t.noPrinter, '');
        deviceSelect.add(option);
        status.textContent = t.openManual;
        manualBox.open = true;
        manualBox.dispatchEvent(new Event('toggle'));
        return;
      }

      deviceSelect.add(new Option(t.choose, ''));
      devices.forEach((device) => {
        const option = new Option(`${device.name} · ${device.protocol}`, device.uri);
        option.dataset.name = device.name;
        deviceSelect.add(option);
      });
      deviceSelect.required = true;
      status.textContent = t.found(devices.length);
      if (devices.length === 1) {
        deviceSelect.selectedIndex = 1;
        applyDevice();
      }
    })
    .catch(() => {
      selectWrap.classList.remove('loading');
      deviceSelect.replaceChildren(new Option(t.unavailable, ''));
      deviceSelect.disabled = false;
      status.textContent = t.useIP;
      manualBox.open = true;
      manualBox.dispatchEvent(new Event('toggle'));
    });
})();
