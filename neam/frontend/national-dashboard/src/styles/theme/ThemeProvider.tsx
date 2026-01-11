import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { ThemeConfig, ThemeMode } from './types';
import { darkTheme, lightTheme } from './themes';

// Theme Context Interface
interface ThemeContextValue {
  theme: ThemeConfig;
  mode: ThemeMode;
  setMode: (mode: ThemeMode) => void;
  toggleMode: () => void;
  isDark: boolean;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

// Provider Props
interface ThemeProviderProps {
  children: React.ReactNode;
  defaultMode?: ThemeMode;
  storageKey?: string;
}

// CSS Custom Properties Generator
const generateCSSVariables = (theme: ThemeConfig): string => {
  const { colors, spacing, typography, transition, zIndex } = theme;
  
  return `
    :root {
      /* Background Colors */
      --color-bg-primary: ${colors.background.primary};
      --color-bg-secondary: ${colors.background.secondary};
      --color-bg-tertiary: ${colors.background.tertiary};
      --color-bg-elevated: ${colors.background.elevated};
      
      /* Text Colors */
      --color-text-primary: ${colors.text.primary};
      --color-text-secondary: ${colors.text.secondary};
      --color-text-tertiary: ${colors.text.tertiary};
      --color-text-inverse: ${colors.text.inverse};
      --color-text-muted: ${colors.text.muted};
      
      /* Semantic Colors */
      --color-primary: ${colors.semantic.primary};
      --color-primary-hover: ${colors.semantic.primaryHover};
      --color-primary-active: ${colors.semantic.primaryActive};
      
      --color-success: ${colors.semantic.success};
      --color-success-hover: ${colors.semantic.successHover};
      --color-success-bg: ${colors.semantic.successBg};
      
      --color-warning: ${colors.semantic.warning};
      --color-warning-hover: ${colors.semantic.warningHover};
      --color-warning-bg: ${colors.semantic.warningBg};
      
      --color-danger: ${colors.semantic.danger};
      --color-danger-hover: ${colors.semantic.dangerHover};
      --color-danger-bg: ${colors.semantic.dangerBg};
      
      --color-info: ${colors.semantic.info};
      --color-info-bg: ${colors.semantic.infoBg};
      
      /* Status Colors */
      --color-status-stable: ${colors.status.stable};
      --color-status-warning: ${colors.status.warning};
      --color-status-critical: ${colors.status.critical};
      --color-status-unknown: ${colors.status.unknown};
      
      /* Border Colors */
      --border-light: ${colors.border.light};
      --border-default: ${colors.border.default};
      --border-dark: ${colors.border.dark};
      
      /* Shadows */
      --shadow-sm: ${colors.shadow.sm};
      --shadow-md: ${colors.shadow.md};
      --shadow-lg: ${colors.shadow.lg};
      --shadow-xl: ${colors.shadow.xl};
      --shadow-glow: ${colors.shadow.glow};
      
      /* Spacing */
      --spacing-xs: ${spacing.xs};
      --spacing-sm: ${spacing.sm};
      --spacing-md: ${spacing.md};
      --spacing-lg: ${spacing.lg};
      --spacing-xl: ${spacing.xl};
      --spacing-xxl: ${spacing.xxl};
      
      /* Typography */
      --font-sans: ${typography.fontFamily.sans};
      --font-mono: ${typography.fontFamily.mono};
      --font-size-xs: ${typography.fontSize.xs};
      --font-size-sm: ${typography.fontSize.sm};
      --font-size-base: ${typography.fontSize.base};
      --font-size-lg: ${typography.fontSize.lg};
      --font-size-xl: ${typography.fontSize.xl};
      --font-size-xxl: ${typography.fontSize.xxl};
      --font-size-xxxl: ${typography.fontSize.xxxl};
      --font-size-display: ${typography.fontSize.display};
      
      /* Transitions */
      --transition-fast: ${transition.duration.fast};
      --transition-normal: ${transition.duration.normal};
      --transition-slow: ${transition.duration.slow};
      --ease: ${transition.easing.ease};
      --ease-in: ${transition.easing.easeIn};
      --ease-out: ${transition.easing.easeOut};
      --ease-in-out: ${transition.easing.easeInOut};
      --ease-sharp: ${transition.easing.sharp};
      
      /* Z-Index */
      --z-dropdown: ${zIndex.dropdown};
      --z-sticky: ${zIndex.sticky};
      --z-fixed: ${zIndex.fixed};
      --z-modal: ${zIndex.modal};
      --z-popover: ${zIndex.popover};
      --z-tooltip: ${zIndex.tooltip};
      --z-toast: ${zIndex.toast};
    }
  `;
};

/**
 * ThemeProvider Component
 * 
 * Manages theme state and provides CSS custom properties to the application.
 * Supports dark/light mode switching with persistence.
 */
export const ThemeProvider: React.FC<ThemeProviderProps> = ({
  children,
  defaultMode = 'dark',
  storageKey = 'neam-theme-mode',
}) => {
  const [mode, setMode] = useState<ThemeMode>(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem(storageKey);
      if (saved && (saved === 'dark' || saved === 'light')) {
        return saved as ThemeMode;
      }
      // Check system preference
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        return 'dark';
      }
    }
    return defaultMode;
  });

  const theme = mode === 'dark' ? darkTheme : lightTheme;

  // Apply CSS variables to document
  useEffect(() => {
    const styleId = 'neam-theme-variables';
    let styleElement = document.getElementById(styleId) as HTMLStyleElement;
    
    if (!styleElement) {
      styleElement = document.createElement('style');
      styleElement.id = styleId;
      document.head.appendChild(styleElement);
    }
    
    styleElement.textContent = generateCSSVariables(theme);
  }, [theme]);

  // Persist mode changes
  useEffect(() => {
    localStorage.setItem(storageKey, mode);
    
    // Update data-theme attribute for any CSS that needs it
    document.documentElement.setAttribute('data-theme', mode);
    
    // Update body class for global styling
    document.body.classList.remove('theme-dark', 'theme-light');
    document.body.classList.add(`theme-${mode}`);
  }, [mode, storageKey]);

  // Listen for system preference changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e: MediaQueryListEvent) => {
      // Only auto-switch if user hasn't explicitly set a preference
      const saved = localStorage.getItem(storageKey);
      if (!saved) {
        setMode(e.matches ? 'dark' : 'light');
      }
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [storageKey]);

  const toggleMode = useCallback(() => {
    setMode((prev) => (prev === 'dark' ? 'light' : 'dark'));
  }, []);

  const value: ThemeContextValue = {
    theme,
    mode,
    setMode,
    toggleMode,
    isDark: mode === 'dark',
  };

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
};

/**
 * useTheme Hook
 * 
 * Access the current theme configuration and mode management functions.
 */
export const useTheme = (): ThemeContextValue => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};

/**
 * useThemeMode Hook
 * 
 * Simplified hook for just the current mode and toggle function.
 */
export const useThemeMode = (): { mode: ThemeMode; toggleMode: () => void; isDark: boolean } => {
  const { mode, toggleMode, isDark } = useTheme();
  return { mode, toggleMode, isDark };
};

export default ThemeProvider;
