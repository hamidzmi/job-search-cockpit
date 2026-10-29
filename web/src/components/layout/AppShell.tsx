import React, { useEffect, useState } from 'react';
import { NavLink, Outlet, useLocation } from 'react-router-dom';
import {
  BarChart3,
  Compass,
  FileText,
  Kanban,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
  Search,
  Settings,
  Sun,
  Users,
} from 'lucide-react';
import { QuickCaptureModal } from '../../features/quick-capture/QuickCaptureModal';
import { useTheme } from '../../lib/theme';

const NAV_ITEMS = [
  {
    to: '/',
    label: 'Flight Deck',
    end: true,
    icon: Compass,
  },
  {
    to: '/opportunities',
    label: 'Pipeline',
    end: false,
    icon: Kanban,
  },
  {
    to: '/network',
    label: 'Network & Companies',
    end: false,
    icon: Users,
  },
  {
    to: '/profile',
    label: 'Candidate Profile & Story Bank',
    end: false,
    icon: FileText,
  },
  {
    to: '/analytics',
    label: 'Analytics & Offers',
    end: false,
    icon: BarChart3,
  },
  {
    to: '/settings',
    label: 'Settings',
    end: false,
    icon: Settings,
  },
];

function safeGetItem(key: string): string | null {
  try {
    if (typeof window !== 'undefined' && typeof window.localStorage?.getItem === 'function') {
      return window.localStorage.getItem(key);
    }
  } catch {
    // ignore
  }
  return null;
}

function safeSetItem(key: string, value: string): void {
  try {
    if (typeof window !== 'undefined' && typeof window.localStorage?.setItem === 'function') {
      window.localStorage.setItem(key, value);
    }
  } catch {
    // ignore
  }
}

export function AppShell() {
  const [quickCaptureOpen, setQuickCaptureOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(() => {
    return safeGetItem('cockpit-sidebar-collapsed') === 'true';
  });
  const { theme, toggleTheme } = useTheme();
  const location = useLocation();

  useEffect(() => {
    safeSetItem('cockpit-sidebar-collapsed', String(collapsed));
  }, [collapsed]);

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

  const currentNav = NAV_ITEMS.find((item) =>
    item.end ? location.pathname === item.to : location.pathname.startsWith(item.to),
  );

  return (
    <div className={`cockpit-layout ${collapsed ? 'sidebar-collapsed' : ''}`}>
      <aside className="cockpit-sidebar">
        <div className="sidebar-brand">
          <span className="brand-badge">JSC</span>
          {!collapsed && (
            <div className="brand-text">
              <h1 className="brand-title">Job Search Cockpit</h1>
              <p className="brand-subtitle">Career Command Center</p>
            </div>
          )}
        </div>

        <button
          type="button"
          className="sidebar-quick-btn"
          onClick={() => setQuickCaptureOpen(true)}
          aria-label="Quick Capture"
          title="Quick Capture (⌘K)"
        >
          <div className="flex items-center gap-2">
            <Plus className="w-4 h-4 text-sky-400 shrink-0" />
            {!collapsed && <span>+ Quick Capture</span>}
          </div>
          {!collapsed && <kbd className="sidebar-shortcut-badge">⌘K</kbd>}
        </button>

        <nav className="sidebar-nav" aria-label="Workspace Navigation">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                title={collapsed ? item.label : undefined}
                className={({ isActive }) =>
                  `sidebar-link ${isActive ? 'sidebar-link-active' : ''}`
                }
              >
                <Icon className="sidebar-icon shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </NavLink>
            );
          })}
        </nav>
      </aside>

      <div className="cockpit-content-wrap">
        <header className="cockpit-header">
          <div className="header-left">
            <button
              type="button"
              className="header-icon-btn"
              onClick={() => setCollapsed((c) => !c)}
              aria-label="Toggle sidebar collapse"
              title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
              {collapsed ? (
                <PanelLeftOpen className="w-4 h-4" />
              ) : (
                <PanelLeftClose className="w-4 h-4" />
              )}
            </button>
            <div className="header-breadcrumbs">
              <span className="text-muted-foreground font-medium text-xs">Cockpit</span>
              <span className="text-muted-foreground text-xs">/</span>
              <span className="font-semibold text-sm text-foreground">
                {currentNav?.label ?? 'Workspace'}
              </span>
            </div>
          </div>

          <div className="header-right">
            <button
              type="button"
              className="header-search-trigger"
              onClick={() => setQuickCaptureOpen(true)}
              aria-label="Search opportunities (⌘K)"
            >
              <Search className="w-3.5 h-3.5 text-muted-foreground" />
              <span className="search-placeholder">Quick capture or search...</span>
              <kbd className="header-kbd">⌘K</kbd>
            </button>

            <button
              type="button"
              className="header-icon-btn"
              onClick={toggleTheme}
              aria-label="Toggle theme"
              title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            >
              {theme === 'dark' ? (
                <Sun className="w-4 h-4 text-amber-400" />
              ) : (
                <Moon className="w-4 h-4 text-slate-700" />
              )}
            </button>
          </div>
        </header>

        <main className="cockpit-main">
          <Outlet />
        </main>
      </div>

      {quickCaptureOpen && (
        <QuickCaptureModal onClose={() => setQuickCaptureOpen(false)} />
      )}
    </div>
  );
}
