import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { AuthProvider } from '@/hooks/useAuth';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Financial Risk Monitor',
  description: 'Advanced portfolio risk management and monitoring platform',
  keywords: 'finance, risk management, portfolio, trading, compliance, VaR, liquidity',
  authors: [{ name: 'Financial Risk Monitor Team' }],
  openGraph: {
    title: 'Financial Risk Monitor',
    description: 'Advanced portfolio risk management and monitoring platform',
    type: 'website',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className={inter.className} suppressHydrationWarning={true}>
        <div className="bg-slate-900 text-white antialiased min-h-screen">
          <AuthProvider>
            {children}
          </AuthProvider>
        </div>
      </body>
    </html>
  );
}