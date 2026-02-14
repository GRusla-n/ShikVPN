import type { Page } from '../types';

const icons: Record<Page, string> = {
  connection: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
  </svg>`,
  config: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
  </svg>`,
  logs: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
  </svg>`,
};

const labels: Record<Page, string> = {
  connection: 'Connection',
  config: 'Configuration',
  logs: 'Logs',
};

export function renderSidebar(
  container: HTMLElement,
  activePage: Page,
  onNavigate: (page: Page) => void
) {
  container.innerHTML = `
    <div class="sidebar-logo">
      <h1>ShikVPN</h1>
      <span>WireGuard Client</span>
    </div>
    ${(Object.keys(icons) as Page[])
      .map(
        (page) => `
      <div class="nav-item ${page === activePage ? 'active' : ''}" data-page="${page}">
        ${icons[page]}
        <span>${labels[page]}</span>
      </div>
    `
      )
      .join('')}
  `;

  container.querySelectorAll('.nav-item').forEach((el) => {
    el.addEventListener('click', () => {
      const page = (el as HTMLElement).dataset.page as Page;
      onNavigate(page);
    });
  });
}
