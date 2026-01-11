# CSIC Regulator Dashboard

A comprehensive React-based regulatory compliance dashboard for monitoring VASP/CASP licensing, energy consumption analytics, and generating regulatory reports.

![Dashboard Preview](./docs/preview.png)

## Features

### Dashboard Overview
- Real-time KPI cards showing key metrics
- Interactive D3.js charts for data visualization
- Recent activity feed and pending actions
- Quick access to critical information

### License Management
- Complete VASP/CASP license lifecycle management
- Advanced filtering and search capabilities
- Compliance score tracking
- License detail modals with full history
- Support for multiple license types (exchange, custodian, wallet, mining)

### Energy Analytics
- Real-time energy consumption monitoring
- Load forecasting with predictive analytics
- Regional energy data visualization
- Grid status monitoring and alerts
- Renewable energy mix tracking

### Regulatory Reports
- Multiple report types (compliance, energy, financial, audit, regulatory)
- Various output formats (PDF, CSV, XLSX, JSON)
- Report scheduling and templates
- Generation progress tracking
- WORM storage compliant

### Settings & Configuration
- Profile management
- Notification preferences
- Security settings (2FA, sessions)
- Theme customization (light/dark/system)
- Service integrations configuration
- API key management

## Tech Stack

- **Frontend Framework**: React 18 with TypeScript
- **Build Tool**: Vite
- **State Management**: Context API + Zustand
- **Styling**: Tailwind CSS
- **Routing**: React Router v6
- **Data Visualization**: D3.js + Recharts
- **HTTP Client**: Axios
- **Icons**: Lucide React
- **Date Handling**: date-fns
- **Testing**: Vitest

## Project Structure

```
frontend/dashboard/
├── src/
│   ├── components/
│   │   ├── common/           # Reusable UI components
│   │   │   ├── Button.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Modal.tsx
│   │   │   └── Table.tsx
│   │   ├── layout/           # Layout components
│   │   │   ├── Layout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Header.tsx
│   │   └── visualizations/   # D3 chart components
│   │       ├── LineChart.tsx
│   │       ├── BarChart.tsx
│   │       └── PieChart.tsx
│   ├── pages/                # Page components
│   │   ├── Dashboard.tsx
│   │   ├── Licenses.tsx
│   │   ├── Energy.tsx
│   │   ├── Reports.tsx
│   │   ├── Settings.tsx
│   │   └── Login.tsx
│   ├── services/             # API services
│   │   ├── api.ts
│   │   ├── licensing.ts
│   │   ├── energy.ts
│   │   └── reporting.ts
│   ├── context/              # React contexts
│   │   ├── ThemeContext.tsx
│   │   ├── AuthContext.tsx
│   │   └── NotificationContext.tsx
│   ├── hooks/                # Custom hooks
│   │   ├── useDataFetch.ts
│   │   └── useD3.ts
│   ├── types/                # TypeScript types
│   │   └── index.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
└── README.md
```

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/csic-platform/regulator-dashboard.git
cd regulator-dashboard/frontend/dashboard
```

2. Install dependencies:
```bash
npm install
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Start development server:
```bash
npm run dev
```

5. Open http://localhost:3000 in your browser

### Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Base API URL | `/api` |
| `VITE_LICENSING_SERVICE_URL` | Licensing service URL | `http://localhost:3000` |
| `VITE_ENERGY_SERVICE_URL` | Energy service URL | `http://localhost:8000` |
| `VITE_REPORTING_SERVICE_URL` | Reporting service URL | `http://localhost:3001` |
| `VITE_ENABLE_MOCK_DATA` | Enable mock data for development | `true` |

### Theme Customization

The dashboard supports light, dark, and system themes. Customize colors in `tailwind.config.js`:

```javascript
theme: {
  extend: {
    colors: {
      primary: {
        50: '#eff6ff',
        100: '#dbeafe',
        // ... custom primary colors
      }
    }
  }
}
```

## API Integration

### Service Configuration

Configure backend service URLs in the Settings page or via environment variables:

```typescript
// Licensing Service (Node.js/NestJS)
GET /api/licenses
GET /api/licenses/:id
POST /api/licenses
PATCH /api/licenses/:id
DELETE /api/licenses/:id

// Energy Service (Python/FastAPI)
GET /api/telemetry/current
GET /api/telemetry/history
GET /api/forecast/load
GET /api/energy/regional

// Reporting Service (Node.js/Express)
GET /api/reports
POST /api/reports/generate
GET /api/reports/:id/download
GET /api/reports/templates
```

### Authentication

The dashboard uses JWT-based authentication. Configure the auth endpoint:

```typescript
POST /api/auth/login
POST /api/auth/logout
GET /api/auth/me
```

## D3.js Charts

### LineChart

```tsx
<LineChart
  data={data}
  width={800}
  height={400}
  categories={['Actual', 'Predicted']}
  colors={['#3b82f6', '#10b981']}
  showLegend={true}
  areaFill={true}
/>
```

### BarChart

```tsx
<BarChart
  data={data}
  height={300}
  colors={['#3b82f6']}
  showValues={true}
  horizontal={false}
/>
```

### PieChart

```tsx
<PieChart
  data={data}
  height={300}
  donut={true}
  innerRadius={60}
  showLegend={true}
/>
```

## Testing

```bash
# Run unit tests
npm run test

# Run tests with UI
npm run test:ui

# Run coverage
npm run test:coverage
```

## Linting

```bash
# Check for issues
npm run lint

# Auto-fix issues
npm run lint:fix
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit changes: `git commit -am 'Add my feature'`
4. Push to branch: `git push origin feature/my-feature`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Contact the development team
- Check the documentation in `/docs`

## Acknowledgments

- [React](https://reactjs.org/)
- [Vite](https://vitejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)
- [D3.js](https://d3js.org/)
- [Lucide Icons](https://lucide.dev/)
