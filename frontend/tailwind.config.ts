import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Modern Financial Dashboard Theme
        background: {
          DEFAULT: '#0a0a0a',
          secondary: '#111111',
          tertiary: '#1a1a1a',
        },
        foreground: {
          DEFAULT: '#ffffff',
          muted: '#a1a1aa',
          subtle: '#71717a',
        },
        primary: {
          DEFAULT: '#00d4ff',
          50: '#f0fdff',
          100: '#ccf7fe',
          200: '#99effd',
          300: '#60e1fa',
          400: '#21c9f0',
          500: '#00d4ff',
          600: '#0090c4',
          700: '#0b729e',
          800: '#155e82',
          900: '#164e6d',
          950: '#0a3149',
        },
        accent: {
          DEFAULT: '#7c3aed',
          50: '#f5f3ff',
          100: '#ede9fe',
          200: '#ddd6fe',
          300: '#c4b5fd',
          400: '#a78bfa',
          500: '#7c3aed',
          600: '#7c3aed',
          700: '#6d28d9',
          800: '#5b21b6',
          900: '#4c1d95',
          950: '#2e1065',
        },
        success: {
          DEFAULT: '#00ff88',
          50: '#f0fdf5',
          100: '#dcfce8',
          200: '#bbf7d1',
          300: '#86efac',
          400: '#4ade80',
          500: '#00ff88',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
          950: '#052e16',
        },
        danger: {
          DEFAULT: '#ff3366',
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#ff3366',
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
          950: '#450a0a',
        },
        warning: {
          DEFAULT: '#ffa500',
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#ffa500',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
          950: '#451a03',
        },
        border: {
          DEFAULT: '#27272a',
          muted: '#18181b',
        },
        card: {
          DEFAULT: 'rgba(17, 17, 17, 0.8)',
          hover: 'rgba(26, 26, 26, 0.9)',
        },
        glass: {
          DEFAULT: 'rgba(255, 255, 255, 0.05)',
          strong: 'rgba(255, 255, 255, 0.1)',
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      fontSize: {
        '2xs': '0.625rem',
      },
      spacing: {
        '18': '4.5rem',
        '72': '18rem',
        '84': '21rem',
        '96': '24rem',
      },
      borderRadius: {
        '2xl': '1rem',
        '3xl': '1.5rem',
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
        'gradient-primary': 'linear-gradient(135deg, #00d4ff 0%, #0090c4 100%)',
        'gradient-success': 'linear-gradient(135deg, #00ff88 0%, #16a34a 100%)',
        'gradient-danger': 'linear-gradient(135deg, #ff3366 0%, #dc2626 100%)',
        'gradient-warning': 'linear-gradient(135deg, #ffa500 0%, #d97706 100%)',
        'gradient-accent': 'linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%)',
        'mesh-gradient': `
          radial-gradient(circle at 20% 80%, rgba(0, 212, 255, 0.1) 0%, transparent 50%),
          radial-gradient(circle at 80% 20%, rgba(124, 58, 237, 0.1) 0%, transparent 50%),
          radial-gradient(circle at 40% 40%, rgba(0, 255, 136, 0.05) 0%, transparent 50%)
        `,
      },
      boxShadow: {
        'glow-primary': '0 0 20px rgba(0, 212, 255, 0.3)',
        'glow-success': '0 0 20px rgba(0, 255, 136, 0.3)',
        'glow-danger': '0 0 20px rgba(255, 51, 102, 0.3)',
        'glow-accent': '0 0 20px rgba(124, 58, 237, 0.3)',
        'glass': '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-left': 'slideLeft 0.3s ease-out',
        'slide-right': 'slideRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'bounce-gentle': 'bounceGentle 2s infinite',
        'pulse-glow': 'pulseGlow 2s ease-in-out infinite',
        'gradient-x': 'gradientX 3s ease infinite',
        'float': 'float 6s ease-in-out infinite',
        'spin-slow': 'spin 3s linear infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideLeft: {
          '0%': { transform: 'translateX(10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideRight: {
          '0%': { transform: 'translateX(-10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.9)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        bounceGentle: {
          '0%, 100%': { transform: 'translateY(-5%)' },
          '50%': { transform: 'translateY(0)' },
        },
        pulseGlow: {
          '0%, 100%': { 
            boxShadow: '0 0 20px rgba(0, 212, 255, 0.2)',
            transform: 'scale(1)',
          },
          '50%': { 
            boxShadow: '0 0 30px rgba(0, 212, 255, 0.4)',
            transform: 'scale(1.02)',
          },
        },
        gradientX: {
          '0%, 100%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-10px)' },
        },
      },
      backdropBlur: {
        xs: '2px',
      },
      screens: {
        '3xl': '1600px',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    function(pluginApi: { addUtilities: (utilities: Record<string, any>) => void }) {
      const newUtilities = {
        '.bg-gradient-primary': {
          background: 'linear-gradient(135deg, #00d4ff 0%, #0090c4 100%)',
        },
        '.bg-gradient-success': {
          background: 'linear-gradient(135deg, #00ff88 0%, #16a34a 100%)',
        },
        '.bg-gradient-danger': {
          background: 'linear-gradient(135deg, #ff3366 0%, #dc2626 100%)',
        },
        '.bg-gradient-warning': {
          background: 'linear-gradient(135deg, #ffa500 0%, #d97706 100%)',
        },
        '.bg-gradient-accent': {
          background: 'linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%)',
        },
        '.shadow-glow-primary': {
          boxShadow: '0 0 20px rgba(0, 212, 255, 0.3)',
        },
        '.shadow-glow-success': {
          boxShadow: '0 0 20px rgba(0, 255, 136, 0.3)',
        },
        '.shadow-glow-danger': {
          boxShadow: '0 0 20px rgba(255, 51, 102, 0.3)',
        },
        '.shadow-glow-accent': {
          boxShadow: '0 0 20px rgba(124, 58, 237, 0.3)',
        },
        '.shadow-glass': {
          boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
        },
        '.glass-card': {
          backdropFilter: 'blur(24px)',
          border: '1px solid rgba(255, 255, 255, 0.1)',
          borderRadius: '1rem',
          boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
          background: 'rgba(17, 17, 17, 0.8)',
          transition: 'all 0.3s ease',
        },
        '.badge': {
          display: 'inline-flex',
          alignItems: 'center',
          padding: '0.25rem 0.75rem',
          borderRadius: '9999px',
          fontSize: '0.75rem',
          fontWeight: '600',
        },
        '.bg-primary': {
          backgroundColor: 'rgb(0, 212, 255)',
        },
        '.text-primary': {
          color: 'rgb(0, 212, 255)',
        },
        '.bg-success': {
          backgroundColor: 'rgb(0, 255, 136)',
        },
        '.text-success': {
          color: 'rgb(0, 255, 136)',
        },
        '.bg-danger': {
          backgroundColor: 'rgb(255, 51, 102)',
        },
        '.text-danger': {
          color: 'rgb(255, 51, 102)',
        },
        '.bg-warning': {
          backgroundColor: 'rgb(255, 165, 0)',
        },
        '.text-warning': {
          color: 'rgb(255, 165, 0)',
        },
        '.bg-accent': {
          backgroundColor: 'rgb(124, 58, 237)',
        },
        '.text-accent': {
          color: 'rgb(124, 58, 237)',
        },
        '.bg-background': {
          backgroundColor: 'rgb(10, 10, 10)',
        },
        '.text-foreground': {
          color: 'rgb(255, 255, 255)',
        },
        '.text-foreground-muted': {
          color: 'rgb(161, 161, 170)',
        },
        '.bg-foreground-muted': {
          backgroundColor: 'rgb(161, 161, 170)',
        },
      }
      pluginApi.addUtilities(newUtilities)
    }
  ],
}

export default config