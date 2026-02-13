import type { ClientConfig } from '../types';

const eyeIcon = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>`;
const eyeOffIcon = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>`;

export async function renderConfigEditor(container: HTMLElement) {
  let cfg: ClientConfig;
  try {
    cfg = await (window as any).go.main.App.GetConfig();
  } catch {
    cfg = {
      server: '',
      api_port: 8080,
      server_public_key: '',
      private_key: '',
      address: '',
      dns: '',
      mtu: 1420,
      persistent_keepalive: 25,
      interface_name: 'wg0',
      api_key: '',
      log_level: 'error',
    };
  }

  container.innerHTML = `
    <div class="config-editor">
      <h2>Configuration</h2>

      <div class="form-group">
        <label>Server</label>
        <input type="text" id="cfg-server" value="${esc(cfg.server)}" placeholder="your-server-ip" />
      </div>

      <div class="form-group">
        <label>API Port</label>
        <input type="number" id="cfg-api-port" value="${cfg.api_port}" min="1" max="65535" />
      </div>

      <div class="form-group">
        <label>Private Key</label>
        <div class="input-with-toggle">
          <input type="password" id="cfg-private-key" value="${esc(cfg.private_key)}" placeholder="Base64-encoded private key" />
          <button class="toggle-visibility" data-target="cfg-private-key">${eyeIcon}</button>
        </div>
      </div>

      <div class="form-group">
        <label>API Key</label>
        <div class="input-with-toggle">
          <input type="password" id="cfg-api-key" value="${esc(cfg.api_key)}" placeholder="Optional API key" />
          <button class="toggle-visibility" data-target="cfg-api-key">${eyeIcon}</button>
        </div>
      </div>

      <div class="form-group">
        <label>Server Public Key</label>
        <input type="text" id="cfg-server-public-key" value="${esc(cfg.server_public_key)}" placeholder="Auto-filled after registration" />
      </div>

      <div class="form-group">
        <label>Address</label>
        <input type="text" id="cfg-address" value="${esc(cfg.address)}" placeholder="Auto-assigned (e.g. 10.0.0.2/24)" />
      </div>

      <div class="form-group">
        <label>DNS</label>
        <input type="text" id="cfg-dns" value="${esc(cfg.dns)}" placeholder="1.1.1.1" />
      </div>

      <div class="form-group">
        <label>MTU</label>
        <input type="number" id="cfg-mtu" value="${cfg.mtu}" min="576" max="65535" />
      </div>

      <div class="form-group">
        <label>Persistent Keepalive (seconds)</label>
        <input type="number" id="cfg-keepalive" value="${cfg.persistent_keepalive}" min="0" max="65535" />
      </div>

      <div class="form-group">
        <label>Interface Name</label>
        <input type="text" id="cfg-interface" value="${esc(cfg.interface_name)}" />
      </div>

      <div class="form-group">
        <label>Log Level</label>
        <select id="cfg-log-level">
          <option value="verbose" ${cfg.log_level === 'verbose' ? 'selected' : ''}>Verbose</option>
          <option value="error" ${cfg.log_level === 'error' ? 'selected' : ''}>Error</option>
          <option value="silent" ${cfg.log_level === 'silent' ? 'selected' : ''}>Silent</option>
        </select>
      </div>

      <div class="button-row">
        <button class="btn btn-primary" id="btn-save">Save</button>
        <button class="btn" id="btn-save-as">Save As...</button>
        <button class="btn" id="btn-load">Load File...</button>
      </div>
    </div>
  `;

  // Toggle visibility buttons
  container.querySelectorAll('.toggle-visibility').forEach((btn) => {
    btn.addEventListener('click', () => {
      const targetId = (btn as HTMLElement).dataset.target!;
      const input = document.getElementById(targetId) as HTMLInputElement;
      if (input.type === 'password') {
        input.type = 'text';
        btn.innerHTML = eyeOffIcon;
      } else {
        input.type = 'password';
        btn.innerHTML = eyeIcon;
      }
    });
  });

  // Save
  container.querySelector('#btn-save')!.addEventListener('click', async () => {
    try {
      await (window as any).go.main.App.SaveConfig(readForm());
    } catch (err: any) {
      alert('Save failed: ' + (err?.message || err));
    }
  });

  // Save As
  container.querySelector('#btn-save-as')!.addEventListener('click', async () => {
    try {
      await (window as any).go.main.App.SaveConfigFileAs(readForm());
    } catch (err: any) {
      alert('Save failed: ' + (err?.message || err));
    }
  });

  // Load
  container.querySelector('#btn-load')!.addEventListener('click', async () => {
    try {
      const loaded = await (window as any).go.main.App.LoadConfigFile();
      if (loaded) {
        renderConfigEditor(container); // re-render with loaded values
      }
    } catch (err: any) {
      alert('Load failed: ' + (err?.message || err));
    }
  });
}

function readForm(): ClientConfig {
  const val = (id: string) => (document.getElementById(id) as HTMLInputElement).value;
  const num = (id: string) => parseInt((document.getElementById(id) as HTMLInputElement).value, 10) || 0;

  return {
    server: val('cfg-server'),
    api_port: num('cfg-api-port'),
    private_key: val('cfg-private-key'),
    api_key: val('cfg-api-key'),
    server_public_key: val('cfg-server-public-key'),
    address: val('cfg-address'),
    dns: val('cfg-dns'),
    mtu: num('cfg-mtu'),
    persistent_keepalive: num('cfg-keepalive'),
    interface_name: val('cfg-interface'),
    log_level: val('cfg-log-level'),
  };
}

function esc(s: string): string {
  if (!s) return '';
  return s.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
