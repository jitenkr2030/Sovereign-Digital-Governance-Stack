import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { ThemeSettings } from '../types';

interface ThemeContextType {
  theme: ThemeSettings;
  isDark: boolean;
  toggleTheme: () => void;
  setThemeMode: (mode: 'light' | 'dark' | 'system') => void;
  setPrimaryColor: (color: string) => void;
}

const defaultSettings: ThemeSettings = {
  mode: 'system',
  primaryColor: '#3b82f6',
  compactMode: false,
};

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<ThemeSettings>(() => {
    const saved = localStorage.getItem('theme_settings');
    return saved ? { ...defaultSettings, ...JSON.parse(saved) } : defaultSettings;
  });

  const [isDark, setIsDark] = useState(false);

  const updateDarkMode = useCallback((mode: 'light' | 'dark' | 'system') => {
    if (mode === 'system') {
      const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      setIsDark(systemDark);
      document.documentElement.classList.toggle('dark', systemDark);
    } else {
      setIsDark(mode === 'dark');
      document.documentElement.classList.toggle('dark', mode === 'dark');
    }
  }, []);

  useEffect(() => {
    updateDarkMode(theme.mode);

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e: MediaQueryListEvent) => {
      if (theme.mode === 'system') {
        updateDarkMode('system');
      }
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [theme.mode, updateDarkMode]);

  const toggleTheme = () => {
    const newMode = isDark ? 'light' : 'dark';
    const newSettings = { ...theme, mode: newMode };
    setTheme(newSettings);
    localStorage.setItem('theme_settings', JSON.stringify(newSettings));
  };

  const setThemeMode = (mode: 'light' | 'dark' | 'system') => {
    const newSettings = { ...theme, mode };
    setTheme(newSettings);
    localStorage.setItem('theme_settings', JSON.stringify(newSettings));
    updateDarkMode(mode);
  };

  const setPrimaryColor = (color: string) => {
    const newSettings = { ...theme, primaryColor: color };
    setTheme(newSettings);
    localStorage.setItem('theme_settings', JSON.stringify(newSettings));
    document.documentElement.style.setProperty('--color-primary', color);
  };

  return (
    <ThemeContext.Provider
      value={{
        theme,
        isDark,
        toggleTheme,
        setThemeMode,
        setPrimaryColor,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
