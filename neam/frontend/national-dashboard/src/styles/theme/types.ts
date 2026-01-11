// Theme Types and Interfaces
export interface ThemeColors {
  // Background colors
  background: {
    primary: string;
    secondary: string;
    tertiary: string;
    elevated: string;
  };
  
  // Text colors
  text: {
    primary: string;
    secondary: string;
    tertiary: string;
    inverse: string;
    muted: string;
  };
  
  // Semantic colors
  semantic: {
    primary: string;
    primaryHover: string;
    primaryActive: string;
    success: string;
    successHover: string;
    successBg: string;
    warning: string;
    warningHover: string;
    warningBg: string;
    danger: string;
    dangerHover: string;
    dangerBg: string;
    info: string;
    infoBg: string;
  };
  
  // Chart colors
  chart: {
    palette: string[];
    gradient: {
      start: string;
      end: string;
    };
  };
  
  // Border and divider
  border: {
    light: string;
    default: string;
    dark: string;
  };
  
  // Shadow
  shadow: {
    sm: string;
    md: string;
    lg: string;
    xl: string;
    glow: string;
  };
  
  // Status indicators
  status: {
    stable: string;
    warning: string;
    critical: string;
    unknown: string;
  };
}

export interface ThemeSpacing {
  xs: string;
  sm: string;
  md: string;
  lg: string;
  xl: string;
  xxl: string;
}

export interface ThemeTypography {
  fontFamily: {
    sans: string;
    mono: string;
  };
  fontSize: {
    xs: string;
    sm: string;
    base: string;
    lg: string;
    xl: string;
    xxl: string;
    xxxl: string;
    display: string;
  };
  fontWeight: {
    normal: number;
    medium: number;
    semibold: number;
    bold: number;
  };
  lineHeight: {
    tight: number;
    normal: number;
    relaxed: number;
  };
}

export interface ThemeBreakpoints {
  sm: string;
  md: string;
  lg: string;
  xl: string;
  xxl: string;
}

export interface ThemeTransition {
  duration: {
    fast: string;
    normal: string;
    slow: string;
  };
  easing: {
    ease: string;
    easeIn: string;
    easeOut: string;
    easeInOut: string;
    sharp: string;
  };
}

export interface ThemeZIndex {
  dropdown: number;
  sticky: number;
  fixed: number;
  modal: number;
  popover: number;
  tooltip: number;
  toast: number;
}

export interface ThemeConfig {
  name: string;
  mode: 'dark' | 'light';
  colors: ThemeColors;
  spacing: ThemeSpacing;
  typography: ThemeTypography;
  breakpoints: ThemeBreakpoints;
  transition: ThemeTransition;
  zIndex: ThemeZIndex;
}

// Default Dark Theme Configuration
export const darkTheme: ThemeConfig = {
  name: 'neam-dark',
  mode: 'dark',
  colors: {
    background: {
      primary: '#0B1120',      // Deep Navy/Black
      secondary: '#1E293B',    // Slate 800
      tertiary: '#334155',     // Slate 700
      elevated: '#0F172A',     // Slightly lighter for elevated surfaces
    },
    text: {
      primary: '#F8FAFC',      // Slate 50
      secondary: '#94A3B8',    // Slate 400
      tertiary: '#64748B',     // Slate 500
      inverse: '#0B1120',      // Dark for light backgrounds
      muted: '#475569',        // Slate 600
    },
    semantic: {
      primary: '#3B82F6',      // Blue 500
      primaryHover: '#2563EB', // Blue 600
      primaryActive: '#1D4ED8',// Blue 700
      success: '#10B981',      // Emerald 500
      successHover: '#059669', // Emerald 600
      successBg: 'rgba(16, 185, 129, 0.1)',
      warning: '#F59E0B',      // Amber 500
      warningHover: '#D97706', // Amber 600
      warningBg: 'rgba(245, 158, 11, 0.1)',
      danger: '#EF4444',       // Red 500
      dangerHover: '#DC2626',  // Red 600
      dangerBg: 'rgba(239, 68, 68, 0.1)',
      info: '#06B6D4',         // Cyan 500
      infoBg: 'rgba(6, 182, 212, 0.1)',
    },
    chart: {
      palette: [
        '#3B82F6', // Blue
        '#8B5CF6', // Purple
        '#EC4899', // Pink
        '#10B981', // Emerald
        '#F59E0B', // Amber
        '#06B6D4', // Cyan
      ],
      gradient: {
        start: '#3B82F6',
        end: '#8B5CF6',
      },
    },
    border: {
      light: '#334155',
      default: '#1E293B',
      dark: '#0F172A',
    },
    shadow: {
      sm: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
      md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
      lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
      xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
      glow: '0 0 20px rgba(59, 130, 246, 0.3)',
    },
    status: {
      stable: '#10B981',
      warning: '#F59E0B',
      critical: '#EF4444',
      unknown: '#64748B',
    },
  },
  spacing: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '1rem',
    lg: '1.5rem',
    xl: '2rem',
    xxl: '3rem',
  },
  typography: {
    fontFamily: {
      sans: '"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      mono: '"JetBrains Mono", "Fira Code", Consolas, monospace',
    },
    fontSize: {
      xs: '0.75rem',
      sm: '0.875rem',
      base: '1rem',
      lg: '1.125rem',
      xl: '1.25rem',
      xxl: '1.5rem',
      xxxl: '1.875rem',
      display: '2.25rem',
    },
    fontWeight: {
      normal: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
    },
    lineHeight: {
      tight: 1.25,
      normal: 1.5,
      relaxed: 1.75,
    },
  },
  breakpoints: {
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
    xxl: '1536px',
  },
  transition: {
    duration: {
      fast: '150ms',
      normal: '250ms',
      slow: '350ms',
    },
    easing: {
      ease: 'ease',
      easeIn: 'ease-in',
      easeOut: 'ease-out',
      easeInOut: 'ease-in-out',
      sharp: 'cubic-bezier(0.4, 0, 0.2, 1)',
    },
  },
  zIndex: {
    dropdown: 1000,
    sticky: 1020,
    fixed: 1030,
    modal: 1050,
    popover: 1060,
    tooltip: 1070,
    toast: 1080,
  },
};

// Light theme (for future expansion)
export const lightTheme: ThemeConfig = {
  ...darkTheme,
  name: 'neam-light',
  mode: 'light',
  colors: {
    ...darkTheme.colors,
    background: {
      primary: '#F8FAFC',
      secondary: '#FFFFFF',
      tertiary: '#F1F5F9',
      elevated: '#FFFFFF',
    },
    text: {
      primary: '#0F172A',
      secondary: '#475569',
      tertiary: '#64748B',
      inverse: '#F8FAFC',
      muted: '#94A3B8',
    },
    border: {
      light: '#E2E8F0',
      default: '#CBD5E1',
      dark: '#94A3B8',
    },
  },
};

export type ThemeMode = 'dark' | 'light';
export type Theme = ThemeConfig;
