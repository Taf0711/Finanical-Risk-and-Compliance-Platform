import type { Metadata, Viewport } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/contexts/AuthContext";

const inter = Inter({ 
  subsets: ["latin"],
  variable: "--font-inter",
  display: 'swap',
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-jetbrains-mono",
  display: 'swap',
});

export const metadata: Metadata = {
  title: "FinRisk Pro - Advanced Financial Risk Analytics Platform",
  description: "Professional-grade financial risk monitoring and compliance management platform with real-time analytics, portfolio optimization, and institutional-level trading capabilities.",
  keywords: [
    "financial risk management",
    "portfolio analytics", 
    "trading platform",
    "risk monitoring",
    "compliance management",
    "monte carlo simulation",
    "var calculation",
    "institutional trading",
    "quantitative finance",
    "real-time analytics"
  ],
  authors: [{ name: "FinRisk Pro Team", url: "https://finrisk.pro" }],
  creator: "FinRisk Pro",
  publisher: "FinRisk Pro",
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://finrisk.pro',
    title: 'FinRisk Pro - Advanced Financial Risk Analytics',
    description: 'Professional-grade financial risk monitoring with real-time analytics and institutional trading capabilities.',
    siteName: 'FinRisk Pro',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'FinRisk Pro - Advanced Financial Risk Analytics',
    description: 'Professional-grade financial risk monitoring with real-time analytics.',
    creator: '@finriskpro',
  },
  icons: {
    icon: [
      { url: '/favicon.ico', sizes: 'any' },
      { url: '/icon.svg', type: 'image/svg+xml' },
    ],
    apple: [
      { url: '/apple-touch-icon.png', sizes: '180x180', type: 'image/png' },
    ],
  },
  manifest: '/manifest.json',
  other: {
    'msapplication-TileColor': '#000000',
    'theme-color': '#000000',
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <head>
        {/* Preconnect to external domains for performance */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="" />
        
        {/* DNS prefetch for better performance */}
        <link rel="dns-prefetch" href="https://api.finrisk.pro" />
        
        {/* Security headers */}
        <meta httpEquiv="X-Content-Type-Options" content="nosniff" />
        <meta httpEquiv="X-Frame-Options" content="DENY" />
        <meta httpEquiv="X-XSS-Protection" content="1; mode=block" />
        <meta httpEquiv="Referrer-Policy" content="strict-origin-when-cross-origin" />
        
        {/* Performance hints */}
        <link rel="preload" href="/fonts/inter-var.woff2" as="font" type="font/woff2" crossOrigin="" />
      </head>
      <body 
        className={`${inter.variable} ${jetbrainsMono.variable} font-sans antialiased`}
        suppressHydrationWarning
      >
        {/* Skip to main content for accessibility */}
        <a 
          href="#main-content" 
          className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 z-50 bg-primary text-white px-4 py-2 rounded-lg focus-ring"
        >
          Skip to main content
        </a>
        
        {/* Main application wrapper */}
        <AuthProvider>
          <div id="root" className="min-h-screen bg-background text-foreground">
            <main id="main-content" className="min-h-screen">
              {children}
            </main>
          </div>
        </AuthProvider>
        
        {/* Loading indicator for page transitions */}
        <div 
          id="loading-indicator" 
          className="fixed top-0 left-0 w-full h-1 bg-gradient-primary transform -translate-x-full transition-transform duration-300 z-50 opacity-0"
          aria-hidden="true"
        />
        
        {/* Toast notifications container */}
        <div 
          id="toast-container" 
          className="fixed top-4 right-4 z-50 space-y-2"
          aria-live="polite"
          aria-label="Notifications"
        />
        
        {/* Modal container */}
        <div id="modal-root" />
        
        {/* Accessibility announcements */}
        <div 
          id="a11y-announcer" 
          className="sr-only" 
          aria-live="polite" 
          aria-atomic="true"
        />
        
        {/* Development tools (only in development) */}
        {process.env.NODE_ENV === 'development' && (
          <div id="dev-tools" className="fixed bottom-4 left-4 z-50">
            {/* Development indicators can go here */}
          </div>
        )}
      </body>
    </html>
  );
}