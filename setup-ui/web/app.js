(() => {
  const deviceSelect = document.querySelector('#device-uri');
  if (!deviceSelect) return;

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
      if (!response.ok) throw new Error('Suche fehlgeschlagen');
      return response.json();
    })
    .then(({devices}) => {
      deviceSelect.replaceChildren();
      selectWrap.classList.remove('loading');
      deviceSelect.disabled = false;

      if (!devices?.length) {
        const option = new Option('Kein Netzwerkdrucker gefunden', '');
        deviceSelect.add(option);
        status.textContent = 'Öffne unten „Mein Drucker wird nicht gefunden“ und trage seine IP-Adresse ein.';
        manualBox.open = true;
        manualBox.dispatchEvent(new Event('toggle'));
        return;
      }

      deviceSelect.add(new Option('Drucker auswählen …', ''));
      devices.forEach((device) => {
        const option = new Option(`${device.name} · ${device.protocol}`, device.uri);
        option.dataset.name = device.name;
        deviceSelect.add(option);
      });
      deviceSelect.required = true;
      status.textContent = `${devices.length} Netzwerkdrucker gefunden.`;
      if (devices.length === 1) {
        deviceSelect.selectedIndex = 1;
        applyDevice();
      }
    })
    .catch(() => {
      selectWrap.classList.remove('loading');
      deviceSelect.replaceChildren(new Option('Druckersuche nicht verfügbar', ''));
      deviceSelect.disabled = false;
      status.textContent = 'Du kannst den Drucker stattdessen über seine IP-Adresse einrichten.';
      manualBox.open = true;
      manualBox.dispatchEvent(new Event('toggle'));
    });
})();
