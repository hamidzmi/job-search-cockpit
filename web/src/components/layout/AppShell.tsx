import React, { useEffect, useState } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { QuickCaptureModal } from '../../features/quick-capture/QuickCaptureModal';

const NAV_ITEMS = [
  { to: '/', label: 'Flight Deck', end: true },
  { to: '/opportunities', label: 'Pipeline', end: false },
  { to: '/network', label: 'Network & Companies', end: false },
  { to: '/profile', label: 'Candidate Profile & Story Bank', end: false },
  { to: '/analytics', label: 'Analytics & Offers', end: false },
  { to: '/settings', label: 'Settings', end: false },
];

export function AppShell() {
  const [quickCaptureOpen, setQuickCaptureOpen] = useState(false);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setQuickCaptureOpen(true);
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  return (
    <div className="cockpit-layout">
      <aside className="cockpit-sidebar">
        <div className="sidebar-brand">
          <span className="brand-badge">JSC</span>
          <div>
            <h1 className="brand-title">Job Search Cockpit</h1>
            <p className="brand-subtitle">Local-First Career Command</p>
          </div>
        </div>

        <button
          type="button"
          className="btn-primary"
          onClick={() => setQuickCaptureOpen(true)}
        >
          + Quick Capture
        </button>

        <nav className="sidebar-nav" aria-label="Workspace Navigation">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `sidebar-link ${isActive ? 'sidebar-link-active' : ''}`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="cockpit-main">
        <Outlet />
      </main>
      {quickCaptureOpen && (
        <QuickCaptureModal onClose={() => setQuickCaptureOpen(false)} />
      )}
    </div>
  );
}
