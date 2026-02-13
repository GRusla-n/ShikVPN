import type { StatusUpdate } from '../types';

declare function Connect(): Promise<void>;
declare function Disconnect(): Promise<void>;
declare function GetStatus(): Promise<StatusUpdate>;

const powerIcon = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
  <path d="M18.36 6.64a9 9 0 1 1-12.73 0"/>
  <line x1="12" y1="2" x2="12" y2="12"/>
</svg>`;

let currentStatus: StatusUpdate = {
  status: 'disconnected',
  assignedIP: '',
  error: '',
};

export async function renderConnectionPanel(container: HTMLElement) {
  try {
    currentStatus = await (window as any).go.main.App.GetStatus();
  } catch {
    // use defaults
  }
  render(container);
}

function render(container: HTMLElement) {
  const s = currentStatus;
  const isConnected = s.status === 'connected';
  const isConnecting = s.status === 'connecting';

  container.innerHTML = `
    <div class="connection-panel">
      <div class="power-button ${s.status}" id="power-btn" title="${isConnected ? 'Disconnect' : 'Connect'}">
        ${powerIcon}
      </div>

      <div class="status-section">
        <div class="status-row">
          <div class="status-dot ${s.status}"></div>
          <span class="status-text">${s.status}</span>
        </div>
        ${s.assignedIP ? `
        <div class="info-cards">
          <div class="info-card">
            <div class="label">Assigned IP</div>
            <div class="value">${s.assignedIP}</div>
          </div>
        </div>` : ''}
        ${s.error ? `<div class="error-message">${escapeHtml(s.error)}</div>` : ''}
      </div>
    </div>
  `;

  const btn = container.querySelector('#power-btn')!;
  btn.addEventListener('click', async () => {
    if (isConnecting) return;
    try {
      if (isConnected) {
        await (window as any).go.main.App.Disconnect();
      } else {
        await (window as any).go.main.App.Connect();
      }
    } catch (err) {
      console.error('Action failed:', err);
    }
  });
}

export function updateConnectionStatus(status: StatusUpdate, container: HTMLElement) {
  currentStatus = status;
  render(container);
}

function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
