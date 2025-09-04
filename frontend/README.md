# Financial Risk Monitor - Frontend

A modern, responsive web application for portfolio risk management and monitoring built with Next.js 14, TypeScript, and Tailwind CSS.

## Features

- **Modern Authentication**: JWT-based authentication with secure cookie storage
- **Real-time Dashboard**: Live portfolio monitoring with WebSocket updates
- **Risk Analytics**: Advanced VaR calculations and liquidity risk assessment
- **Transaction Management**: Comprehensive transaction history with filtering and export
- **Responsive Design**: Mobile-first design with professional fintech UI
- **Real-time Updates**: Live price feeds and risk metric updates
- **Interactive Charts**: Data visualization with Recharts
- **Compliance Ready**: Built-in KYC/AML status monitoring

## Tech Stack

- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **UI Components**: Headless UI, Heroicons
- **Charts**: Recharts
- **Animations**: Framer Motion
- **Forms**: React Hook Form with Zod validation
- **HTTP Client**: Axios
- **State Management**: React Context API
- **Authentication**: JWT with secure cookies

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Backend API running on port 8080

### Installation

1. Install dependencies:
```bash
npm install
```

2. Create environment file:
```bash
cp .env.local.example .env.local
```

3. Update environment variables:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

4. Run the development server:
```bash
npm run dev
```

5. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
│   ├── dashboard/          # Dashboard page
│   ├── login/             # Login page
│   ├── register/          # Registration page
│   ├── trades/            # Transaction history page
│   └── layout.tsx         # Root layout
├── components/            # Reusable UI components
│   ├── DashboardLayout.tsx
│   ├── StatCard.tsx
│   └── ...
├── hooks/                 # Custom React hooks
│   └── useAuth.ts         # Authentication hook
├── lib/                   # Utility libraries
│   └── api.ts             # API client
└── types/                 # TypeScript type definitions
    └── index.ts
```

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint

## Key Components

### Authentication
- JWT-based authentication with automatic token refresh
- Secure cookie storage with HttpOnly flags
- Protected routes with automatic redirects

### Dashboard
- Real-time portfolio overview
- Live risk metrics with WebSocket updates
- Interactive charts and data visualization
- Recent transactions and active alerts

### Transaction History
- Advanced filtering and sorting
- Export to CSV functionality
- Detailed transaction modal with compliance info
- Real-time status updates

## API Integration

The frontend integrates with the Go backend API:

- **Authentication**: `/api/v1/auth/*`
- **Portfolios**: `/api/v1/portfolios/*`
- **Transactions**: `/api/v1/transactions/*`
- **Risk Metrics**: `/api/v1/risk/*`
- **Alerts**: `/api/v1/alerts/*`
- **WebSocket**: `/ws`

## Styling

The application uses Tailwind CSS with a custom configuration:

- Dark theme optimized for financial applications
- Professional color palette with blue/purple gradients
- Responsive breakpoints for mobile, tablet, and desktop
- Custom animations and transitions
- Backdrop blur effects for modern glass morphism

## Development

### Adding New Pages

1. Create page component in `src/app/[page-name]/page.tsx`
2. Add navigation link in `DashboardLayout.tsx`
3. Update types if needed in `src/types/index.ts`

### Adding New Components

1. Create component in `src/components/`
2. Follow existing patterns for styling and animations
3. Export from component file
4. Import where needed

### API Integration

1. Add API methods to `src/lib/api.ts`
2. Update types in `src/types/index.ts`
3. Use in components with proper error handling

## Deployment

### Production Build

```bash
npm run build
npm run start
```

### Environment Variables

Set the following in production:

```env
NEXT_PUBLIC_API_URL=https://your-api-domain.com/api/v1
NEXT_PUBLIC_WS_URL=wss://your-api-domain.com
NODE_ENV=production
```

## Contributing

1. Follow TypeScript strict mode
2. Use ESLint and Prettier for code formatting
3. Write meaningful commit messages
4. Test on multiple screen sizes
5. Ensure accessibility standards

## License

This project is proprietary software for Financial Risk Monitor.